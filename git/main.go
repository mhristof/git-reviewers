package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mhristof/git-reviewers/util"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

func RepoReviewers() []string {
	files := map[string]bool{}
	// var files []string
	for _, line := range util.Eval("git log --name-only --pretty=format:'%N'") {
		if len(line) == 0 {
			continue
		}
		files[line] = true
	}
	fileList := util.Keys(files)

	var authors []string
	for _, file := range fileList {
		fromFile := util.Eval(fmt.Sprintf("git log --pretty=format:'%%ae' -- '%s' --all", file))
		authors = append(authors, fromFile...)
	}

	authors = util.Uniq(authors)
	fmt.Println(fmt.Sprintf("authors: %+v", authors))

	return authors
}

// FileReviewer Get at list of suitable reviewers for the given file. This
// function will check `git blame` and people that have merged changes in the
// file.
func FileReviewer(file string) []string {
	var authors []string
	authors = append(authors, FileCodeOwners(file)...)
	authors = append(authors, MergeRequests(file)...)
	authors = util.Uniq(authors)

	return authors
}

// FileCodeOwners Figure out who the code owners are for the given file.
func FileCodeOwners(file string) []string {
	blame := util.Eval(fmt.Sprintf("git blame --line-porcelain %s", file))

	users := map[string]bool{}

	for _, line := range blame {
		if !strings.HasPrefix(line, "author-mail") {
			continue
		}

		if strings.HasSuffix(line, "<not.committed.yet>") {
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

		fields := strings.Fields(item)

		if len(fields) < 2 {
			continue
		}

		return fields[1]
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

func Branch() string {
	data, err := ioutil.ReadFile(".git/HEAD")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Panic("cannot open .git/HEAD")
	}

	return path.Base(string(data))
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

func User(email string) string {
	cache := userCacheLoad()
	defer userCacheDump(cache)

	if username, ok := cache[email]; ok {
		return username
	}

	git, err := gitlab.NewClient(os.Getenv("GITLAB_TOKEN"))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("cannot create gitlab connection")

		return ""
	}

	users, _, err := git.Users.ListUsers(&gitlab.ListUsersOptions{
		Search: &email,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"err":   err,
			"email": email,
		}).Error("cannot find user")

		return ""
	}

	if len(users) != 1 {
		cache[email] = ""
		log.WithFields(log.Fields{
			"len(users)": len(users),
			"email":      email,
		}).Error("cannot find user, im confused.")

		return ""
	}

	cache[email] = users[0].Username

	return users[0].Username
}
