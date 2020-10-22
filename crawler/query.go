package crawler

import (
	"context"
	"github.com/google/martian/log"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"os"
	"strings"
)

var clientV4s []*githubv4.Client

func init() {
	initGithubV4ClientList()
}

func initGithubV4ClientList() {
	sysTokens := os.Getenv("GITHUB_TOKENS")
	tokens := strings.Split(sysTokens, ":")
	clientV4s = make([]*githubv4.Client, len(tokens))
	for i, token := range tokens {
		clientV4s[i] = NewGithubV4Client(token)
	}
}

// NewGithubV4Client new clientv4 by github tokens.
func NewGithubV4Client(token string) *githubv4.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return githubv4.NewClient(httpClient)
}

type ClientV4 struct {
	client *githubv4.Client
	index  int
}

func NewClientV4() ClientV4 {
	return ClientV4{clientV4s[0], 0}
}

func (c ClientV4) QueryExceedRateLimit(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	for {
		err := c.client.Query(ctx, q, variables)
		if err != nil {
			if c.index == len(clientV4s)-1 {
				log.Errorf("All tokens has been used.")
				return err
			}
			c.index++
			c.client = clientV4s[c.index]
		}
	}

}
