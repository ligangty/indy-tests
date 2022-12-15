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
	"io/ioutil"
	"net/http"
)

func GetAlignLog(pncBaseUrl, buildId string) string {
	alignUrl := pncBaseUrl + "/pnc-rest/v2/builds/" + buildId + "/logs/align"
	req, err := http.NewRequest(http.MethodGet, alignUrl, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept", "text/plain")

	var c http.Client
	resp, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(responseData)
}
