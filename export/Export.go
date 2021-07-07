package export

import (
	"github.com/nyudlts/go-aspace"
	"log"
	"os"
	"path/filepath"
)

var (
	repositoryID  int
	slug          string
	repositoryDir string
	exportDir     string
	failureDir    string
	c             *aspace.ASClient
)

func ExportRepository(s string, rId int, workDir string, client *aspace.ASClient) error {
	c = client
	slug = s
	repositoryID = rId
	repositoryDir = filepath.Join(workDir, slug)
	exportDir = filepath.Join(repositoryDir, "exports")
	failureDir = filepath.Join(repositoryDir, "failures")

	//create the repository directory
	if _, err := os.Stat(repositoryDir); os.IsNotExist(err) {
		innerErr := os.Mkdir(repositoryDir, 0777)
		if innerErr != nil {
			log.Fatalf("FATAL\tcould not create a repository directory at %s", repositoryDir)
		} else {
			log.Println("INFO\tcreated repository directory", repositoryDir)
		}
	} else {
		log.Println("INFO\trepository directory exists, skipping creation of", repositoryDir)
	}

	//create the repository export directory
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		innerErr := os.Mkdir(exportDir, 0777)
		if innerErr != nil {
			log.Fatalf("FATAL\tcould not create an exports directory at %s", exportDir)
		} else {
			log.Println("INFO\tcreated exports directory", exportDir)
		}
	} else {
		log.Println("INFO\texports directory exists, skipping creation of", exportDir)
	}

	//create the repository failure directory
	if _, err := os.Stat(failureDir); os.IsNotExist(err) {
		innerErr := os.Mkdir(failureDir, 0777)
		if innerErr != nil {
			log.Fatalf("FATAL\tcould not create a failure directory at %s", failureDir)
		} else {
			log.Println("INFO\tcreated repository directory", failureDir)
		}
	} else {
		log.Println("INFO\tfailures directory exists, skipping creation of", failureDir)
	}

	err := processRepository()
	if err != nil {
		return err
	}
	return nil

}

func processRepository() error {
	resourceIDs, err := c.GetResourceIDs(repositoryID)
	if err != nil {
		return err
	}
	log.Printf("INFO\tfound %d resources in %s repository", len(resourceIDs), slug)
	if len(resourceIDs) > 1 {
		//split resource slice into chunks
		repoChunks := chunk(resourceIDs, len(resourceIDs)/2)
		log.Printf("INFO\tSplit repositry slice into %d sub-slices", len(repoChunks))
		uriChannel := make(chan []string)
		for i, chunk := range repoChunks {
			go ExportFindingAidChunk(chunk, uriChannel, i+1)
		}
		uris := []string{}
		for range repoChunks {
			chunk := <-uriChannel
			log.Println("INFO\tAdding", len(chunk), "uris to uri list")
			uris = append(uris, chunk...)
		}
		if len(uris) == len(resourceIDs) {
			log.Printf("INFO\t%d of %d resources exported")
		} else {
			log.Printf("ERROR\t%d of %d resources exported")
		}
	}
	return nil
}

func ExportFindingAidChunk(resourceIds []int, uriChannel chan []string, workerID int) {
	log.Println("INFO\tStarting worker", workerID, "processing", len(resourceIds), "resources")

	uris := []string{}
	for _, resourceID := range resourceIds {
		resource, err := c.GetResource(repositoryID, resourceID)
		log.Printf("INFO\tworker %d requesting resource id %d from %s", workerID, resourceID, slug)
		if err != nil {
		}
		uris = append(uris, resource.URI)
	}

	log.Println("INFO\tworker", workerID, "done")
	uriChannel <- uris
}

func chunk(xs []int, chunkSize int) [][]int {
	if len(xs) == 0 {
		return nil
	}
	divided := make([][]int, (len(xs)+chunkSize-1)/chunkSize)
	prev := 0
	i := 0
	till := len(xs) - chunkSize
	for prev < till {
		next := prev + chunkSize
		divided[i] = xs[prev:next]
		prev = next
		i++
	}
	divided[i] = xs[prev:]
	return divided
}
