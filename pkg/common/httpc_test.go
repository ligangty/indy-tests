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
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetHost(t *testing.T) {
	Convey("GetHost should return correct hostname", t, func() {
		Convey("https://www.google.com host is www.google.com", func() {
			So(GetHost("https://www.google.com"), ShouldEqual, "www.google.com")
		})
		Convey("http://www.test.com host is www.test.com", func() {
			So(GetHost("http://www.test.com"), ShouldEqual, "www.test.com")
		})
	})
}

func TestGetPort(t *testing.T) {
	Convey("GetPort should return correct port", t, func() {
		Convey("https://www.google.com port is empty", func() {
			So(GetPort("https://www.google.com"), ShouldEqual, "")
		})
		Convey("http://www.test.com:8080 port is 8080", func() {
			So(GetPort("http://www.test.com:8080"), ShouldEqual, "8080")
		})
	})
}

func TestDownload(t *testing.T) {
	Convey("Download should succeed", t, func() {
		Convey("File can be download", func() {
			fileUrl := "https://repo.maven.apache.org/maven2/io/netty/netty-all/4.1.9.Final/netty-all-4.1.9.Final.pom"
			fileLoc := "/tmp/netty-all-4.1.9.Final.pom"
			So(FileOrDirExists(fileLoc), ShouldBeFalse)
			DownloadFile(fileUrl, fileLoc)
			So(FileOrDirExists(fileLoc), ShouldBeTrue)
			os.RemoveAll(fileLoc)
		})
	})
}

/*
 * We need to start up a local httproxy to run this test. For indy, run "{indy-home}/bin/test-setup.sh -e"
 * By default the httproxy is disabled. We need to update the conf.d/httproxy.conf as below
 * [httprox]
 * enabled=true
 *
 * You can test 'https' url by changing fileUrl from 'http://' to 'https://'.
 * Note: you need to enable the MITM config in the conf.d/httproxy.conf accordingly.
 *
 func TestDownloadByProxy(t *testing.T) {
	Convey("Download by proxy should succeed", t, func() {
		Convey("File can be download", func() {
			fileUrl := "http://registry.npmjs.org/npm/-/npm-6.14.5.tgz"
			fileLoc := "/tmp/npm-6.14.5.tgz"
			trackingId := "test-build-123"
			indyProxyUrl := "http://localhost:8081"
			So(FileOrDirExists(fileLoc), ShouldBeFalse)
			DownloadFileByProxy(fileUrl, fileLoc, indyProxyUrl, trackingId+TRACKING_SUFFIX, "pass")
			So(FileOrDirExists(fileLoc), ShouldBeTrue)
			os.RemoveAll(fileLoc)
		})
	})
}
*/
