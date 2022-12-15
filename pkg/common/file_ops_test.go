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
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFileExists(t *testing.T) {
	Convey("TestFileExists", t, func() {
		Convey("File should exists", func() {
			So(FileOrDirExists("/usr/bin/bash"), ShouldBeTrue)
		})
		Convey("File should not exists", func() {
			So(FileOrDirExists("/kljsdflksdjf"), ShouldBeFalse)
		})
	})

}

func nowInMillis() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
