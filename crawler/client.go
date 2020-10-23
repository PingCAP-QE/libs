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
	"fmt"

	"github.com/google/martian/log"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var clientV4s []*githubv4.Client
var clientsIsOpen = false

// The Client used for client pool
type ClientV4 struct {
	client *githubv4.Client
	index  int
}

// NewGithubV4Client return the first Client in the pool
func NewGithubV4Client() ClientV4 {
	if !clientsIsOpen {
		panic(fmt.Errorf("clients need to be init before use it , you could init it by InitGithubV4Client"))
	}
	return ClientV4{clientV4s[0], 0}
}

// initGithubV4Client init the clients pool.
func InitGithubV4Client(tokens []string) {
	clientV4s = make([]*githubv4.Client, len(tokens))
	for i, token := range tokens {
		src := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpClient := oauth2.NewClient(context.Background(), src)
		clientV4s[i] = githubv4.NewClient(httpClient)
	}
	clientsIsOpen = true
}

// RefreshGithubV4Client just init client pool again
func RefreshGithubV4Client(tokens []string) {
	InitGithubV4Client(tokens)
}

// QueryWithClientsPool package the client pool , you could use it just like client.Query in githubv4 package
func (c ClientV4) QueryWithClientsPool(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	for {
		err := c.client.Query(ctx, q, variables)
		if err != nil {
			if c.index == len(clientV4s)-1 {
				log.Errorf("All tokens has been used, but could not stop the steps of errors.")
				return err
			}
			c.index++
			c.client = clientV4s[c.index]
		} else {
			return nil
		}
	}

}
