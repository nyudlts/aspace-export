package aspace_xport

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nyudlts/go-aspace"
)

type ResourceInfo struct {
	RepoID     int
	RepoSlug   string
	ResourceID int
}

var client *aspace.ASClient

func CreateAspaceClient(config string, environment string, timeout int) error {
	var err error
	client, err = aspace.NewClient(config, environment)
	if err != nil {
		return err
	}
	return nil
}

// check the application flags
func CheckFlags(config string, environment string, format string, resource int, repository int) error {
	//check if the config file is set
	if config == "" {
		return fmt.Errorf("location of go-aspace config file is mandatory, set the --config option when running aspace-export")
	}

	//check that the config exists
	if _, err := os.Stat(config); os.IsNotExist(err) {
		return fmt.Errorf("go-aspace config file does not exist at %s", config)
	}
	//check that the environment is set
	if environment == "" {
		return fmt.Errorf("environment to run export against is mandatory, set the --env option when running aspace=export")
	}

	//check that the format is either `ead` or `marc`
	if format != "marc" && format != "ead" {
		return fmt.Errorf("format must be either `ead` or `marc`, set the --format option when running aspace-export")
	}

	//check that a repository id is set if a resource id is set
	if resource != 0 && repository == 0 {
		return fmt.Errorf("a single resource can not be exported if the repository is not specified, set the --repository option when running aspace-export")
	}

	return nil
}

// check that a path exists and is a directory
func CheckPath(path string) error {
	fi, err := os.Stat(path)
	if err == nil {
		if !fi.IsDir() {
			return fmt.Errorf("path %s is not a directory", path)
		}
	} else {
		return err
	}
	return nil
}

// get a map of repository slugs and an id --TO DO reverse map order -- index by ID
func GetRepositoryMap(repository int, environment string) (map[string]int, error) {
	repositories := make(map[string]int)

	if repository != 0 {
		repositoryObject, err := client.GetRepository(repository)
		if err != nil {
			return repositories, err
		}
		repositories[repositoryObject.Slug] = repository
	} else {
		//export all repositories
		repositoryIds, err := client.GetRepositories()
		if err != nil {
			return repositories, err
		}

		for _, r := range repositoryIds {
			repositoryObject, err := client.GetRepository(r)
			if err != nil {
				return repositories, err
			}
			repositories[repositoryObject.Slug] = r
		}
	}
	return repositories, nil
}

// get a slice of ResourceInfo objects for a repository
func GetResourceIDs(repMap map[string]int, resource int) ([]ResourceInfo, error) {

	resources := []ResourceInfo{}

	for repositorySlug, repositoryID := range repMap {
		if resource != 0 {
			resources = append(resources, ResourceInfo{
				RepoID:     repositoryID,
				RepoSlug:   repositorySlug,
				ResourceID: resource,
			})
			continue
		}

		resourceIDs, err := client.GetResourceIDs(repositoryID)
		if err != nil {
			return resources, err
		}

		for _, resourceID := range resourceIDs {
			resources = append(resources, ResourceInfo{
				RepoID:     repositoryID,
				RepoSlug:   repositorySlug,
				ResourceID: resourceID,
			})
		}
	}

	return resources, nil
}

// check that a work directory does not exist if so create it
func CreateWorkDirectory(workDirPath string) error {
	//determine if the directory already exists or if there is an error, if so return an error
	if _, err := os.Stat(workDirPath); err == nil {
		return fmt.Errorf("work directory %s already exists")
	} else if errors.Is(err, os.ErrNotExist) {
		//the workDir doesn't exist -- create it if there are no other errors
	} else {
		return err
	}

	//create the work directory
	err := os.Mkdir(workDirPath, 0755)
	if err != nil {
		return err
	}

	return nil
}

// create the repository, export, failure and unpublished sub directories in the work directory
func CreateExportDirectories(workDirPath string, repositoryMap map[string]int, unpublishedResources bool) error {
	for slug := range repositoryMap {

		repositoryDir := filepath.Join(workDirPath, slug)
		exportDir := filepath.Join(repositoryDir, "exports")
		unpublishedDir := filepath.Join(repositoryDir, "unpublished")

		if _, err := os.Stat(repositoryDir); err != nil {
			if err := os.Mkdir(repositoryDir, 0755); err != nil {
				return err
			}
			PrintAndLog(fmt.Sprintf("created repository directory %s", repositoryDir), INFO)
		} else {
			PrintAndLog(fmt.Sprintf("repository directory %s already exists, skipping", repositoryDir), INFO)
		}

		//create the repository export directory
		if _, err := os.Stat(exportDir); err != nil {
			if err := os.Mkdir(exportDir, 0755); err != nil {
				return err
			}
			PrintAndLog(fmt.Sprintf("created export directory %s", exportDir), INFO)
		} else {
			PrintAndLog(fmt.Sprintf("exports directory %s already exists, skipping", exportDir), INFO)
		}

		//create the unpublished directory if needed
		if unpublishedResources == true {
			if _, err := os.Stat(unpublishedDir); err != nil {
				if err := os.Mkdir(unpublishedDir, 0755); err != nil {
					return err
				}
				PrintAndLog(fmt.Sprintf("created unpublished directory %s", unpublishedDir), INFO)
			} else {
				PrintAndLog(fmt.Sprintf("unpublished directory %s already exists, skipping", unpublishedDir), INFO)
			}
		}
	}

	return nil
}

func MoveLogfile(workDir string) error {
	newLogLoc := filepath.Join(workDir, Logfile)
	if err := os.Rename(Logfile, newLogLoc); err != nil {
		return fmt.Errorf("could not move log file: %s", err.Error())
	}
	PrintOnly(fmt.Sprintf("moved log file to %s", newLogLoc), INFO)
	return nil
}

func DeleteEmptyDirectories(workDir string) error {
	emptyDirectories := []string{}
	if err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			children, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			if len(children) == 0 {
				emptyDirectories = append(emptyDirectories, path)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if len(emptyDirectories) > 0 {
		for _, dir := range emptyDirectories {
			if err := os.Remove(dir); err != nil {
				PrintOnly(fmt.Sprintf("failed to remove empty directory %s", dir), WARNING)
			} else {
				PrintOnly(fmt.Sprintf("removed empty directory %s", dir), INFO)
			}
		}
	}
	return nil
}

func PrintReport() error {
	report, err := os.ReadFile(reportFile)
	if err != nil {
		return err
	}
	fmt.Println("\n" + string(report))
	return nil
}
