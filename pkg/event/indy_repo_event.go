package event

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	common "github.com/commonjava/indy-tests/pkg/common"
)

const (
	DEFAULT_SHARED_GROUP       = "builds-untested+shared-imports+public"
	DEFAULT_MVN_CENTRAL        = "central"
	DEFAULT_NPM_CENTRAL        = "npmjs"
	TYPE_MVN                   = "maven"
	TYPE_NPM                   = "npm"
	MISSING_CONTENT_PATH       = "org/missing/1.0/missing-1.0.pom"
	ACCESSIBLE_REMOTE_PATH     = "org/apache/apache/2/apache-2.pom"
	GOING_MERGED_HOSTED_PATH   = "org/apache/apache/666/apache-666.pom"
	MERGED_MAVEN_METADATA_PATH = "org/apache/apache/maven-metadata.xml"
	REMOTE_VERSION_TAG         = "<version>2</version>"
	LATEST_HOSTED_VERSION_TAG  = "<latest>666</latest>"
	HOSTED_POM_CONTENT         = "<project><modelVersion>4.0.0</modelVersion><groupId>org.apache</groupId><artifactId>apache</artifactId><version>666</version><packaging>pom</packaging></project>"
)

type BuildMetadata struct {
	buildType     string
	centralName   string
	sharedGrpName string
}

func decideMeta(buildType string) *BuildMetadata {
	if buildType == TYPE_MVN {
		return &BuildMetadata{
			buildType:     buildType,
			centralName:   DEFAULT_MVN_CENTRAL,
			sharedGrpName: DEFAULT_SHARED_GROUP,
		}
	} else if buildType == TYPE_NPM {
		return &BuildMetadata{
			buildType:     buildType,
			centralName:   DEFAULT_NPM_CENTRAL,
			sharedGrpName: DEFAULT_SHARED_GROUP,
		}
	}
	return nil
}

func prepareIndyRepos(indyURL, buildName string, buildMeta BuildMetadata, additionalRepos []string, dryRun bool) {
	if dryRun {
		fmt.Printf("Dry run prepareIndyRepos\n")
		return
	}
	prepareIndyHosted(indyURL, buildMeta.buildType, buildName, false)
	prepareIndyHostedUploadContent(indyURL, buildMeta.buildType, buildName)
	prepareIndyRemote(indyURL, buildMeta.buildType, buildName)
	prepareIndyGroup(indyURL, buildName, buildMeta, additionalRepos)
}

func prepareIndyHosted(indyURL, buildType, buildName string, disabled bool) {
	hostedVars := IndyHostedVars{
		Name:     buildName,
		Type:     buildType,
		Disabled: disabled,
	}
	URL := fmt.Sprintf("%s/api/admin/stores/%s/hosted/%s", indyURL, buildType, buildName)

	hosted := IndyHostedTemplate(&hostedVars)
	fmt.Printf("Start creating/updating hosted repo %s, disabled: %v\n", buildName, disabled)
	result := putRequest(URL, strings.NewReader(hosted))
	if !result {
		fmt.Printf("Error: Failed to create/update hosted repo %s, disabled: %v.\n\n", buildName, disabled)
		os.Exit(1)
	}
	fmt.Printf("Create/Update hosted repo %s successfully, disabled: %v\n", buildName, disabled)
}

// Prepare the content to be merged
func prepareIndyHostedUploadContent(indyURL, buildType, buildName string) {
	URL := fmt.Sprintf("%s/api/content/%s/hosted/%s/%s", indyURL, buildType, buildName, GOING_MERGED_HOSTED_PATH)
	fmt.Printf("Upload going merged content to hosted repo, path: %s\n", URL)
	result := putRequest(URL, strings.NewReader(HOSTED_POM_CONTENT))
	if !result {
		fmt.Printf("Error: Failed to upload content to hosted repo %s.\n\n", buildName)
		os.Exit(1)
	}
	fmt.Printf("Upload content to hosted repo %s successfully\n", buildName)
}

