package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/mhristof/git-reviewers/git"
	"github.com/mhristof/git-reviewers/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var version = "devel"

var rootCmd = &cobra.Command{
	Use:   "git-reviewers",
	Short: "Show potential code ownerse for a repo.",
	Long: fmt.Sprintf(heredoc.Doc(`
		Find out people with code changes for files and repositories.

		If a file is passed, then 'git blame' is used as well as any merges
		that touch the file provided.

		If no argument is provided, then all files are checked from the repository

		Cache file: %s
	`), git.CacheLocation()),
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		branch, err := cmd.Flags().GetBool("branch")
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Panic("cannot get 'branch' flag")
		}

		log.WithFields(log.Fields{
			"git.Branch()": git.Branch(),
			"git.Main()":   git.Main(),
			"branch":       branch,
		}).Debug("current branch")

		if branch {
			args = util.Eval(fmt.Sprintf("git diff --name-only %s", git.Main()))
			// empty line at the end of the array
			args = args[0 : len(args)-1]
		}

		if len(args) == 0 {
			args = git.Files()
		}

		g := git.NewFromFiles(args)

		log.WithFields(log.Fields{
			"g": fmt.Sprintf("%+v", g),
		}).Debug("Reviewers")

		fmt.Println(strings.Join(g.Reviewers(), ","))
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
	branch := git.Branch() != git.Main()
	rootCmd.PersistentFlags().BoolP(
		"branch", "b", branch,
		"Detect reviewers for current branch. Enabled when branch name is not 'main'",
	)
	rootCmd.PersistentFlags().StringSliceP(
		"bots", "",
		[]string{"semantic-release-bot@martynus.net"},
		"Bot list definition. Used with --human",
	)
	rootCmd.PersistentFlags().BoolP("human", "H", true, "Show human reviewers only.")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Increase verbosity")
	rootCmd.PersistentFlags().BoolP("username", "u", false, "Show the username instead of the email.")
}

// Execute The main function for the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
