package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mhristof/git-reviewers/keychain"
	"github.com/mhristof/git-reviewers/util"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

type Git struct {
	Blame             map[string]struct{}
	Merge             map[string]struct{}
	EligibleApprovers map[string]struct{}
	RepoContributors  []string
	ApprovalsRequired int64
}

func New() Git {
	approvers, required := EligibleApprovers()
	return Git{
		Blame:             map[string]struct{}{},
		Merge:             map[string]struct{}{},
		EligibleApprovers: approvers,
		ApprovalsRequired: required,
	}
}

func NewFromFiles(files []string) Git {
	g := New()
	for _, file := range files {
		fromBlame := Blame(file)
		log.WithFields(log.Fields{
			"fromBlame": fromBlame,
			"file":      file,
		}).Debug("git blame reviewers")

		g.Blame = util.Merge(g.Blame, fromBlame)

		fromMerge := Merge(file)
		log.WithFields(log.Fields{
			"fromMerge": fromMerge,
			"file":      file,
		}).Debug("git merge reviewers")

		g.Merge = util.Merge(g.Merge, fromMerge)
	}

	return g
}

func Files() []string {
	files := map[string]struct{}{}
	// var files []string
	for _, line := range util.Eval("git log --name-only --pretty=format:'%N'") {
		if len(line) == 0 {
			continue
		}
		files[line] = struct{}{}
	}

	return util.Keys(files)
}

func RepoReviewers() []string {
	files := map[string]struct{}{}
	// var files []string
	for _, line := range util.Eval("git log --name-only --pretty=format:'%N'") {
		if len(line) == 0 {
			continue
		}
		files[line] = struct{}{}
	}
	fileList := util.Keys(files)

	var authors []string
	for _, file := range fileList {
		fromFile := util.Eval(fmt.Sprintf("git log --pretty=format:'%%ae' -- '%s' --all", file))
		authors = append(authors, fromFile...)
	}

	authors = util.Uniq(authors)

	return authors
}

// FileReviewer Get at list of suitable reviewers for the given file. This
// function will check `git blame` and people that have merged changes in the
// file.
// func FileReviewer(file string) []string {
// 	var authors []string
// 	authors = append(authors, FileCodeOwners(file)...)
// 	authors = append(authors, MergeRequests(file)...)
// 	authors = util.Uniq(authors)

// return authors
// }

// FileCodeOwners Figure out who the code owners are for the given file.
func Blame(file string) map[string]struct{} {
	blame := util.Eval(fmt.Sprintf("git blame --line-porcelain %s", file))

	users := map[string]struct{}{}

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

		user := User(email)

		if user == "" {
			continue
		}

		users[user] = struct{}{}
	}

	return users
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

	return strings.TrimSuffix(path.Base(string(data)), "\n")
}

// MergeRequests Find out users that have merged changes into this `file`.
func Merge(file string) map[string]struct{} {
	commits := util.Eval(fmt.Sprintf(
		"git --no-pager log --pretty=format:'%%H' %s -- %s", Main(), file,
	))

	users := map[string]struct{}{}

	for _, commit := range commits {
		childCommit := child(commit)
		childCommitMessage := util.Eval(fmt.Sprintf("git show %s --pretty=format:'%%s'", childCommit))[0]

		if !strings.HasPrefix(childCommitMessage, "Merge branch") {
			continue
		}

		author := util.Eval(fmt.Sprintf("git show --pretty=format:'%%ae' %s", childCommit))[0]

		log.WithFields(log.Fields{
			"commit":        commit,
			"child(commit)": childCommit,
			"author":        author,
		}).Debug("found author from child")

		users[User(author)] = struct{}{}
	}

	return users
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

	git, err := gitlab.NewClient(keychain.Item("GITLAB_TOKEN"))
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

func (g *Git) Reviewers() []string {
	ret := util.Merge(g.Blame, g.Merge)

	delete(ret, User(Email()))
	delete(g.EligibleApprovers, User(Email()))

	currentApprovers := 0
	for k := range ret {
		if _, ok := g.EligibleApprovers[k]; ok {
			currentApprovers++

			log.WithFields(log.Fields{
				"k":                k,
				"currentApprovers": currentApprovers,
			}).Debug("found eligible approver")
		}
	}

	log.WithFields(log.Fields{
		"currentApprovers": currentApprovers,
		"ret":              ret,
	}).Debug("approvers from Blame and Merge")

	for approver := range g.EligibleApprovers {
		if currentApprovers >= 2*int(g.ApprovalsRequired) {
			break
		}

		ret[approver] = struct{}{}
		currentApprovers++

		log.WithFields(log.Fields{
			"currentApprovers": currentApprovers,
			"approver":         approver,
		}).Debug("added new approver")
	}
	return util.Keys(ret)
}
