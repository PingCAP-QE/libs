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
			_, fieldErrors := ParseCommentBody(*c.Body)
			for k, errs := range fieldErrors {
				for _, e := range errs {
					t.Logf("%s: %s", k, e.Error())
				}
			}
		}
	}
}
func TestGmail(t *testing.T) {
	test := "## Please edit this comment to complete the following information\n\n### Not a bug\n\n1. Remove the 'type/bug' label\n2. Add notes to indicate why it is not a bug\n\n### Duplicate bug\n\n1. Add the 'type/duplicate' label\n2. Add the link to the original bug\n\n### Bug\n\nNote: Make Sure that 'component', and 'severity' labels are added\nExample for how to fill out the template: https://github.com/pingcap/tidb/issues/20100\n\n#### 1. Root Cause Analysis (RCA)\n<!-- Write down the reason why this bug occurs -->\n\n#### 2. Symptom\n\n<!-- What will the user see when this bug occurs. The error message may be in the terminal, log or monitoring -->\n\n#### 3. All Trigger Conditions\n\n<!-- All the user scenarios that may trigger this bug -->\n\n#### 4. Workaround (optional)\n\n#### 5. Affected versions\n[v4.0.1:v4.1.5]\n<!--\nIn the format of [start_version:end_version], multiple version ranges are\naccepted. If the bug only affects the unreleased version, please input:\n\"unreleased\". For example:\n\nNotes:\n  1. Do not use any white spaces in '[]'.\n  2. The range in '[]' is a closed interval\n  3. The version format is `v$Major.$Minor.$Patch`, the $Majoy and $Minor\n     number in a version range should be the same. [v3.0.1:v3.1.2] is\n     invalid because the $Minor number of the version range is different.\n\nExample 1: [v3.0.1:v3.0.5], [v4.0.1:v4.0.5]\nExample 2: unreleased\n-->\n\n#### 6. Fixed versions\n[v4.0.7]\n<!--\nThe first released version that contains this fix in each minor version. If the bug's affected version has been released, the fixed version should be a detailed version number; If the bug doesn't affect any released version, the fixed version can be \"master\". \n\nExample 1: v3.0.13, v4.0.5\nExample 2: master\n-->"
	_, errMaps := ParseCommentBody(test)
	t.Log(errMaps)
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
			infos, _ = ParseCommentBody(*c.Body)
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

	infos, _ := ParseCommentBody(*comment.Body)

	v := reflect.ValueOf(*infos)
	for i := 0; i < v.NumField(); i++ {
		t.Log(v.Type().Field(i).Name, ":", v.Field(i))

		t.Log("-------------------------")
	}
}
