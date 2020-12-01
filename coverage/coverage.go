package coverage

import (
    "database/sql"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"
)

// ProcessCoverage gets the coverage of {owner}/{repo} after each commit in the past year through codecov's API, saves into mysql
func ProcessCoverage(owner, repo string) error {
    client := http.Client{}
    req, err := http.NewRequest("GET", "https://codecov.io/api/gh/"+owner+"/"+repo+"/branch/master/graphs/commits.json?method=min&agg=day&time=365d&inc=totals&order=asc", strings.NewReader(""))
    if err != nil {
        return err
    }

    req.Header.Set("Authorization", "token a3a4a9847e4b4e30b1ddc4dd725bf110")

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

    db, err := sql.Open("mysql", os.Getenv("GITHUB_DSN"))
    if err != nil {
        return err
    }
    tx, err := db.Begin()
    if err != nil {
        return err
    }

    commits := message.(map[string]interface{})["commits"]
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
        _, err = tx.Exec("INSERT INTO COVERAGE_TIMELINE(REPO, TIME, COVERAGE) VALUES(?, ?, ?)", "pd", t, coverage)
        if err != nil {
            return err
        }
    }

    tx.Commit()

    return nil
}