func prepareIndyRemote(indyURL, buildType, buildName string) {
	fmt.Println("Start remote repo creation validation.")
	fmt.Printf("==========================================\n\n")

	remoteVars := IndyRemoteVars{
		Name: buildName,
		Type: buildType,
	}
	URL := fmt.Sprintf("%s/api/admin/stores/%s/remote/%s", indyURL, buildType, buildName)

	remote := IndyRemoteTemplate(&remoteVars)
	fmt.Printf("Start creating remote repo %s\n", buildName)
	result := putRequest(URL, strings.NewReader(remote))
	if !result {
		fmt.Printf("Error: Failed to create remote repo %s.\n\n", buildName)
		os.Exit(1)
	}
	fmt.Printf("Create remote repo %s successfully\n", buildName)

	fmt.Printf("Waiting 60s...\n")
	time.Sleep(60 * time.Second)

	// Verify the remote content path after repo creation
	remoteContentURL := fmt.Sprintf("%s/api/content/%s/remote/%s/%s", indyURL, buildType, buildName, ACCESSIBLE_REMOTE_PATH)
	_, _, contentResult := getRequest(remoteContentURL)
	if !contentResult {
		fmt.Printf("Error: Failed to get content %s in remote %s.\n\n", remoteContentURL, buildName)
		os.Exit(1)
	}
	fmt.Printf("Get remote content %s successfully\n", remoteContentURL)
	fmt.Println("==========================================")
	fmt.Printf("Finish remote repo creation validation.\n\n")
}

func prepareIndyGroup(indyURL, buildName string, buildMeta BuildMetadata, additionalRepos []string) {
	fmt.Println("Start group repo creation validation.")
	fmt.Printf("==========================================\n\n")

	var constituents []string
	buildType := buildMeta.buildType
	hostedChild := fmt.Sprintf("%s:hosted:%s", buildType, buildName)
	remoteChild := fmt.Sprintf("%s:remote:%s", buildType, buildName)
	constituents = append(constituents, hostedChild)

	URL := fmt.Sprintf("%s/api/admin/stores/%s/group/%s", indyURL, buildType, buildName)

	created := createOrUpdateGroupRepo(URL, buildName, buildMeta, additionalRepos, constituents)
	if !created {
		fmt.Printf("Error: Failed to create group repo %s.\n\n", buildName)
		os.Exit(1)
	}
	fmt.Printf("Create group repo %s successfully\n", buildName)

	// Update group to add remote constituent
	constituents = append(constituents, remoteChild)
	updated := createOrUpdateGroupRepo(URL, buildName, buildMeta, additionalRepos, constituents)
	if !updated {
		fmt.Printf("Error: Failed to update group repo %s.\n\n", buildName)
		os.Exit(1)
	}
	fmt.Printf("Update group repo %s successfully\n", buildName)

	fmt.Printf("Waiting 60s...\n")
	time.Sleep(60 * time.Second)

	// Verify the merged path and metadata after remote constituent update
	grpContentURL := fmt.Sprintf("%s/api/content/%s/group/%s/%s", indyURL, buildType, buildName, ACCESSIBLE_REMOTE_PATH)
	_, _, result := getRequest(grpContentURL)
	if !result {
		fmt.Printf("Error: Failed to get merged path %s in group %s.", grpContentURL, buildName)
		os.Exit(1)
	}
	fmt.Printf("Get affected group merged path %s successfully\n", grpContentURL)

	grpMetadataURL := fmt.Sprintf("%s/api/content/%s/group/%s/%s", indyURL, buildType, buildName, MERGED_MAVEN_METADATA_PATH)
	metadata, _, mergedResult := getRequest(grpMetadataURL)
	if !mergedResult {
		fmt.Printf("Error: Failed to get group metadata, path: %s.\n\n", grpMetadataURL)
		os.Exit(1)
	}
	index := strings.Index(metadata, REMOTE_VERSION_TAG)
	if index < 0 {
		fmt.Printf("Error: Failed to get correct merged metadata content, path: %s.\n\n", grpMetadataURL)
		os.Exit(1)
	}
	fmt.Printf("Get correct merged metadata content successfully, path: %s\n", grpMetadataURL)
	fmt.Println("==========================================")
	fmt.Printf("Finish group repo creation validation.\n\n")
}

