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
	"os"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"
)

var clientv4 *githubv4.Client
var client *github.Client

func init() {
	_ = os.Getenv("GITHUB_TOKEN")
	clientv4 = NewGithubV4Client("151673ee8e5bcbe84f1ba2a158e7b8633e324286")
	client = NewGithubClient("151673ee8e5bcbe84f1ba2a158e7b8633e324286")
}

func TestFetchIssueWithComments(t *testing.T) {
	issueWithComments, errs := FetchIssueWithComments(clientv4, "Andrewmatilde", "demo", []string{"bug"})
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
