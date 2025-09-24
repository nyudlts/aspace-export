aspace-export, v1.1.2
=====================
Command-line utility for bulk export, validation and reformatting of EAD finding aids and MARC records from Archivespace.

Install From Binary
-------------------
1. Download the latest binary for Mac or linux https://github.com/nyudlts/aspace-export/releases/tag/v1.1.2
3. Enter your ArchivesSpace credentials into the go-aspace.yml file included in the zip.

Install With Go
---------------
$ go install github.com/nyudlts/aspace-export

Build From Source
-----------------
Pre-requisite: libxml2 and libxml2 development headers<br>
$ make build<br>
$ sudo make install //installs aspace-export to /usr/local/bin

Package Distribution
-----------------
$ make package VERSION="release version" OS="osx,linux", this will build the app, zip the app, the sample go-aspace.yml file and the readme into a zip located in the /bin directory<br>
example: $ make package VERSION="v1.0.0b" OS="linux", this will create /bin/linux/aspace-export-linux-v1.0.0b.zip

Run
---
$ aspace-export --config /path/to/go-aspace.yml --environment your-environment-key --format ead-or-marc [options] 
<br><br>**notes:**
* If no esport-location is set the program will create a directory hierarchy at the in the current working directory named: `aspace-export-[timestamp]. A subdirectory will be created for each repository that was exported, with the name of the repository's slug. If the export-location is set a suubdirectoy for each repository exported will be created in that directory.
* Within each repository directory there will be an `exports` directory containing all exported finding aids and a `failures` directory for any file that fails to export from ArchivesSpace
* A log file will be created named `aspace-export.log` which will be created in the root of output directory as defined in the --export-location option.
* A Report with statistics will be created named `aspace-export-report.txt` will be created in the root of output directory as defined in the --export-location option.

**example output structure**<br>
/path/top/eexport-location/aspace-exports-[timestamp]<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;aspace-exports.log<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;aspace-exports-report.txt<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;/tamwag<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;/exports<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;tam_001.xml<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;tam_002.xml<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;/invalid<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;tam_003.xml<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;/failures<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;tam_004.xml<br>

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
7. could not create a aspace-export directory at    the location set at --export-location 
8. could not create subdirectories in the aspace-export 



