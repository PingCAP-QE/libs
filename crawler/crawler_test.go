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
	"testing"

	"github.com/google/go-github/v32/github"
)

var client *github.Client

func init() {
	token := os.Getenv("GITHUB_TOKEN")
	client = NewGithubClient(token)
	InitGithubV4Client([]string{token})
}

func TestFetchIssueWithComments1(t *testing.T) {
	clientV4 := NewGithubV4Client()
	issueWithComments, errs := FetchIssueWithCommentsByLabels(clientV4, "Andrewmatilde", "demo", []string{"bug"})
	if errs != nil {
		panic(errs[0])
	}

	url := *FetchLatestArtifactUrl(client, "Andrewmatilde", "demo")
	byteList := DownloadAndUnzipArtifact(url)
	s := byteList[0]
	var issuesDataExpected []github.Issue
	err := json.Unmarshal(s, &issuesDataExpected)
	if err != nil {
		panic(err)
	}
	if len(issuesDataExpected) != len(*issueWithComments) {
		t.Errorf("issueWithComments size : %d; expected %d", len(*issueWithComments), len(issuesDataExpected))
		return
	}
}
func TestFetchIssueWithComments2(t *testing.T) {
	clientV4 := NewGithubV4Client()
	issueWithComments, errs := FetchIssueWithCommentsByLabels(clientV4, "pingcap", "tidb", []string{"type/bug"})
	if errs != nil {
		fmt.Println(len(errs))
	}
	fmt.Println(issueWithComments)
}
