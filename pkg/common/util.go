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

package common

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	BUILD_TEST_           = "build-test-"
	ENVAR_TEST_MOUNT_PATH = "TEST_MOUNT_PATH"
	MAVEN_METADATA_XML    = "maven-metadata.xml"
	REDHAT_               = "redhat-"
)

var (
	versionTgzRegexp  = regexp.MustCompile(`(-([0-9]+))?\.tgz$`)
	versionRegexp     = regexp.MustCompile(`redhat-([0-9]+)`)
	regularFileRegexp = regexp.MustCompile(`(\.(gz|tgz|jar)|pom.xml)$`)
)

func RePanic(e error) {
	if e != nil {
		panic(e)
	}
}

func AlterUploadPath(rawPath, storeKey, newReleaseNumber string) string {
	if strings.HasPrefix(storeKey, "npm") {
		// Not sure whether it covers all cases.
		// npm path is like "/@redhat/opossum/-/opossum-6.2.1.tgz", replace to sth like '/@redhat/opossum/-/opossum-6.2.1-94465.tgz'
		return versionTgzRegexp.ReplaceAllString(rawPath, "-"+newReleaseNumber+".tgz")
	}
	return versionRegexp.ReplaceAllString(rawPath, REDHAT_+newReleaseNumber) // replace with new rel number
}

// generate a random 5 digit number for a build repo like "build-test-9xxxx"
func GenerateRandomBuildName() string {
	rand.Seed(time.Now().UnixNano())
	min := 90000
	max := 99999
	return fmt.Sprintf(BUILD_TEST_+"%v", rand.Intn(max-min)+min)
}

type MultiError struct {
	errors []string
}

func (e *MultiError) Error() string {
	return strings.Join(e.errors, ", ")
}

func (e *MultiError) Append(err string) {
	e.errors = append(e.errors, err)
}

func IsMetadata(path, storeKey string) bool {
	if strings.HasPrefix(storeKey, "npm") {
		return !strings.HasSuffix(path, ".tgz") // we consider files not ending with .tgz as npm metadata
	}
	return strings.Index(path, MAVEN_METADATA_XML) > 0
}

func IsRegularFile(fileLoc string) bool {
	return regularFileRegexp.MatchString(fileLoc)
}

func Md5Check(fileLoc, md5str string) {
	// Skip non-regular (i.e, metadata) files
	if IsRegularFile(fileLoc) {
		f, err := os.Open(fileLoc)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		h := md5.New()
		if _, err := io.Copy(h, f); err != nil {
			log.Fatal(err)
		}
		calculated := fmt.Sprintf("%x", h.Sum(nil))
		match := strings.EqualFold(md5str, calculated)
		if !match {
			log.Fatal(fmt.Sprintf("Md5 not match, expected: %s, calculated: %s", md5str, calculated))
		}
	}
}
