/*
 *  Copyright (C) 2021-2023 Red Hat, Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *          http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package integrationtest

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	logger "github.com/sirupsen/logrus"

	"github.com/commonjava/indy-tests/pkg/buildtest"
	"github.com/commonjava/indy-tests/pkg/common"
	"github.com/commonjava/indy-tests/pkg/dataset"
	"github.com/commonjava/indy-tests/pkg/datest"
	"github.com/commonjava/indy-tests/pkg/promotetest"
)

const (
	DEFAULT_ROUTINES     = 4
	TMP_METADATA_DIR     = "/tmp/metadata"
	PROMOTE_TARGET_STORE = "pnc-builds"
)

/*
 * Run integration test. If dryRun is true, it prints the repo creation, file down/upload, promote, rollback,
 * and clean-up info without really doing them. Otherwise, it will run (in order):
 *
 * a. Load the corresponding dataset
 * b. Retrieve the metadata files in da.json to check they can be downloaded successfully
 * c. Create a temporary build group G (same as PNC build group), and a hosted repo A
 * d. Download the files in tracking "downloads" from the group G
 * e. Download the files in tracking "uploads" from origin, and rename all the files (jar, pom, and so on)
 *    with a new version suffix and upload them to the hosted repo A. Seal the folo record afterwards.
 * f. Retrieve the metadata files that will be affected by promotion, check that the new version not exists
 * g. Promote the files in hosted repo A to target hosted repo, e.g, pnc-builds
 * h. Retrieve the metadata files from step #f again, check that the new version is available
 * i. Rollback the promotion
 * j. Retrieve the metadata files from step #f again, check that the new version is gone
 * k. Clean up. Delete the build group G and the hosted repo A. Delete folo record.
 */
func Run(indyBaseUrl, datasetRepoUrl, buildId, promoteTargetStore, metaCheckRepo string, clearCache, dryRun, keepPod, sidecar bool, indyProxyUrl string) {
	if indyProxyUrl != "" {
		fmt.Println("Enable generic proxy: " + indyProxyUrl)
	}

	if sidecar {
		fmt.Println("Enable sidecar")
		indyBaseUrl = "http://localhost:8080"
	}

	//a. Clone dataset repo
	datasetRepoDir := cloneRepo(datasetRepoUrl)
	fmt.Printf("Clone SUCCESS, dir: %s\n", datasetRepoDir)

	//Load the info.json
	var info dataset.Info
	infoFileLoc := getInfoFileLoc(datasetRepoDir, buildId)
	json.Unmarshal(common.ReadByteFromFile(infoFileLoc), &info)

	//Load the additional-repos.json
	additionalRepos := getAdditionalRepos(datasetRepoDir, buildId)

	start := time.Now()

	//b. Retrieve the metadata files in da.json
	retrieveAlignmentMetadata(indyBaseUrl, datasetRepoDir, buildId, info)
	t := time.Now()
	fmt.Printf("Retrieve metadata SUCCESS, elapsed(s): %f\n", t.Sub(start).Seconds())

	//c/d/e. Create a mock build group, download files, rename to-be-uploaded files
	packageType := getPackageType(info)
	foloFileLoc := path.Join(datasetRepoDir, buildId, dataset.TRACKING_JSON)
	foloTrackContent := common.GetFoloRecordFromFile(foloFileLoc)
	originalIndy := getOriginalIndyBaseUrl(foloTrackContent.Uploads[0].LocalUrl)
	buildName := common.GenerateRandomBuildName()
	prev := t
	buildSuccess := buildtest.DoRun(originalIndy, indyBaseUrl, indyProxyUrl, packageType, buildName, foloTrackContent, additionalRepos, DEFAULT_ROUTINES, clearCache, dryRun)
	t = time.Now()
	fmt.Printf("Create mock group(%s) and download/upload SUCCESS, elapsed(s): %f\n", buildName, t.Sub(prev).Seconds())

	//k. Delete the temp group and the hosted repo, and folo record
	defer cleanUp(indyBaseUrl, packageType, buildName, dryRun)

	// Advanced checks
	if buildSuccess && !dryRun {
		if !verifyFoloRecord(indyBaseUrl, buildName, foloTrackContent) {
			return
		}
	}

	//f. Retrieve the metadata files which will be affected by promotion
	metaFiles := calculateMetadataFiles(foloTrackContent)
	metaFilesLoc := path.Join(TMP_METADATA_DIR, "before-promote")
	newVersionNum := buildName[len(common.BUILD_TEST_):]
	exists := true
	passed, e := retrieveMetadataAndValidate(indyBaseUrl, packageType, metaCheckRepo, metaFiles, metaFilesLoc, newVersionNum, !exists)
	if !passed {
		logger.Infof("Metadata validate failed (before). Errors: %s", e.Error())
		panic("Metadata validate failed")
	}
	fmt.Printf("Metadata validate (before) SUCCESS\n")

	//g. Promote the files in hosted repo A to hosted repo pnc-builds
	foloTrackId := buildName
	sourceStore, targetStore := getPromotionSrcTargetStores(packageType, buildName, promoteTargetStore, foloTrackContent)
	resp, _, success := promotetest.DoRun(indyBaseUrl, foloTrackId, sourceStore, targetStore, newVersionNum, foloTrackContent, dryRun)
	if !success {
		fmt.Printf("Promote failed, %s\n", resp)
		panic("Promote failed")
	}

	//h. Retrieve the metadata files again, check the new version
	fmt.Printf("Waiting 30s...\n")
	time.Sleep(30 * time.Second) // wait for Indy event handled

	metaFilesLoc = path.Join(TMP_METADATA_DIR, "after-promote")
	passed, e = retrieveMetadataAndValidate(indyBaseUrl, packageType, metaCheckRepo, metaFiles, metaFilesLoc, newVersionNum, exists)
	if !passed {
		logger.Infof("Metadata validate failed (after promotion). Errors: %s", e.Error())
		panic("Metadata validate failed")
	}
	fmt.Printf("Metadata validate (after promotion) SUCCESS\n")

	//i. Rollback the promotion
	fmt.Printf("Rollback:\n%s\n", resp)
	promotetest.Rollback(indyBaseUrl, resp, dryRun)

	//h. Retrieve the metadata files again, check the new version is GONE
	fmt.Printf("Waiting 30s...\n")
	time.Sleep(30 * time.Second)

	metaFilesLoc = path.Join(TMP_METADATA_DIR, "rollback")
	passed, e = retrieveMetadataAndValidate(indyBaseUrl, packageType, metaCheckRepo, metaFiles, metaFilesLoc, newVersionNum, !exists)
	if !passed {
		logger.Infof("Metadata validate failed (rollback). Errors: %s", e.Error())
		panic("Metadata validate failed")
	}
	fmt.Printf("Metadata validate (rollback) SUCCESS\n")

	// Pause and keep pod for debugging
	if keepPod {
		fmt.Printf("Waiting 30m...\n")
		time.Sleep(30 * time.Minute)
	}
}

