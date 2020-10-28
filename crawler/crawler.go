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
	"fmt"
	"sync"

	"github.com/google/martian/log"
	"github.com/shurcooL/githubv4"
)

// Issue define issue data fetched from github api v4
type Issue struct {
	DatabaseId githubv4.Int
	Number     githubv4.Int
	Author     struct {
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

	Assignees struct {
		Nodes []struct {
			Login githubv4.String
			Email githubv4.String
		}
	} `graphql:"assignees(first: 100)"`
	Title githubv4.String
	Body  githubv4.String

	TimelineItems struct {
		Nodes []struct {
			Typename             string `graphql:"__typename"`
			CrossReferencedEvent struct {
				Actor struct {
					Login githubv4.String
				}
				CreatedAt githubv4.DateTime
			} `graphql:"... on CrossReferencedEvent"`
		}
	} `graphql:"timelineItems(first: 100)"`
}

// IssueConnection define IssueConnection fetched from github api v4
type IssueConnection struct {
	Nodes    []Issue
	PageInfo struct {
		EndCursor   githubv4.String
		HasNextPage bool
	}
}

type issueQuery struct {
	Repository struct {
		IssueConnection `graphql:"issues(first: 100, after: $commentsCursor, states:$states, filterBy: {labels:$labels,since: $issueDateTime})"`
	} `graphql:"repository(owner: $owner, name: $name)"`
	RateLimit struct {
		Limit     githubv4.Int
		Cost      githubv4.Int
		Remaining githubv4.Int
		ResetAt   githubv4.DateTime
	}
}

func (q issueQuery) GetPageInfo() PageInfo {
	return q.Repository.PageInfo
}

func (q issueQuery) GetQuery() Query {
	return Query(q)
}

// fetchIssuesByLabelsStates fetch issues by labels & states
// More info of issues could be found in https://docs.github.com/en/free-pro-team@latest/graphql/reference/objects#issue
// If there are empty in labels ,you will not get anything.
// TODO: find way to make the input labels work like omitempty.
func fetchIssuesByLabelsStates(client ClientV4,
	owner, name string, labels []string, states []githubv4.IssueState, since githubv4.DateTime) (*[]Issue, error) {
	var query issueQuery

	if len(labels) == 0 {
		log.Errorf("If there are empty in labels ,you will not get anything.")
		err := fmt.Errorf("if there are empty in labels ,you will not get anything from %v/%v", owner, name)
		return nil, err
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
		"issueDateTime":  since,
	}

	queryList, err := FetchAllQueries(client, &query, variables)
	if err != nil {
		log.Errorf(" fetch issue error")
		return nil, err
	}

	var issues []Issue
	for _, query := range queryList {
		issueQueryInstance := query.(issueQuery)
		issues = append(issues, issueQueryInstance.Repository.IssueConnection.Nodes...)
	}

	return &issues, nil
}

// Comment define Comment fetched from github api v4
type Comment struct {
	DatabaseId     githubv4.Int
	Body           string
	ViewerCanReact bool
	Author         struct {
		Login     string
		AvatarURL string `graphql:"avatarUrl(size: 72)"`
	}
}

type commentQuery struct {
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
	RateLimit struct {
		Limit     githubv4.Int
		Cost      githubv4.Int
		Remaining githubv4.Int
		ResetAt   githubv4.DateTime
	}
}

func (q commentQuery) GetPageInfo() PageInfo {
	return q.Repository.Issue.Comments.PageInfo
}

func (q commentQuery) GetQuery() Query {
	return Query(q)
}

// fetchCommentsByIssuesNumbers fetch comments by issues number
// More info of comments could be found in https://docs.github.com/en/free-pro-team@latest/graphql/reference/interfaces#comment
func fetchCommentsByIssuesNumbers(client ClientV4, owner, name string, issueNumber int) (*[]Comment, error) {
	var query commentQuery
	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String(owner),
		"repositoryName":  githubv4.String(name),
		"issueNumber":     githubv4.Int(issueNumber),
		"commentsCursor":  (*githubv4.String)(nil),
	}

	queryList, err := FetchAllQueries(client, &query, variables)
	if err != nil {
		log.Errorf("fetch comments error")
		return nil, err
	}

	var comments []Comment
	for _, query := range queryList {
		commentQueryInstance := query.(commentQuery)
		comments = append(comments, commentQueryInstance.Repository.Issue.Comments.Nodes...)
	}

	return &comments, nil
}

// IssueWithComments define
type IssueWithComments struct {
	Issue
	Comments *[]Comment
}

// FetchIssueWithCommentsByLabels fetch issue combined with comments
// If there are empty in labels ,you will not get anything.
func FetchIssueWithCommentsByLabels(client ClientV4, owner, name string, labels []string, since githubv4.DateTime, count ...int) (*[]IssueWithComments, []error) {
	issues, err := fetchIssuesByLabelsStates(client, owner, name, labels,
		[]githubv4.IssueState{githubv4.IssueStateClosed, githubv4.IssueStateOpen}, since)
	if err != nil {
		return nil, []error{err}
	}
	if issues == nil {
		return nil, nil
	}

	issuesSize := len(*issues)
	if count != nil && count[0] < issuesSize {
		issuesSize = count[0]
	}

	issueWithComments := make([]IssueWithComments, issuesSize)
	for i, issue := range (*issues)[0:issuesSize] {
		issueWithComments[i].Issue = issue
	}

	var mux sync.Mutex
	var errs []error
	wg := sync.WaitGroup{}
	wg.Add(issuesSize)

	for i := range (*issues)[0:issuesSize] {
		go func(index int) {
			defer wg.Done()
			comments, err := fetchCommentsByIssuesNumbers(client, owner, name, int(issueWithComments[index].Number))
			if err != nil {
				mux.Lock()
				errs = append(errs, err)
				mux.Unlock()
			}
			issueWithComments[index].Comments = comments
		}(i)
	}
	wg.Wait()
	if len(errs) > 0 {
		return nil, errs
	}
	return &issueWithComments, nil
}

// The structure of a Query was:
// 1. Define the graphQL data struct of data you want.
//		the rule of the graphQL data struct could be found in https://docs.github.com/en/free-pro-team@latest/graphql
//		and https://github.com/shurcooL/githubv4
// 2. Define variable input to graphQL
// 3. Use FetchAllQueries to get Query data list
// 4. Turn query data list into data struct you want
// 5. Output
// You can read fetchCommentsByIssuesNumbers & fetchIssuesByLabelsStates as examples.
