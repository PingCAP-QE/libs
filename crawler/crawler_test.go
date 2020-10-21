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
	"os"
	"reflect"
	"testing"

	"github.com/shurcooL/githubv4"
)

var Client *githubv4.Client

func init() {
	token := os.Getenv("GITHUB_TOKEN")
	Client = NewGithubV4Client(token)
}

func TestFetchIssueWithComments(t *testing.T) {
	issueWithComments, errs := FetchIssueWithComments(Client, "pingcap", "tidb", []string{"type/bug"})
	if errs != nil {
		panic(errs[0])
	}

	v := reflect.ValueOf(*issueWithComments)
	for i := 0; i < v.NumField(); i++ {
		t.Log(v.Type().Field(i).Name, ":", v.Field(i))
		t.Log("-------------------------")
	}
}
