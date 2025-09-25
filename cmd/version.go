/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Show the version number and build information for ami-util.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("ami-util version %s\n", Version) //nolint:forbidigo
		fmt.Printf("commit: %s\n", CommitSHA)        //nolint:forbidigo
		fmt.Printf("build time: %s\n", BuildTime)    //nolint:forbidigo
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
