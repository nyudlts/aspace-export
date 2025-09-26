package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	export "github.com/nyudlts/aspace-export/aspace_xport"
	"github.com/nyudlts/go-aspace"
)

const appVersion = "v1.1.3-beta"

var (
	config               string
	debug                bool
	environment          string
	exportLoc            string
	formattedTime        string
	format               string
	help                 bool
	reformat             bool
	repository           int
	resource             int
	resourceInfo         []export.ResourceInfo
	startTime            time.Time
	timeout              int
	unpublishedNotes     bool
	unpublishedResources bool
	version              bool
	workDir              string
	workers              int
)

func init() {
	flag.StringVar(&config, "config", "", "location of go-aspace configuration file")
	flag.StringVar(&environment, "environment", "", "environment key of instance to export from")
	flag.IntVar(&repository, "repository", 0, "ID of repository to be exported, leave blank to export all repositories")
	flag.IntVar(&resource, "resource", 0, "ID of a single resource to be exported")
	flag.IntVar(&timeout, "timeout", 20, "client timeout")
	flag.IntVar(&workers, "workers", 8, "number of concurrent workers")
	flag.StringVar(&exportLoc, "export-location", "", "location to export finding aids")
	flag.BoolVar(&help, "help", false, "display the help message")
	flag.BoolVar(&version, "version", false, "display the version of the tool and go-aspace library")
	flag.BoolVar(&reformat, "reformat", false, "tab reformat the output file")
	flag.StringVar(&format, "format", "", "format of export: ead or marc")
	flag.BoolVar(&unpublishedNotes, "include-unpublished-notes", false, "include unpublished notes")
	flag.BoolVar(&unpublishedResources, "include-unpublished-resources", false, "include unpublished resources")
	flag.BoolVar(&debug, "debug", false, "")
}

func printHelp() {
	fmt.Println("usage: aspace-export [options]")
	fmt.Println("options:")
	fmt.Println("  --config           path/to/the go-aspace configuration file					mandatory")
	fmt.Println("  --environment      environment key in config file of the instance to run export against   	mandatory")
	fmt.Println("  --format           the export format either `ead` or `marc`					mandatory")
	fmt.Println("  --export-location  path/to/the location to export finding aids                            	default `.`")
	fmt.Println("  --include-unpublished-notes		include unpublished notes in exports			default `false`")
	fmt.Println("  --include-unpublished-resources	include unpublished resources in exports		default `false`")
	fmt.Println("  --reformat         tab reformat ead xml files							default `false`")
	fmt.Println("  --repository       ID of the repository to be exported, `0` will export all repositories	default `0` ")
	fmt.Println("  --resource         ID of the resource to be exported, `0` will export all resources		default `0` ")
	fmt.Println("  --timeout          client timout in seconds							default `20`")
	fmt.Println("  --workers          number of concurrent export workers to create				default `8`")
	fmt.Println("  --validate         validate exported finding aids against ead2002 schema			default `false`")
	fmt.Println("  --debug	     print debug messages							default `false`")
	fmt.Println("  --version          print the version and version of client version")
}

