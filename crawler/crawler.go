package crawler

import (
	"context"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"sync"
)

type Issue struct {
	Number githubv4.Int
	Author struct {
		Login     string
		AvatarURL string `graphql:"avatarUrl(size: 72)"`
	}
	Closed    githubv4.Boolean
	ClosedAt  githubv4.DateTime
	CreatedAt githubv4.DateTime
	Labels    struct {
		Nodes []struct {
			Name githubv4.String
		}
	} `graphql:"labels(first: 100)"`
	Title githubv4.String
}

// fetchIssues fetch issues by labels & states
// More info of issues could be found in https://docs.github.com/en/free-pro-team@latest/graphql/reference/objects#issue
func fetchIssues(client *githubv4.Client,
	owner, name string, labels []string, states []githubv4.IssueState) (*[]Issue, error) {
	type IssueConnection struct {
		Nodes    []Issue
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage bool
		}
	}
	var query struct {
		Repository struct {
			IssueConnection `graphql:"issues(first: 100, after: $commentsCursor, states:$states,filterBy: {labels:$labels})"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	labelsV4 := make([]githubv4.String, len(labels))
	for i, label := range labels {
		labelsV4[i] = githubv4.String(label)
	}
	variables := map[string]interface{}{
		"owner":          githubv4.String(owner),
		"name":           githubv4.String(name),
		"labels":         labelsV4,
		"states":         states,
		"commentsCursor": (*githubv4.String)(nil),
	}
	var issues []Issue
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			return nil, err
		}
		issues = append(issues, query.Repository.IssueConnection.Nodes...)
		if !query.Repository.IssueConnection.PageInfo.HasNextPage {
			break
		}
		variables["commentsCursor"] = githubv4.NewString(query.Repository.IssueConnection.PageInfo.EndCursor)
	}
	return &issues, nil
}

type Comment struct {
	Body           string
	ViewerCanReact bool
}

// fetchIssueComments fetch comments by issues number
// More info of comments could be found in https://docs.github.com/en/free-pro-team@latest/graphql/reference/interfaces#comment
func fetchIssueComments(client *githubv4.Client, owner, name string, issueNumber int) (*[]Comment, error) {
	var query struct {
		Repository struct {
			Issue struct {
				Comments struct {
					Nodes    []Comment
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"comments(first: 100, after: $commentsCursor)"`
			} `graphql:"issue(number: $issueNumber)"`
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}
	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String(owner),
		"repositoryName":  githubv4.String(name),
		"issueNumber":     githubv4.Int(issueNumber),
		"commentsCursor":  (*githubv4.String)(nil),
	}
	var allComments []Comment
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			return nil, err
		}
		allComments = append(allComments, query.Repository.Issue.Comments.Nodes...)
		if !query.Repository.Issue.Comments.PageInfo.HasNextPage {
			break
		}
		variables["commentsCursor"] = githubv4.NewString(query.Repository.Issue.Comments.PageInfo.EndCursor)
	}
	return &allComments, nil
}

type IssueWithComments struct {
	Issue
	comments *[]Comment
}

// FetchIssueWithComments fetch issue combined with comments
func FetchIssueWithComments(client *githubv4.Client, owner, name string, labels []string) (*[]IssueWithComments, []error) {
	issues, err := fetchIssues(client, owner, name, labels,
		[]githubv4.IssueState{githubv4.IssueStateClosed, githubv4.IssueStateOpen})
	if err != nil {
		return nil, []error{err}
	}
	issueWithComments := make([]IssueWithComments, len(*issues))
	for i, issue := range *issues {
		issueWithComments[i].Issue = issue
	}
	var mux sync.Mutex
	var errs []error
	wg := sync.WaitGroup{}
	wg.Add(len(*issues))
	for i, _ := range *issues {
		go func(index int) {
			defer wg.Done()
			comments, err := fetchIssueComments(client, owner, name, int(issueWithComments[index].Number))
			if err != nil {
				mux.Lock()
				errs = append(errs, err)
				mux.Unlock()
			}
			issueWithComments[index].comments = comments
		}(i)
	}
	wg.Wait()
	if len(errs) > 0 {
		return nil, errs
	}
	return &issueWithComments, nil
}

// NewGithubV4Client new client by github tokens.
func NewGithubV4Client(token string) *githubv4.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return githubv4.NewClient(httpClient)
}
