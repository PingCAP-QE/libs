package crawler

import (
	"context"
	"github.com/google/martian/log"
	"reflect"

	"github.com/shurcooL/githubv4"
)

type PageInfo struct {
	EndCursor   githubv4.String
	HasNextPage bool
}

type Query interface {
	GetPageInfo() PageInfo
}

// FetchAllQueries just travel all the query among pages.
// You must input a Query pointer just like it used in fetchIssuesByLabelsStates : FetchAllQueries(client,&query,variables),
// because query of client need a pointer query input.
func FetchAllQueries(client ClientV4, q Query, variables map[string]interface{}) ([]Query, error) {
	var queryList []Query
	for {
		err := client.QueryWithClients(context.Background(), q, variables)
		if err != nil {
			log.Errorf("Fail to fetch query %v, because: %v", reflect.TypeOf(q), err)
			return nil, err
		}
		queryList = append(queryList, q)
		if !q.GetPageInfo().HasNextPage {
			break
		}
		variables["commentsCursor"] = githubv4.NewString(q.GetPageInfo().EndCursor)
	}
	return queryList, nil
}
