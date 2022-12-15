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

package buildtest

import (
	"fmt"
	"os"
	"regexp"

	common "github.com/commonjava/indy-tests/pkg/common"
)

// Not used
func PrepareEntriesByLog(logUrl string) map[string][]string {
	log, err := common.GetRespAsPlaintext(logUrl)
	if err != nil {
		httpErr := err.(common.HTTPError)
		fmt.Printf("Request failed! Log url: %s, response status: %d, error message: %s\n", logUrl, httpErr.StatusCode, httpErr.Message)
		os.Exit(1)
	}
	result, err := ParseLog(log)
	if err != nil {
		fmt.Printf("Log parse failed! Log url: %s, error message: %s\n", logUrl, err.Error())
		os.Exit(1)
	}

	return result
}

func ParseLog(logCnt string) (map[string][]string, error) {
	if common.IsEmptyString(logCnt) {
		return nil, fmt.Errorf("The log content is empty!")
	}
	downloadR := regexp.MustCompile(`\[INFO\] Downloaded from indy-mvn:\s*(https{0,1}:\/\/.+)\s{1}(\(.+at.+\))`)
	uploadR := regexp.MustCompile(`\[INFO\] Uploaded to indy-mvn:\s*(https{0,1}:\/\/.+)\s{1}(\(.+at.+\))`)
	result := make(map[string][]string)
	downloads := collectEntries(downloadR, logCnt)
	if downloads != nil {
		result["downloads"] = downloads
	}
	uploads := collectEntries(uploadR, logCnt)
	if uploads != nil {
		result["uploads"] = uploads
	}
	return result, nil
}

func collectEntries(reg *regexp.Regexp, content string) []string {
	matches := reg.FindAllStringSubmatch(content, -1)
	if matches != nil {
		entries := make([]string, 0)
		for i := range matches {
			entries = append(entries, matches[i][1])
		}
		return entries
	}
	return nil
}