func createOrUpdateGroupRepo(URL, buildName string, buildMeta BuildMetadata, additionalRepos, childRepos []string) bool {
	var constituents []string
	if childRepos != nil {
		constituents = append(constituents, childRepos...)
	}
	if additionalRepos != nil {
		constituents = append(constituents, additionalRepos...)
	}

	groupVars := IndyGroupVars{
		Name:         buildName,
		Type:         buildMeta.buildType,
		Constituents: constituents,
	}
	group := IndyGroupTemplate(&groupVars)

	fmt.Printf("Start creating/updating group repo %s\n", buildName)
	return putRequest(URL, strings.NewReader(group))
}

func DeleteIndyRepos(indyURL, packageType, buildName string, uploads map[string][]string) {
	if !delAllowed(buildName) {
		return
	}
	deleteIndyHosted(indyURL, packageType, buildName, uploads)
	deleteIndyRemote(indyURL, packageType, buildName)
	deleteIndyGroup(indyURL, packageType, buildName)
}

func delAllowed(buildName string) bool {
	if strings.HasPrefix(buildName, common.BUILD_TEST_) {
		return true
	}
	fmt.Printf("Can not delete repo(s) %s (not test repo)", buildName)
	return false
}

// Verify cleanup for the hosted repo deleting
func deleteIndyHosted(indyURL, buildType, repoName string, uploads map[string][]string) {
	fmt.Println("Start hosted repo cleanup.")
	fmt.Printf("==========================================\n\n")
	storePath := fmt.Sprintf("%s/%s/%s", buildType, "hosted", repoName)
	hostedRepo := fmt.Sprintf("%s/api/admin/stores/%s", indyURL, storePath)
	_, _, result := getRequest(hostedRepo)
	if !result {
		fmt.Printf("Error: Failed to get hosted repo %s.\n\n", repoName)
		os.Exit(1)
	}

	fmt.Printf("Waiting 60s...\n")
	time.Sleep(60 * time.Second)
	// Verify the contents, merged paths and metadata before deleting hosted repo
	for _, upload := range uploads {
		targetPath := strings.Split(upload[2], storePath)[1]
		contentURL := fmt.Sprintf("%s/api/content/%s%s", indyURL, storePath, targetPath)
		_, _, result := getRequest(contentURL)
		if !result {
			fmt.Printf("Error: Failed to get hosted content %s.\n\n", contentURL)
			os.Exit(1)
			break
		}
		grpContentURL := fmt.Sprintf("%s/api/content/%s/group/%s%s", indyURL, buildType, repoName, targetPath)
		_, _, result = getRequest(grpContentURL)
		if !result {
			fmt.Printf("Error: Failed to get merged path %s in group %s.\n\n", grpContentURL, repoName)
			os.Exit(1)
			break
		}
	}
	fmt.Printf("Get all hosted contents and affected group merged paths successfully\n")

	grpMetadataURL := fmt.Sprintf("%s/api/content/%s/group/%s/%s", indyURL, buildType, repoName, MERGED_MAVEN_METADATA_PATH)
	metadata, _, mergedResult := getRequest(grpMetadataURL)
	if !mergedResult {
		fmt.Printf("Error: Failed to get group metadata, path: %s.\n\n", grpMetadataURL)
		os.Exit(1)
	}
	index := strings.Index(metadata, LATEST_HOSTED_VERSION_TAG)
	if index < 0 {
		fmt.Printf("Error: Failed to get correct merged metadata content, path: %s.\n\n", grpMetadataURL)
		os.Exit(1)
	}
	fmt.Printf("Get correct merged metadata content successfully, path: %s\n", grpMetadataURL)

	// Delete hosted repo
	deleteURL := fmt.Sprintf("%s/api/admin/stores/%s/hosted/%s?deleteContent=true", indyURL, buildType, repoName)
	fmt.Printf("Start deleting hosted repo %s\n", repoName)
	result = delRequest(deleteURL)
	if result {
		fmt.Printf("Delete hosted repo %s successfully\n", repoName)
	}

	// Recreate hosted repo
	prepareIndyHosted(indyURL, buildType, repoName, false)

	_, _, result = getRequest(hostedRepo)
	if !result {
		fmt.Printf("Error: Failed to get hosted repo %s after recreating.\n\n", repoName)
		os.Exit(1)
	}

	fmt.Printf("Waiting 60s...\n")
	time.Sleep(60 * time.Second)

	// Verify the contents after deleting
	for _, upload := range uploads {
		targetPath := strings.Split(upload[2], storePath)[1]
		contentURL := fmt.Sprintf("%s/api/content/%s%s", indyURL, storePath, targetPath)
		_, _, result := getRequest(contentURL)
		if result {
			fmt.Printf("Error: Content %s is still existed after repo %s is removed.\n\n", contentURL, repoName)
			os.Exit(1)
			break
		}
	}
	fmt.Printf("Remove all hosted contents successfully\n")

	// Remove hosted repo
	fmt.Printf("Start deleting hosted repo %s\n", repoName)
	result = delRequest(deleteURL)
	if result {
		fmt.Printf("Delete hosted repo %s successfully\n", repoName)
	}

	fmt.Printf("Waiting 60s...\n")
	time.Sleep(60 * time.Second)

	// Verify affected group cleanup
	verifyHostedAffectedGroupCleanup(indyURL, buildType, repoName, uploads)
	fmt.Println("==========================================")
	fmt.Printf("Finish hosted repo cleanup.\n\n")
}