func verifyFoloRecord(indyBaseUrl, buildName string, originalTrackContent common.TrackedContent) bool {
	trackedContent := common.GetFoloRecord(indyBaseUrl, buildName)
	// For debug
	// b, _ := json.MarshalIndent(trackedContent, "", "  ")
	// fmt.Println(string(b))

	if trackedContent.TrackingKey.Id != buildName {
		logger.Infof("Verify folo record FAILED! TrackingKey.Id, expected: %s, got: %s", buildName, trackedContent.TrackingKey.Id)
		return false
	}

	uploadErrors := checkFoloEntries("Uploads", common.AlterUploadPath, buildName, trackedContent.Uploads, originalTrackContent.Uploads)
	downloadErrors := checkFoloEntries("Downloads", nil, "", trackedContent.Downloads, originalTrackContent.Downloads)

	if len(uploadErrors) > 0 || len(downloadErrors) > 0 {
		return false
	}
	logger.Info("Verify folo record SUCCESS!")
	return true
}

func checkFoloEntries(title string, alterPath func(string, string, string) string, buildName string,
	entries, originalEntries []common.TrackedContentEntry) []string {

	logger.Infof("Verify folo records - %s", title)
	m := make(map[string]string) // key: path, value: md5
	for _, v := range entries {
		m[v.Path] = v.Md5
	}

	var errors []string
	// Check original paths are in the map and the md5 matches
	for _, v := range originalEntries {
		p := v.Path
		if alterPath != nil {
			p = alterPath(v.Path, v.StoreKey, buildName[len(common.BUILD_TEST_):])
		}
		// Not check metadata. e.g, when downloading a metadata through a group, folo will ingore it.
		if common.IsRegularFile(p) {
			if m[p] == "" {
				errors = append(errors, "[Missing] "+v.Path)
			} else if m[p] != v.Md5 {
				errors = append(errors, "[Md5-Error] "+v.Path)
			}
		}
	}

	if len(errors) > 0 {
		logger.Info("FAILED! Errors:")
		for _, v := range errors {
			fmt.Println(v)
		}
	}
	return errors
}

func getAdditionalRepos(datasetRepoDir, buildId string) []string {
	fileLoc := path.Join(datasetRepoDir, buildId, dataset.ADDITIONAL_REPOS)
	if !common.FileOrDirExists(fileLoc) {
		fileLoc = path.Join(datasetRepoDir, strings.Split(buildId, "/")[0], dataset.ADDITIONAL_REPOS) // Check group location
	}
	if common.FileOrDirExists(fileLoc) {
		byteValue := common.ReadByteFromFile(fileLoc)
		var arr []string
		json.Unmarshal([]byte(byteValue), &arr)
		return arr
	}
	return nil
}

