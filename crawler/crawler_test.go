package crawler

import (
	"github.com/shurcooL/githubv4"
	"os"
	"reflect"
	"testing"
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