// Verify cleanup for the remote repo deleting
func deleteIndyRemote(indyURL, buildType, repoName string) {
	fmt.Println("Start remote repo cleanup.")
	fmt.Printf("==========================================\n\n")
	remotedRepo := fmt.Sprintf("%s/api/admin/stores/%s/remote/%s", indyURL, buildType, repoName)
	_, _, result := getRequest(remotedRepo)
	if !result {
		fmt.Printf("Error: Failed to get remote repo %s.\n\n", repoName)
		os.Exit(1)
	}

	// Verify NFC creation
	remoteMissingURL := fmt.Sprintf("%s/api/content/%s/remote/%s/%s", indyURL, buildType, repoName, MISSING_CONTENT_PATH)
	_, missingCode, _ := getRequest(remoteMissingURL)
	// 	nfcVerified := false
	if missingCode == 404 {
		// 		nfcVerified = true
		isCached := isNFCCached(indyURL, buildType, "remote", repoName)
		if !isCached {
			fmt.Printf("Error: Failed to cache NFC for remote repo %s.\n\n", repoName)
			os.Exit(1)
		}
		fmt.Printf("Remote repo %s NFC caches successfully\n", repoName)
	}

	// Delete remote repo
	URL := fmt.Sprintf("%s/api/admin/stores/%s/remote/%s?deleteContent=true", indyURL, buildType, repoName)
	fmt.Printf("Start deleting remote repo %s\n", repoName)
	result = delRequest(URL)
	if result {
		fmt.Printf("Delete remote repo %s successfully\n", repoName)
	}

	fmt.Printf("Waiting 60s...\n")
	time.Sleep(60 * time.Second)

	// Verify NFC cleanup
	// 	if nfcVerified {
	// 		isCached := isNFCCached(indyURL, buildType, "remote", repoName)
	// 		if isCached {
	// 			fmt.Printf("Error: Failed to remove NFC cache for remote repo %s.\n\n", repoName)
	// 			os.Exit(1)
	// 		}
	// 		fmt.Printf("Remove NFC for remote repo %s successfully\n", repoName)
	// 	}

	// Verify affected group cleanup
	verifyRemoteAffectedGroupCleanup(indyURL, buildType, repoName)
	fmt.Println("==========================================")
	fmt.Printf("Finish remote repo cleanup.\n\n")
}

