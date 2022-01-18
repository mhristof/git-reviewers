package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mhristof/git-reviewers/util"
	log "github.com/sirupsen/logrus"
)

func curl(url string) []byte {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("PRIVATE-TOKEN", os.Getenv("GITLAB_TOKEN"))

	client := &http.Client{}

	log.WithFields(log.Fields{
		"req": req,
	}).Debug("curl")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return body
}

type ApprovalRulesResp []struct {
	ApprovalsRequired    int64 `json:"approvals_required"`
	ContainsHiddenGroups bool  `json:"contains_hidden_groups"`
	EligibleApprovers    []struct {
		AvatarURL string `json:"avatar_url"`
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		State     string `json:"state"`
		Username  string `json:"username"`
		WebURL    string `json:"web_url"`
	} `json:"eligible_approvers"`
	Groups            []interface{} `json:"groups"`
	ID                int64         `json:"id"`
	Name              string        `json:"name"`
	ProtectedBranches []struct {
		AllowForcePush            bool  `json:"allow_force_push"`
		CodeOwnerApprovalRequired bool  `json:"code_owner_approval_required"`
		ID                        int64 `json:"id"`
		MergeAccessLevels         []struct {
			AccessLevel            int64       `json:"access_level"`
			AccessLevelDescription string      `json:"access_level_description"`
			GroupID                interface{} `json:"group_id"`
			UserID                 interface{} `json:"user_id"`
		} `json:"merge_access_levels"`
		Name             string `json:"name"`
		PushAccessLevels []struct {
			AccessLevel            int64       `json:"access_level"`
			AccessLevelDescription string      `json:"access_level_description"`
			GroupID                interface{} `json:"group_id"`
			UserID                 int64       `json:"user_id"`
		} `json:"push_access_levels"`
		UnprotectAccessLevels []interface{} `json:"unprotect_access_levels"`
	} `json:"protected_branches"`
	RuleType string `json:"rule_type"`
	Users    []struct {
		AvatarURL string `json:"avatar_url"`
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		State     string `json:"state"`
		Username  string `json:"username"`
		WebURL    string `json:"web_url"`
	} `json:"users"`
}

func Remote() string {
	url := strings.TrimSuffix(util.Eval("git config --get remote.origin.url")[0], "\n")

	return url
}

func Project() string {
	remote := Remote()

	if !strings.HasPrefix(remote, "git@gitlab.com") {
		log.WithFields(log.Fields{
			"remote": remote,
		}).Warning("cannot handle remote")

		return ""
	}

	project := strings.TrimSuffix(
		strings.Split(remote, ":")[1],
		".git",
	)

	return project
}

func EligibleApprovers() (map[string]struct{}, int64) {
	var resp ApprovalRulesResp

	data := curl(fmt.Sprintf(
		"https://gitlab.com/api/v4/projects/%s/approval_rules",
		url.QueryEscape(Project()),
	))

	err := json.Unmarshal(data, &resp)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("cannot decode gitlab response")

		return map[string]struct{}{}, 0
	}

	users := map[string]struct{}{}

	for _, item := range resp {
		for _, user := range item.EligibleApprovers {
			users[user.Username] = struct{}{}
		}
	}

	return users, resp[0].ApprovalsRequired
}
