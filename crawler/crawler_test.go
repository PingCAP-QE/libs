// Copyright 2020 PingCAP-QE libs Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package crawler

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"
)

var client *github.Client

func init() {
	tokenEnvString := os.Getenv("GITHUB_TOKEN")
	tokens := strings.Split(tokenEnvString, ":")
	client = NewGithubClient(tokens[0])
	InitGithubV4Client(tokens)
}

func TestFetchIssueWithCommentsByLabels(t *testing.T) {
	// get issueWithComments by graphQL githubV4 api
	clientV4 := NewGithubV4Client()
	issueWithComments, errs := FetchIssueWithCommentsByLabels(clientV4, "Andrewmatilde", "demo", []string{"bug"}, githubv4.DateTime{})
	if errs != nil {
		panic(errs[0])
	}

	// get issues by artifact
	url := *FetchLatestArtifactUrl(client, "Andrewmatilde", "demo")
	byteList := DownloadAndUnzipArtifact(url)
	s := byteList[0]
	var issuesDataExpected []github.Issue
	err := json.Unmarshal(s, &issuesDataExpected)
	if err != nil {
		panic(err)
	}

	// compare length of issues
	if len(issuesDataExpected) != len(*issueWithComments) {
		t.Errorf("issueWithComments size : %d; expected %d", len(*issueWithComments), len(issuesDataExpected))
		return
	}

	// compare length of comments and find if there are any different issues
	for _, issueWithComment := range *issueWithComments {
		hasIssue := false
		for _, issue := range issuesDataExpected {
			if *issue.Number == int(issueWithComment.Number) {
				hasIssue = true
				if len(*issueWithComment.Comments) != issue.GetComments() {
					t.Errorf("issueWithComment gets different comments size from issue from demo artifact, issue %v.", issue.Title)
				}
			}
		}
		if hasIssue != true {
			t.Errorf("FetchIssueWithCommentsByLabels get different issue from demo artifact.")
		}
	}
}

func TestFetchIssueWithCommentsByLabels_Show(t *testing.T) {
	clientV4 := NewGithubV4Client()
	since := githubv4.DateTime{Time: time.Now().AddDate(0, 0, -10)}
	issueWithComments, errs := FetchIssueWithCommentsByLabels(clientV4, "pingcap", "tidb", []string{"type/bug"}, since)
	if errs != nil {
		fmt.Println(len(errs))
		t.Errorf(errs[0].Error())
	}

	fmt.Println(len(*issueWithComments))

	issueWithComments, errs = FetchIssueWithCommentsByLabels(clientV4, "pingcap", "tidb", []string{}, githubv4.DateTime{}, 10)
	if errs != nil {
		if len(errs) != 1 {
			for _, err := range errs {
				t.Errorf(err.Error())
			}
		} else if errs[0].Error() !=
			fmt.Errorf("if there are empty in labels ,"+
				"you will not get anything from %s/%s", "pingcap", "tidb").Error() {
			t.Error(errs[0])
		}
	}
}
