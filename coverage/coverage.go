package coverage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// ProcessCoverage gets the coverage of {owner}/{repo} after each commit in the past year through codecov's API, saves into mysql
func ProcessCoverage(db *sql.DB, owner, repo string) error {
	log.Printf("Processing %s\n", owner+"/"+repo)
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://codecov.io/api/gh/"+owner+"/"+repo+"/branch/master/graphs/commits.json?method=min&agg=day&time=365d&inc=totals&order=asc", strings.NewReader(""))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var message interface{}

	json.Unmarshal(body, &message)

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	commits := message.(map[string]interface{})["commits"]
	if commits == nil {
		return fmt.Errorf("cannot find coverage of %v/%v", owner, repo)
	}

	for _, commit := range commits.([]interface{}) {
		timestamp := commit.(map[string]interface{})["timestamp"]
		totals := commit.(map[string]interface{})["totals"]
		coverage, err := strconv.ParseFloat(totals.(map[string]interface{})["c"].(string), 64)
		if err != nil {
			return err
		}
		t, err := time.Parse("2006-01-02 15:04:05", timestamp.(string))
		if err != nil {
			return err
		}
		_, err = tx.Exec("INSERT INTO coverage_timeline(repo_id, time, coverage) SELECT id, ?, ? FROM repository WHERE repo_name = ?", t, coverage, repo)
		if err != nil {
			return err
		}
	}

	tx.Commit()

	log.Printf("Finish %s\n", owner+"/"+repo)

	return nil
}
