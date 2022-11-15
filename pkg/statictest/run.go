package buildtest

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	common "github.com/commonjava/indy-tests/pkg/common"
)

const (
	TMP_DOWNLOAD_DIR = "/tmp/download"
	TMP_UPLOAD_DIR   = "/tmp/upload"
	PROXY_           = "proxy-"
)

func Run(originalIndy, foloId, staticIndy string, processNum int) {
	origIndy := originalIndy
	if !strings.HasPrefix(origIndy, "http") {
		origIndy = "http://" + origIndy
	}
	foloTrackContent := common.GetFoloRecord(origIndy, foloId)
	DoRun(originalIndy, staticIndy, foloTrackContent, processNum, false)
}

// Refer the original indy folo track entries to download from static-proxy indy server
func DoRun(originalIndy, staticIndy string, foloTrackContent common.TrackedContent,
	processNum int, dryRun bool) bool {

	common.ValidateTargetIndyOrExit(originalIndy)
	common.ValidateTargetIndyOrExit(staticIndy)

	trackingId := foloTrackContent.TrackingKey.Id
	downloadDir := prepareDownloadDirectories(trackingId)
	downloads := prepareDownloadEntriesByFolo(staticIndy, foloTrackContent)
	failedArtifacts := make(map[string]int)
	var mu sync.Mutex

	downloadFunc := func(md5str, originalArtiURL, targetArtiURL string) bool {
		fileLoc := path.Join(downloadDir, path.Base(targetArtiURL))
		if dryRun {
			fmt.Printf("Dry run download, url: %s\n", targetArtiURL)
			return true
		}
		success, status := common.DownloadFile(targetArtiURL, fileLoc)
		if success {
			common.Md5Check(fileLoc, md5str)
		} else {
			mu.Lock()
			failedArtifacts[targetArtiURL] = status
			mu.Unlock()
		}
		return success
	}

	if len(downloads) > 0 {
		fmt.Println("Start handling downloads artifacts.")
		fmt.Printf("==========================================\n\n")
		if processNum > 1 {
			common.ConcurrentRun(processNum, downloads, downloadFunc)
		} else {
			for _, down := range downloads {
				downloadFunc(down[0], down[1], down[2])
			}
		}
		fmt.Println("==========================================")
		broken := false
		for art, status := range failedArtifacts {
			if status > 400 {
				if status == http.StatusNotFound {
					fmt.Printf("%s is not found in the static proxy server. \n", art)
				} else {
					fmt.Printf("%s can not be downloaded with error status: %v. \n", art, status)
					broken = true
				}
			}
		}
		fmt.Println("==========================================")
		if broken {
			fmt.Printf("Build test failed due to some downloading errors. Please see above logs to see the details.\n\n")
			os.Exit(1)
		}
		fmt.Printf("Downloads artifacts handling finished.\n\n")
	}

	return true
}

// For downloads entries, we will get the paths and inject them to the final url of target indy
// as they should be directly download from target indy.
func prepareDownloadEntriesByFolo(targetIndyURL string,
	foloRecord common.TrackedContent) map[string][]string {
	targetIndy := normIndyURL(targetIndyURL)
	result := make(map[string][]string)
	for _, down := range foloRecord.Downloads {
		var p string
		downUrl := ""
		p = path.Join("api/content/maven/group/static", down.Path)
		downUrl = fmt.Sprintf("%s%s", targetIndy, p)
		result[down.Path] = []string{down.Md5, "", downUrl}
	}
	return result
}

func normIndyURL(indyURL string) string {
	indy := indyURL
	if !strings.HasPrefix(indy, "http") {
		indy = "http://" + indy
	}
	if !strings.HasSuffix(indy, "/") {
		indy = indy + "/"
	}
	return indy
}

func prepareDownloadDirectories(buildId string) string {
	// use "/tmp/download", which will be dropped after each run
	downloadDir := TMP_DOWNLOAD_DIR
	if !common.FileOrDirExists(downloadDir) {
		os.MkdirAll(downloadDir, os.FileMode(0755))
	}
	if !common.FileOrDirExists(downloadDir) {
		fmt.Printf("Error: cannot create directory %s for file downloading.\n", downloadDir)
		os.Exit(1)
	}

	fmt.Printf("Prepared download dir: %s", downloadDir)
	return downloadDir
}
