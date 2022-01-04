package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mhristof/git-reviewers/util"
	log "github.com/sirupsen/logrus"
)

// CodeOwners Figure out who the code owners are for the given file.
func CodeOwners(file string) []string {
	blame := util.Eval(fmt.Sprintf("git blame --line-porcelain %s", file))

	users := map[string]bool{}

	for _, line := range blame {
		if !strings.HasPrefix(line, "author-mail") {
			continue
		}

		fields := strings.Fields(line)

		email := strings.ReplaceAll(fields[1], ">", "")
		email = strings.ReplaceAll(email, "<", "")
		users[email] = true
	}

	return util.Keys(users)
}

func child(commit string) string {
	lines := util.Eval("git rev-list --all --children")
	for _, item := range lines {
		if !strings.HasPrefix(item, commit) {
			continue
		}

		return strings.Split(item, " ")[1]
	}

	return ""
}

// Main Return the main branch of the current git repositorry.
func Main() string {
	ret := "main"

	err := filepath.Walk(".git/refs/heads/",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, "/master") {
				ret = "master"
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	return ret
}

// MergeRequests Find out users that have merged changes into this `file`.
func MergeRequests(file string) []string {
	commits := util.Eval(fmt.Sprintf(
		"git --no-pager log --pretty=format:'%%H' %s -- %s", Main(), file,
	))

	users := map[string]bool{}

	for _, commit := range commits {
		childCommit := child(commit)
		childCommitMessage := util.Eval(fmt.Sprintf("git show %s --pretty=format:'%%s'", childCommit))[0]

		if !strings.HasPrefix(childCommitMessage, "Merge branch") {
			log.WithFields(log.Fields{
				"childCommit":        childCommit,
				"childCommitMessage": childCommitMessage,
			}).Debug("skipping child commit - not a merge")

			continue
		}

		author := util.Eval(fmt.Sprintf("git show --pretty=format:'%%ae' %s", childCommit))[0]

		log.WithFields(log.Fields{
			"commit":        commit,
			"child(commit)": childCommit,
			"author":        author,
		}).Debug("found author from child")

		users[author] = true
	}

	return util.Keys(users)
}

// Email Return the current user git email.
func Email() string {
	return util.Eval("git config user.email")[0]
}
