package extractor

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var client *github.Client

func init() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "GITHUBTOKEN"},
	)

	tc := oauth2.NewClient(ctx, ts)
	client = github.NewClient(tc)

}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func TestValidateCommentBody(t *testing.T) {
	owner := "pingcap"
	repo := "tidb"
	issueNum := 18792

	comments, _, err := client.Issues.ListComments(context.TODO(), owner, repo, issueNum, nil)
	must(err)

	for _, c := range comments {
		if ContainsBugTemplate(*c.Body) {
			fieldErrors := ValidateCommentBody(*c.Body)
			for k, errs := range fieldErrors {
				for _, e := range errs {
					t.Logf("%s: %s", k, e.Error())
				}
			}
		}
	}
}

func TestParseCommentBodyFromIssue(t *testing.T) {
	owner := "pingcap"
	repo := "tidb"
	issueNum := 18792

	comments, _, err := client.Issues.ListComments(context.TODO(), owner, repo, issueNum, nil)
	must(err)

	var infos *BugInfos
	for _, c := range comments {
		if ContainsBugTemplate(*c.Body) {
			infos, err = ParseCommentBody(*c.Body)
			must(err)
		}
	}

	v := reflect.ValueOf(*infos)
	for i := 0; i < v.NumField(); i++ {
		t.Log(v.Type().Field(i).Name, ":", v.Field(i))

		t.Log("-------------------------")
	}
}

func TestParseCommentBody(t *testing.T) {
	owner := "pingcap"
	repo := "tidb"
	var commentId int64 = 689389165

	comment, _, err := client.Issues.GetComment(context.TODO(), owner, repo, commentId)
	must(err)

	infos, err := ParseCommentBody(*comment.Body)
	must(err)

	v := reflect.ValueOf(*infos)
	for i := 0; i < v.NumField(); i++ {
		t.Log(v.Type().Field(i).Name, ":", v.Field(i))

		t.Log("-------------------------")
	}
}