// Verify cleanup for the group repo deleting
func deleteIndyGroup(indyURL, buildType, repoName string) {
	URL := fmt.Sprintf("%s/api/admin/stores/%s/group/%s", indyURL, buildType, repoName)
	fmt.Printf("Start deleting group repo %s\n", repoName)
	result := delRequest(URL)
	if result {
		fmt.Printf("Delete group repo %s successfully\n", repoName)
	}
}

func verifyHostedAffectedGroupCleanup(indyURL, buildType, repoName string, uploads map[string][]string) {
	// Verify constituent is removed from the affected group
	verifyGroupConstituents(indyURL, buildType, "hosted", repoName)

	// Verify the merged paths are removed from the affected group
	for _, upload := range uploads {
		childStorePath := fmt.Sprintf("%s/%s/%s", buildType, "hosted", repoName)
		targetPath := strings.Split(upload[2], childStorePath)[1]
		grpContentURL := fmt.Sprintf("%s/api/content/%s/group/%s%s", indyURL, buildType, repoName, targetPath)
		_, _, result := getRequest(grpContentURL)
		if result {
			fmt.Printf("Error: Content %s is still existed in group %s.\n\n", grpContentURL, repoName)
			os.Exit(1)
			break
		}
	}
	fmt.Printf("Remove all hosted contents from the affected group successfully\n")

	grpMetadataURL := fmt.Sprintf("%s/api/content/%s/group/%s/%s", indyURL, buildType, repoName, MERGED_MAVEN_METADATA_PATH)
	metadata, _, mergedResult := getRequest(grpMetadataURL)
	if !mergedResult {
		fmt.Printf("Error: Failed to get group metadata, path: %s.\n\n", grpMetadataURL)
		os.Exit(1)
	}
	index := strings.Index(metadata, LATEST_HOSTED_VERSION_TAG)
	if index >= 0 {
		fmt.Printf("Error: Failed to remove version from the merged metadata content, path: %s.\n\n", grpMetadataURL)
		os.Exit(1)
	}
	fmt.Printf("Remove version from the merged metadata content successfully, path: %s\n", grpMetadataURL)
}

func verifyRemoteAffectedGroupCleanup(indyURL, buildType, repoName string) {
	// Verify constituent is removed from the affected group
	verifyGroupConstituents(indyURL, buildType, "remote", repoName)

	// Verify the merged path and metadata is removed from the affected group
	grpContentURL := fmt.Sprintf("%s/api/content/%s/group/%s/%s", indyURL, buildType, repoName, ACCESSIBLE_REMOTE_PATH)
	_, _, result := getRequest(grpContentURL)
	if result {
		fmt.Printf("Error: Content %s is still existed in group %s.\n\n", grpContentURL, repoName)
		os.Exit(1)
	}
	fmt.Printf("Remove remote content from the affected group successfully\n")
}

func verifyGroupConstituents(indyURL, buildType, storeType, repoName string) {
	groupURL := fmt.Sprintf("%s/api/admin/stores/%s/group/%s", indyURL, buildType, repoName)
	groupBody, _, result := getRequest(groupURL)
	if !result {
		fmt.Printf("Error: Failed to get group repo %s.\n\n", repoName)
		os.Exit(1)
	}

	var group map[string]interface{}
	err := json.Unmarshal([]byte(groupBody), &group)
	if err != nil {
		fmt.Printf("Error: Group %s parse failed! Error message: %s.\n\n", repoName, err.Error())
		os.Exit(1)
	}
	constituents := []string{fmt.Sprint(group["constituents"])}
	repoKey := strings.Join([]string{buildType, storeType, repoName}, ":")
	if common.Contains(constituents, repoKey) {
		fmt.Printf("Error: Failed to remove %s repo %s from group repo %s.\n\n", storeType, repoName, repoName)
		os.Exit(1)
	}
	fmt.Printf("%s repo %s is removed from Group repo %s\n", storeType, repoName, repoName)
}

