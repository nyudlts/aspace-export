aspace-export, v1.1.4
=====================
Command-line utility for bulk export, validation and reformatting of EAD finding aids and MARC records from Archivespace.

Install From Binary
-------------------
1. Download the latest binary for Windows or linux https://github.com/nyudlts/aspace-export/releases/tag/v1.1.2
2. Download the go-aspace.yml template from, [https://github.com/nyudlts/aspace-export/blob/main/go-aspace.yml](go-aspace.yml), and fill in your archivesspace credentials.

Install With Go
---------------
If you have go on your system you can:
1. Install using go
<code>$ go install github.com/nyudlts/aspace-export</code>

2. Download the go-aspace.yml template from, [https://github.com/nyudlts/aspace-export/blob/main/go-aspace.yml](go-aspace.yml), and fill in your archivesspace credentials.

Usage Examples
--------------
1. **export all repositories and resources to ead.xml**<br>
<code>$ aspace-export --config go-aspace.yml --environment local --format ead</code>

2. **export all resources from repository 2 as marc xml**<br>
<code>$ aspace-export --config go-aspace.yml --environment local --format marc--repository 2</code>

3. **export a single resource to a specific directory**<br>
<code>$ aspace-export --config go-aspace.yml --environment local --format marc --repository 2 --resource 10 --export-location /home/aspace/exports</code>

Notes
-----
* If the `export-location` is not set the program will create a directory hierarchy at the in the current working directory named: `aspace-export-[timestamp]. A subdirectory will be created for each repository that was exported, with the name of the repository's short name. 
* If the `export-locatio`n is set a subdirectory for each repository exported will be created in that directory.
* If the `export-location` is set but, does not exist, aspace-export will attempt to create it.
* Within each repository directory there will be an `exports` directory containing all exported finding aids. If the --include-unpublished-resources flag is set a `unpublished` will be created in addition to the `exports` directory.
* A log file will be created named `aspace-export-[timestamp].log` which will be created in the root of output directory as defined in the --export-location option.
* A short summary report with statistics will be created named `aspace-export-report-[timestamp].txt` will be created in the root of output directory as defined in the --export-location option.

example output structure
------------------------
<pre>
/path/to/export-location/
        aspace-exports.log
        aspace-exports-report.txt
        /tamwag
                /exports
                        tam_001.xml
                        tam_002.xml
                /failures
                        tam_004.xml
</pre>

Command-Line Arguments
----------------------
--config, path/to/go-aspace.yml configuration file, required<br>
--environment, environment key in config file of the instance to export from, required<br>
--export-location, path/to/the location to export resources, default: `.`<br>
--format, format of export: ead or marc, default: `ead`<br>
--include-unpublished-resources, include unpublished resources in exports, default: `false`<br>
--include-unpublished-notes, include unpublished notes in exports, default: `false`<br>
--reformat, tab-reformat ead files (marcxml are tab-formatted by ArchivesSpace), default: `false`<br>
--repository, ID of the repository to be exported, `0` will export all repositories, default: `0`<br>
--resource, ID of the resource to be exported, `0` will export all resources, default: `0`<br>
--timeout, client timeout in seconds to, default: `20`<br>
--version, print the application and go-aspace client version<br>
--workers, number of concurrent export workers to create, default: `8`<br>
--help, print this help screen<br>

Exit Error Codes
----------------
0. no errors
1. could not create a log file to write to
2. mandatory options not set
3. the location set at export-location set does not exist or is not a directory
4. go-aspace library could not create an aspace-client 
5. could not get a list of repositories from ArchivesSpace
6. could not get a list of resources from ArchivesSpace
7. could not create a aspace-export directory at the location set at --export-location 
8. could not create subdirectories in the aspace-export 



