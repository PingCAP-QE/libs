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
	"context"
	"reflect"

	"github.com/google/martian/log"
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
		err := client.QueryWithClientsPool(context.Background(), q, variables)
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
