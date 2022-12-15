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

package datest

import (
	"fmt"
	"os"
	"strconv"

	"github.com/commonjava/indy-tests/pkg/datest"

	"github.com/spf13/cobra"
)

var targetIndy, daGroup, dataDir string
var processNum int

func NewDATestCmd() *cobra.Command {

	exec := &cobra.Command{
		Use:   "datest $targetIndy $daGroup $processNum",
		Short: "To do a da test based on the alignment logs from PNC build",
		Run: func(cmd *cobra.Command, args []string) {
			if !validate(args) {
				cmd.Help()
				os.Exit(1)
			}
			processNum, err := strconv.Atoi(args[3])
			if err == nil {
				fmt.Println(processNum)
			}
			datest.Run(args[0], args[1], args[2], processNum)
		},
	}

	return exec
}

func validate(args []string) bool {
	if len(args) <= 2 {
		fmt.Printf("there are at least 4 non-empty arguments: targetIndy, daGroup, dataDir, processNum!\n\n")
		return false
	}
	return true
}
