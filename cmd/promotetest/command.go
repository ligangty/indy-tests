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

package promotetest

import (
	"fmt"
	"os"

	"github.com/commonjava/indy-tests/pkg/promotetest"
	"github.com/spf13/cobra"
)

var targetIndy, trackingId, promoteTarget string

func NewPromoteTestCmd() *cobra.Command {

	exec := &cobra.Command{
		Use:   "promote $targetIndy $foloTrackId $promoteTarget",
		Short: "To do a promote test with an existed folo tracking report and an target indy hosted repo",
		Run: func(cmd *cobra.Command, args []string) {
			if !validate(args) {
				cmd.Help()
				os.Exit(1)
			}

			promotetest.Run(args[0], args[1], args[2])
		},
	}

	return exec
}

func validate(args []string) bool {
	if len(args) <= 2 {
		fmt.Printf("there are at least 3 non-empty arguments: targetIndy, trackingId, promoteTarget!\n\n")
		return false
	}
	return true
}
