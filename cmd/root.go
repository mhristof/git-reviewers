package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mhristof/git-reviewers/git"
	"github.com/mhristof/git-reviewers/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var version = "devel"

var rootCmd = &cobra.Command{
	Use:     "git-reviewers",
	Short:   "Find out potential reviewers for PRs.",
	Long:    `Figure out who would be a good reviewer for a change.`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		var authors []string
		authors = append(authors, git.CodeOwners(args[0])...)
		authors = append(authors, git.MergeRequests(args[0])...)

		authors = util.Uniq(authors)

		for i, author := range authors {
			if author == git.Email() {
				authors = append(authors[:i], authors[i+1:]...)

				break
			}
		}

		fmt.Print(strings.Join(authors, ","))
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		Verbose(cmd)
	},
}

// Verbose Increase verbosity.
func Verbose(cmd *cobra.Command) {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		log.Panic(err)
	}

	if verbose {
		log.SetLevel(log.DebugLevel)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Increase verbosity")
}

// Execute The main function for the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