func getInfoFileLoc(datasetRepoDir, buildId string) string {
	// Support group build dataset. E.g., in case of "11237/builds/###", info.json is in dir "11237/".
	if strings.Contains(buildId, "builds") {
		toks := strings.Split(buildId, "/")
		return path.Join(datasetRepoDir, toks[0], dataset.INFO_JSON)
	}
	return path.Join(datasetRepoDir, buildId, dataset.INFO_JSON)
}

func getPromotionSrcTargetStores(packageType, buildName, targetStoreName string, foloTrackContent common.TrackedContent) (string, string) {
	toks := strings.Split(foloTrackContent.Uploads[0].StoreKey, ":")
	sourceStore := fmt.Sprintf("%s:%s:%s", toks[0], toks[1], buildName)
	if targetStoreName == "" {
		targetStoreName = PROMOTE_TARGET_STORE
	}
	targetStore := packageType + ":hosted:" + targetStoreName
	fmt.Printf("Get promotion sourceStore: %s, targetStore: %s\n", sourceStore, targetStore)
	return sourceStore, targetStore
}

func calculateMetadataFiles(foloTrackContent common.TrackedContent) []string {
	paths := []string{}
	for _, up := range foloTrackContent.Uploads {
		if strings.HasSuffix(up.Path, ".pom") {
			versionsDir := path.Dir(up.Path)
			artifactDir := path.Dir(versionsDir)
			metadataPath := path.Join(artifactDir, common.MAVEN_METADATA_XML)
			paths = append(paths, metadataPath)
		}
	}
	return paths
}

func retrieveMetadataAndValidate(indyBaseUrl, packageType, metaCheckRepo string, metaFiles []string, filesLoc, versionNumber string, exist bool) (bool, error) {
	if metaCheckRepo == "" {
		fmt.Printf("Skip metadata check, no metaCheckRepo specified.\n")
		return true, nil
	}

	repoType := "group"
	repoName := metaCheckRepo

	// Also support the full repo name, e.g, maven:group:test-builds
	index := strings.Index(metaCheckRepo, ":")
	if index > 0 {
		toks := strings.Split(metaCheckRepo, ":")
		repoType = toks[1]
		repoName = toks[2]
	}

	// Download meta files
	for _, p := range metaFiles {
		url := common.GetIndyContentUrl(indyBaseUrl, packageType, repoType, repoName, p)
		common.DownloadFile(url, path.Join(filesLoc, p))
	}

	// Check version
	success := true
	var e common.MultiError
	for _, p := range metaFiles {
		file := path.Join(filesLoc, p)
		// read file and see if version exist
		if !common.FileOrDirExists(file) {
			fmt.Printf("Check metadata FAILED, file: %s, file not exists\n", file)
			success = false
			e.Append(p)
			continue
		} else {
			content := string(common.ReadByteFromFile(file))
			fmt.Printf("Check metadata, file: %s, content:\n%s\n", file, content)
			index := strings.Index(content, common.REDHAT_+versionNumber)
			isExist := false
			if index >= 0 {
				isExist = true
			}
			if isExist != exist {
				success = false
				e.Append(p)
			}
		}
	}
	return success, &e
}

func cloneRepo(datasetRepoUrl string) string {
	return common.DownloadRepo(datasetRepoUrl)
}

func retrieveAlignmentMetadata(indyBaseUrl, datasetRepoDir, buildId string, info dataset.Info) {
	fileLoc := path.Join(datasetRepoDir, buildId, dataset.DA_JSON)

	// Read jsonFile
	byteValue := common.ReadByteFromFile(fileLoc)

	// Parse it
	var arr []string
	json.Unmarshal([]byte(byteValue), &arr)

	var urls []string
	packageType := getPackageType(info)
	groupName := "DA"
	if info.TemporaryBuild {
		groupName = "DA-temporary-builds"
	}

	for _, v := range arr {
		u := common.GetIndyContentUrl(indyBaseUrl, packageType, "group", groupName, v)
		urls = append(urls, u)
	}

	if logger.IsLevelEnabled(logger.DebugLevel) {
		for _, v := range urls {
			fmt.Println(v)
		}
	}

	datest.LookupMetadataByRoutines(urls, DEFAULT_ROUTINES)
}

func getPackageType(info dataset.Info) string {
	packageType := "maven"
	if strings.EqualFold(info.BuildType, "NPM") {
		packageType = "npm"
	}
	return packageType
}

func getOriginalIndyBaseUrl(localUrl string) string {
	u, err := url.Parse(localUrl)
	if err != nil {
		panic(err)
	}
	return u.Scheme + "://" + u.Host
}

func cleanUp(indyBaseUrl, packageType, buildName string, dryRun bool) {
	if dryRun {
		fmt.Printf("Dry run cleanUp\n")
		return
	}

	buildtest.DeleteIndyTestRepos(indyBaseUrl, packageType, buildName)

	if common.DeleteFoloRecord(indyBaseUrl, buildName) {
		fmt.Printf("Delete folo record %s SUCCESS\n", buildName)
	} else {
		fmt.Printf("Delete folo record %s FAILED\n", buildName)
	}
}