func isNFCCached(indyURL, buildType, storeType, repoName string) bool {
	nfcURL := fmt.Sprintf("%s/api/nfc/%s/%s/%s", indyURL, buildType, storeType, repoName)
	nfcContent, _, result := getRequest(nfcURL)
	fmt.Printf("NFC URL: %s\n", nfcURL)
	if !result {
		fmt.Printf("Don't get NFC for store key %s:%s:%s\n", buildType, storeType, repoName)
		return false
	}
	index := strings.Index(nfcContent, MISSING_CONTENT_PATH)
	if index < 0 {
		fmt.Printf("Don't find missing content in store %s:%s:%s NFC caches\n", buildType, storeType, repoName)
		return false
	}
	return true
}

func updateIndyReposEnablement(indyURL, packageType, buildName string) {
	fmt.Println("Start repo enablement/disablement cleanup.")
	fmt.Printf("==========================================\n\n")

	grpContentURL := fmt.Sprintf("%s/api/content/%s/group/%s/%s", indyURL, packageType, buildName, GOING_MERGED_HOSTED_PATH)
	grpMetadataURL := fmt.Sprintf("%s/api/content/%s/group/%s/%s", indyURL, packageType, buildName, MERGED_MAVEN_METADATA_PATH)

	// Disable the hosted repo
	prepareIndyHosted(indyURL, packageType, buildName, true)

	fmt.Printf("Waiting 60s...\n")
	time.Sleep(60 * time.Second)

	metadata, _, _ := getRequest(grpMetadataURL)
	index := strings.Index(metadata, LATEST_HOSTED_VERSION_TAG)
	if index >= 0 {
		fmt.Printf("Error: Failed to remove version from the merged metadata content, path: %s.\n\n", grpMetadataURL)
		os.Exit(1)
	}
	fmt.Printf("Remove version from the merged metadata content successfully, path: %s\n", grpMetadataURL)

	// Enable the hosted repo
	prepareIndyHosted(indyURL, packageType, buildName, false)

	fmt.Printf("Waiting 60s...\n")
	time.Sleep(60 * time.Second)

	// Verify the merged path and metadata after hosted enabled
	_, _, result := getRequest(grpContentURL)
	if !result {
		fmt.Printf("Error: Failed to merge content into group, path: %s.\n\n", grpContentURL)
		os.Exit(1)
	}
	fmt.Printf("Merge content into group successfully, path: %s\n", grpContentURL)

	metadata, _, _ = getRequest(grpMetadataURL)
	index = strings.Index(metadata, LATEST_HOSTED_VERSION_TAG)
	if index < 0 {
		fmt.Printf("Error: Failed to merge version into metadata content, path: %s.\n\n", grpMetadataURL)
		os.Exit(1)
	}
	fmt.Printf("Merge version into metadata content successfully, path: %s\n", grpMetadataURL)
	fmt.Println("==========================================")
	fmt.Printf("Finish repo enablement/disablement cleanup.\n\n")
}

func getRequest(url string) (string, int, bool) {
	content, code, succeeded := common.HTTPRequest(url, common.MethodGet, nil, true, nil, nil, "", false)
	debugFailureRequest(succeeded, content)
	return content, code, succeeded
}

var authenticator = common.DecideAuthenticator()

func postRequest(url string, data io.Reader) (string, bool) {
	content, _, succeeded := common.HTTPRequest(url, common.MethodPost, authenticator, true, data, nil, "", false)
	debugFailureRequest(succeeded, content)
	return content, succeeded
}

func putRequest(url string, data io.Reader) bool {
	content, _, succeeded := common.HTTPRequest(url, common.MethodPut, authenticator, false, data, nil, "", false)
	debugFailureRequest(succeeded, content)
	return succeeded
}

func delRequest(url string) bool {
	content, _, succeeded := common.HTTPRequest(url, common.MethodDelete, authenticator, false, nil, nil, "", false)
	debugFailureRequest(succeeded, content)
	return succeeded
}

func debugFailureRequest(succeeded bool, respText string) {
	if !succeeded {
		fmt.Printf("Debug for respText: %s\n", respText)
	}
}
