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
	issueNum := 19929

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

func TestParse(t *testing.T) {
	var tmp = "## Please edit this comment to complete the following information\r\n\r\n### Not a bug\r\n\r\n1. Remove the 'type/bug' label\r\n2. Add notes to indicate why it is not a bug\r\n\r\n### Duplicate bug\r\n\r\n1. Add the 'type/duplicate' label\r\n2. Add the link to the original bug\r\n\r\n### Bug\r\n\r\nNote: Make Sure that 'component', and 'severity' labels are added\r\nExample for how to fill out the template: https://github.com/pingcap/tidb/issues/20100\r\n\r\n#### 1. Root Cause Analysis (RCA) (optional) \r\n<!-- Write down the reason why this bug occurs -->\r\n\r\n#### 2. Symptom (optional)\r\n\r\n<!-- What will the user see when this bug occurs. The error message may be in the terminal, log or monitoring -->\r\n\r\n#### 3. All Trigger Conditions (optional)\r\n\r\n<!-- All the user scenarios that may trigger this bug -->\r\n\r\n#### 4. Workaround (optional)\r\n\r\n#### 5. Affected versions\r\n\r\n<!--\r\nIn the format of [start_version:end_version], multiple version ranges are\r\naccepted. If the bug only affects the unreleased version, please input:\r\n\"unreleased\". For example:\r\n\r\nNotes:\r\n  1. Do not use any white spaces in '[]'.\r\n  2. The range in '[]' is a closed interval\r\n  3. The version format is `v$Major.$Minor.$Patch`, the $Majoy and $Minor\r\n     number in a version range should be the same. [v3.0.1:v3.1.2] is\r\n     invalid because the $Minor number of the version range is different.\r\n\r\nExample 1: [v3.0.1:v3.0.5], [v4.0.1:v4.0.5]\r\nExample 2: unreleased\r\n-->\r\n[v3.0.0:v3.0.19]\r\n#### 6. Fixed versions\r\nv3.0.20\r\n<!--\r\nThe first released version that contains this fix in each minor version. If the bug's affected version has been released, the fixed version should be a detailed version number; If the bug doesn't affect any released version, the fixed version can be \"master\";  \r\n\r\nExample 1: v3.0.13, v4.0.5\r\nExample 2: master\r\n-->"
	_, errMaps := ParseCommentBody(tmp)
	t.Log(errMaps)
}

func TestParseCommentBodyFromIssue(t *testing.T) {
	owner := "pingcap"
	repo := "tidb"
	issueNum := 19929

	comments, _, err := client.Issues.ListComments(context.TODO(), owner, repo, issueNum, nil)
	must(err)

	for _, c := range comments {
		if ContainsBugTemplate(*c.Body) {
			infos, errs := ParseCommentBody(*c.Body)

			v := reflect.ValueOf(*infos)
			for i := 0; i < v.NumField(); i++ {
				t.Log(v.Type().Field(i).Name, ":", v.Field(i))

				t.Log("-------------------------")
			}

			t.Log(errs)
		}
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
