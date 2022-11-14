package main

import (
	"fmt"
	"os"

	"github.com/commonjava/indy-tests/cmd/buildtest"
	"github.com/commonjava/indy-tests/cmd/dataset"
	"github.com/commonjava/indy-tests/cmd/datest"
	"github.com/commonjava/indy-tests/cmd/event"
	"github.com/commonjava/indy-tests/cmd/integrationtest"
	"github.com/commonjava/indy-tests/cmd/promotetest"
	"github.com/commonjava/indy-tests/cmd/statictest"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "indy-test",
		Short: "indy-test is a tool to do indy integration test against runnable indy server",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	rootCmd.AddCommand(buildtest.NewBuildTestCmd())
	rootCmd.AddCommand(promotetest.NewPromoteTestCmd())
	rootCmd.AddCommand(datest.NewDATestCmd())
	rootCmd.AddCommand(dataset.NewDatasetCmd())
	rootCmd.AddCommand(integrationtest.NewIntegrationTestCmd())
	rootCmd.AddCommand(event.NewEventTestCmd())
	rootCmd.AddCommand(statictest.NewStaticTestCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