func main() {

	//parse the flags
	flag.Parse()

	//check for the help message flag `--help`
	if help == true {
		printHelp()
		os.Exit(0)
	}

	//check for the version flag `--version`
	if version == true {
		fmt.Printf("  aspace-export %s <https://github.com/nyudlts/aspace-export>\n", appVersion)
		fmt.Printf("  go-aspace %s <https://github.com/nyudlts/go-aspace>\n", aspace.LibraryVersion)
		fmt.Println()
		os.Exit(0)
	}

	//create timestamp for files
	startTime = time.Now()
	formattedTime = startTime.Format("20060102-050403")

	//starting the application
	export.PrintOnly(fmt.Sprintf("aspace-export %s", appVersion), export.INFO)

	//create logger
	err := export.CreateLogger(debug, fmt.Sprintf("aspace-export-%s.log", formattedTime))
	if err != nil {
		export.PrintAndLog(err.Error(), export.ERROR)
		printHelp()
		os.Exit(1)
	}
	export.LogOnly(fmt.Sprintf("aspace-export %s", appVersion), export.INFO)

	//check critical flags
	err = export.CheckFlags(config, environment, format, resource, repository)
	if err != nil {
		export.PrintAndLog(err.Error(), export.FATAL)
		err = export.CloseLogger()
		if err != nil {
			export.PrintAndLog(err.Error(), export.ERROR)
		}
		printHelp()
		os.Exit(2)
	}

	export.PrintAndLog("all mandatory options set", export.INFO)

	//get the absolute path of the export location
	if exportLoc == "" {
		workDir = fmt.Sprintf("aspace-exports-%s", formattedTime)
		if err = export.CreateWorkDirectory(workDir); err != nil {
			export.PrintAndLog(err.Error(), export.FATAL)
			err = export.CloseLogger()
			if err != nil {
				export.PrintAndLog(err.Error(), export.ERROR)
			}
			os.Exit(7)
		}
		export.PrintAndLog(fmt.Sprintf("working directory created at %s", workDir), export.INFO)
	} else {
		workDir = exportLoc
	}

	workDir, err = filepath.Abs(workDir)
	if err != nil {
		export.PrintAndLog(err.Error(), export.ERROR)
		os.Exit(3)
	}

	//check that export location exists
	err = export.CheckPath(workDir)
	if err != nil {
		export.PrintAndLog(err.Error(), export.FATAL)
		err = export.CloseLogger()
		if err != nil {
			export.PrintAndLog(err.Error(), export.ERROR)
		}
		os.Exit(3)
	}

	abs, _ := filepath.Abs(workDir)
	export.PrintAndLog(fmt.Sprintf("%s exists and is a directory", abs), export.INFO)

	//get a go-aspace api client
	err = export.CreateAspaceClient(config, environment, timeout)
	if err != nil {
		export.PrintAndLog(fmt.Sprintf("failed to create a go-aspace client %s", err.Error()), export.FATAL)
		err = export.CloseLogger()
		if err != nil {
			export.PrintAndLog(err.Error(), export.ERROR)
		}
		os.Exit(4)
	} else {
		export.PrintAndLog(fmt.Sprintf("go-aspace client created, using go-aspace %s", aspace.LibraryVersion), export.INFO)
	}

	//get a map of repositories to be exported
	repositoryMap, err := export.GetRepositoryMap(repository, environment)
	if err != nil {
		export.PrintAndLog(err.Error(), export.FATAL)
		err = export.CloseLogger()
		if err != nil {
			export.PrintAndLog(err.Error(), export.ERROR)
		}
		os.Exit(5)
	}
	export.PrintAndLog(fmt.Sprintf("%d repositories returned from ArchivesSpace", len(repositoryMap)), export.INFO)

	//get a slice of resourceInfo
	resourceInfo, err = export.GetResourceIDs(repositoryMap, resource)
	if err != nil {
		export.PrintAndLog(err.Error(), export.FATAL)
		err = export.CloseLogger()
		if err != nil {
			export.PrintAndLog(err.Error(), export.ERROR)
		}
		os.Exit(6)
	}
	export.PrintAndLog(fmt.Sprintf("%d resources returned from ArchivesSpace", len(resourceInfo)), export.INFO)

	//Create the repository export and failure directories
	err = export.CreateExportDirectories(workDir, repositoryMap, unpublishedResources)
	if err != nil {
		export.PrintAndLog(err.Error(), export.FATAL)
		err = export.CloseLogger()
		if err != nil {
			export.PrintAndLog(err.Error(), export.ERROR)
		}
		os.Exit(8)
	}

	//Validate the export format
	xportFormat, err := export.GetExportFormat(format)
	if err != nil {
		export.PrintAndLog(err.Error(), export.FATAL)
		err = export.CloseLogger()
		if err != nil {
			export.PrintAndLog(err.Error(), export.ERROR)
		}
		os.Exit(9)
	}

	//create ExportOptions struct
	xportOptions := export.ExportOptions{
		WorkDir:              workDir,
		Format:               xportFormat,
		UnpublishedNotes:     unpublishedNotes,
		UnpublishedResources: unpublishedResources,
		Workers:              workers,
		Reformat:             reformat,
		Timestamp:            formattedTime,
	}

	//export resources
	export.PrintAndLog(fmt.Sprintf("processing %d resources", len(resourceInfo)), export.INFO)
	err = export.ExportResources(xportOptions, startTime, formattedTime, &resourceInfo)
	if err != nil {
		export.PrintAndLog(err.Error(), export.FATAL)
		err = export.CloseLogger()
		if err != nil {
			export.PrintAndLog(err.Error(), export.ERROR)
		}
		os.Exit(10)
	}

	export.PrintAndLog("closing logger", export.INFO)
	if err := export.CloseLogger(); err != nil {
		export.LogOnly(fmt.Sprintf("failed to close logger: %s", err.Error()), export.WARNING)
	}

	if err := export.DeleteEmptyDirectories(workDir); err != nil {
		export.LogOnly(fmt.Sprintf("failed to delete empty directories: %s", err.Error()), export.WARNING)
	}

	if err := export.MoveLogfile(workDir); err != nil {
		export.PrintOnly(fmt.Sprintf("failed to move log file: %s", err.Error()), export.ERROR)
	} else {
		export.PrintOnly("moved log to work directory", export.INFO)
	}

	export.PrintOnly("aspace export complete", export.INFO)

	//print the report
	if err := export.PrintReport(); err != nil {
		export.PrintOnly(fmt.Sprintf("failed to print report file: %s", err.Error()), export.WARNING)
	}

	os.Exit(0)
}
