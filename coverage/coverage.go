package coverage

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// ProcessCoverage gets the coverage of {owner}/{repo} after each commit in the past year through codecov's API, saves into mysql
func ProcessCoverage(owner, repo, token string) {
	log.Printf("Processing %s\n", owner+"/"+repo)
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://codecov.io/api/gh/"+owner+"/"+repo+"/branch/master/graphs/commits.json?method=min&agg=day&time=365d&inc=totals&order=asc", strings.NewReader(""))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "token "+token)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var message interface{}

	json.Unmarshal(body, &message)

	db, err := sql.Open("mysql", os.Getenv("GITHUB_DSN"))
	if err != nil {
		panic(err)
	}
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	commits := message.(map[string]interface{})["commits"]
	for _, commit := range commits.([]interface{}) {
		timestamp := commit.(map[string]interface{})["timestamp"]
		totals := commit.(map[string]interface{})["totals"]
		coverage, err := strconv.ParseFloat(totals.(map[string]interface{})["c"].(string), 64)
		if err != nil {
			panic(err)
		}
		t, err := time.Parse("2006-01-02 15:04:05", timestamp.(string))
		if err != nil {
			panic(err)
		}
		_, err = tx.Exec("INSERT INTO coverage_timeline(repo, time, coverage) VALUES(?, ?, ?)", repo, t, coverage)
		if err != nil {
			panic(err)
		}
	}

	tx.Commit()

	log.Printf("Finish %s\n", owner+"/"+repo)
}
