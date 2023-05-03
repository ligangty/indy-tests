package promotetest

import (
	"fmt"
	"os"

	"github.com/commonjava/indy-tests/pkg/common"
)

func Run(targetIndy, foloTrackId, targetStore string) {
	indyHost, validated := common.ValidateTargetIndy(targetIndy)
	if !validated {
		os.Exit(1)
	}

	indyURL := "http://" + indyHost
	foloTrackContent := common.GetFoloRecord(indyURL, foloTrackId)
	DoRun(indyURL, foloTrackId, "", targetStore, "", foloTrackContent, false)
}

func DoRun(indyBaseUrl, foloTrackId, sourceStore, targetStore, newVersionNum string,
	foloTrackContent common.TrackedContent, dryRun bool) (string, int, bool) {
	if foloTrackContent.Uploads == nil && len(foloTrackContent.Uploads) == 0 {
		fmt.Printf("There are not any uploads records in folo build %s, promotion will be ignored!\n", foloTrackId)
		return "", 200, true
	}

	paths := []string{}

	if sourceStore == "" {
		sourceStore = foloTrackContent.Uploads[0].StoreKey
	}

	for _, up := range foloTrackContent.Uploads {
		if common.IsMetadata(up.Path, up.StoreKey) {
			continue // ignore matedata
		}
		if newVersionNum == "" {
			paths = append(paths, up.Path)
		} else {
			// replace version with newVersionNum, e.g, xxx-redhat-### to xxx-redhat-<newVersion>
			altered := common.AlterUploadPath(up.Path, up.StoreKey, newVersionNum)
			paths = append(paths, altered)
		}
	}

	return promote(indyBaseUrl, foloTrackId, sourceStore, targetStore, paths, dryRun)
}
