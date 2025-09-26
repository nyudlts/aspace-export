package aspace_xport

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nyudlts/go-aspace"
)

var (
	numSkipped    = 0
	reportFile    string
	results       []ExportResult
	startTime     time.Time
	executionTime time.Duration
	formattedTime string
	exportOptions ExportOptions
	resourceInfo  *[]ResourceInfo
)

const (
	EAD ExportFormat = iota
	MARC
	UNSUPPORTED
)

type ExportOptions struct {
	WorkDir              string
	Format               ExportFormat
	UnpublishedNotes     bool
	UnpublishedResources bool
	Workers              int
	Reformat             bool
	Timestamp            string
}

type ExportFormat int

func GetExportFormat(xportFormat string) (ExportFormat, error) {
	switch xportFormat {
	case "ead":
		return EAD, nil
	case "marc":
		return MARC, nil
	default:
		return UNSUPPORTED, fmt.Errorf("unsupported format error, %s, supported formats are `ead` or `marc`")
	}
}

type ExportResult struct {
	Status string
	URI    string
	Error  string
}

func ExportResources(options ExportOptions, stTime time.Time, fTime string, resInfo *[]ResourceInfo) error {
	exportOptions = options
	startTime = stTime
	formattedTime = fTime
	resourceInfo = resInfo
	resourceChunks := chunkResources()
	resultChannel := make(chan []ExportResult)

	for i, chunk := range resourceChunks {
		go exportChunk(chunk, resultChannel, i+1)
	}

	for range resourceChunks {
		chunk := <-resultChannel
		results = append(results, chunk...)
	}

	if err := CreateReport(); err != nil {
		return fmt.Errorf("Could not create results report")
	}

	return nil
}

func chunkResources() [][]ResourceInfo {
	var divided [][]ResourceInfo
	ri := *resourceInfo
	chunkSize := (len(ri) + exportOptions.Workers - 1) / exportOptions.Workers

	for i := 0; i < len(*resourceInfo); i += chunkSize {
		end := i + chunkSize

		if end > len(*resourceInfo) {
			end = len(*resourceInfo)
		}

		divided = append(divided, ri[i:end])
	}
	return divided
}

func exportChunk(resourceInfoChunk []ResourceInfo, resultChannel chan []ExportResult, workerID int) {
	PrintAndLog(fmt.Sprintf("starting [worker %d] processing %d resources", workerID, len(resourceInfoChunk)), INFO)
	var results = []ExportResult{}

	//loop through the chunk
	for i, rInfo := range resourceInfoChunk {

		if i > 1 && (i-1)%50 == 0 {
			PrintOnly(fmt.Sprintf("[worker %d] has completed %d exports", workerID, i-1), INFO)
		}
		//get the resource object
		var res *aspace.Resource
		res, err := client.GetResource(rInfo.RepoID, rInfo.ResourceID)
		if err != nil {
			PrintAndLog(fmt.Sprintf("[worker %d] could not retrieve /repositories/%d/resources/%d retrying, code: %s", workerID, rInfo.RepoID, rInfo.ResourceID, err.Error()), WARNING)
			var err2 error
			res, err2 = client.GetResource(rInfo.RepoID, rInfo.ResourceID)
			if err2 != nil {
				PrintAndLog(fmt.Sprintf("[worker %d] could not retrieve resource /repositories/%d/resources/%d/ on 2nd attempt, code: %s", workerID, rInfo.RepoID, rInfo.ResourceID, err2), ERROR)
				results = append(results, ExportResult{Status: "ERROR", URI: fmt.Sprintf("repositories/%d/resources/%d", rInfo.RepoID, rInfo.ResourceID), Error: err.Error()})
				continue
			}
		}

		//check if the resource is set to be published
		if exportOptions.UnpublishedResources == false && res.Publish != true {
			LogOnly(fmt.Sprintf("[worker %d]  resource %s not set to publish, skipping", workerID, res.URI), INFO)
			numSkipped = numSkipped + 1
			results = append(results, ExportResult{Status: "SKIPPED", URI: res.URI, Error: ""})
			continue
		}

		switch exportOptions.Format {
		case MARC:
			results = append(results, exportMarc(rInfo, *res, workerID))
		case EAD:
			results = append(results, exportEAD(rInfo, *res, workerID))
		default:
			//there's an unsupported format, this shouldn't be possible
		}
	}

	PrintAndLog(fmt.Sprintf("[worker %d] finished, processed %d resources", workerID, len(results)), INFO)
	resultChannel <- results
}

func exportMarc(info ResourceInfo, res aspace.Resource, workerID int) ExportResult {
	var marcBytes []byte
	var err error
	//get the marc record
	marcBytes, err = client.GetMARCAsByteArray(info.RepoID, info.ResourceID, exportOptions.UnpublishedNotes)
	if err != nil {
		PrintAndLog(fmt.Sprintf("[worker %d] could not retrieve %s as marc xml, code: %s, retrying", workerID, res.URI, err.Error()), WARNING)
		marcBytes, err = client.GetMARCAsByteArray(info.RepoID, info.ResourceID, exportOptions.UnpublishedNotes)
		if err != nil {
			PrintAndLog(fmt.Sprintf("[worker %d] could not retrieve %s as marc xml on 2nd attempt, code: %s", workerID, res.URI, err.Error()), ERROR)
			return ExportResult{Status: "ERROR", URI: res.URI, Error: err.Error()}
		}
	}

	//create the output filename
	marcFilename := strings.ToLower(fmt.Sprintf("%s_%s.xml", res.EADID, formattedTime))

	//set the location to write the marc record
	var marcPath string
	if exportOptions.UnpublishedResources == true && res.Publish == false {
		marcPath = filepath.Join(exportOptions.WorkDir, info.RepoSlug, "unpublished", marcFilename)
	} else {
		marcPath = filepath.Join(exportOptions.WorkDir, info.RepoSlug, "exports", marcFilename)
	}

	//validate the output
	warning := false
	var warningType = ""

	//write the marc file
	err = os.WriteFile(marcPath, marcBytes, 0777)
	if err != nil {
		LogOnly(fmt.Sprintf("[worker %d]  could not write the marc record %s", workerID, res.URI), ERROR)
		return ExportResult{Status: "ERROR", URI: "", Error: err.Error()}
	}

	//return the result
	if warning == true {
		LogOnly(fmt.Sprintf("[worker %d]  exported resource %s - %s with warning", workerID, res.URI, marcFilename), WARNING)
		return ExportResult{Status: "WARNING", URI: res.URI, Error: warningType}
	}
	LogOnly(fmt.Sprintf("[worker %d] exported resource %s - %s", workerID, res.URI, res.EADID), INFO)
	return ExportResult{Status: "SUCCESS", URI: res.URI, Error: ""}
}

func exportEAD(info ResourceInfo, res aspace.Resource, workerID int) ExportResult {

	//get the ead as bytes
	eadBytes, err := client.GetEADAsByteArray(info.RepoID, info.ResourceID, exportOptions.UnpublishedNotes)
	if err != nil {
		LogOnly(fmt.Sprintf("INFO [worker %d] could not retrieve resource %s", workerID, res.URI), ERROR)
		return ExportResult{Status: "ERROR", URI: res.URI, Error: err.Error()}
	}

	//create the output filename
	eadFilename := fmt.Sprintf("%s.xml", res.EADID)
	outputFile := filepath.Join(exportOptions.WorkDir, info.RepoSlug, "exports", eadFilename)

	//validate the output
	warning := false
	var warningType = ""

	//create the output file
	err = os.WriteFile(outputFile, eadBytes, 0777)
	if err != nil {
		LogOnly(fmt.Sprintf("[worker %d] could not write the ead file %s", workerID, res.URI), ERROR)
		return ExportResult{Status: "ERROR", URI: "", Error: err.Error()}
	}

	//reformat the ead with tabs
	if exportOptions.Reformat == true {
		err = tabReformatXML(outputFile)
		if err != nil {
			LogOnly(fmt.Sprintf("[worker %d] could not reformat %s", workerID, outputFile), WARNING)
		}
	}

	//return the result

	if warning == true {
		LogOnly(fmt.Sprintf("[worker %d] exported resource %s - %s with warning", workerID, res.URI, eadFilename), WARNING)
		return ExportResult{Status: "WARNING", URI: res.URI, Error: warningType}
	}
	LogOnly(fmt.Sprintf("[worker %d] exported resource %s - %s", workerID, res.URI, res.EADID), INFO)
	return ExportResult{Status: "SUCCESS", URI: res.URI, Error: ""}
}

func tabReformatXML(path string) error {

	//lint the ead file
	reformattedBytes, err := exec.Command("xmllint", "--format", path).Output()
	if err != nil {
		return fmt.Errorf("could not reformat %s", path)
	}

	//delete the original
	err = os.Remove(path)
	if err != nil {
		return fmt.Errorf("could not delete %s", path)
	}

	//rewrite the file
	err = os.WriteFile(path, reformattedBytes, 0644)
	if err != nil {
		return fmt.Errorf("could not write reformated bytes to %s", path)
	}

	return nil
}

func MergeIDs(r aspace.Resource) string {
	ids := r.ID0
	for _, i := range []string{r.ID1, r.ID2, r.ID3} {
		if i != "" {
			ids = ids + "_" + i
		}
	}
	return ids
}

func CreateReport() error {
	//seperate result types
	successes := []ExportResult{}
	errors := []ExportResult{}
	warnings := []ExportResult{}
	skipped := []ExportResult{}

	for _, result := range results {
		switch result.Status {
		case "SUCCESS":
			successes = append(successes, result)
		case "ERROR":
			errors = append(errors, result)
		case "WARNING":
			warnings = append(warnings, result)
		case "SKIPPED":
			skipped = append(skipped, result)
		default:
		}
	}

	executionTime = time.Since(startTime)

	reportFile = filepath.Join(exportOptions.WorkDir, fmt.Sprintf("aspace-export-report-%s.txt", exportOptions.Timestamp))
	report, err := os.Create(reportFile)
	if err != nil {
		return err
	}

	defer report.Close()
	writer := bufio.NewWriter(report)
	msg := "ASPACE-EXPORT REPORT\n====================\n"
	msg = msg + fmt.Sprintf("Execution Time: %v", executionTime)
	msg = msg + fmt.Sprintf("\n%d Resources processed:\n", len(results))
	msg = msg + fmt.Sprintf("  %d Successful exports\n", len(successes))
	msg = msg + fmt.Sprintf("  %d Skipped resources\n", len(skipped))
	msg = msg + fmt.Sprintf("  %d Exports with warnings\n", len(warnings))

	if len(warnings) > 0 {
		for _, w := range warnings {
			w.Error = strings.ReplaceAll(w.Error, "\n", " ")
			msg = msg + fmt.Sprintf("    %v\n", w)
		}
	}

	msg = msg + fmt.Sprintf("  %d Errors Encountered\n", len(errors))
	if len(errors) > 0 {
		for _, e := range errors {
			e.Error = strings.ReplaceAll(e.Error, "\n", " ")
			msg = msg + fmt.Sprintf("    %v\n", e)
		}
	}

	_, err = writer.WriteString(msg)
	if err != nil {
		return err
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}
