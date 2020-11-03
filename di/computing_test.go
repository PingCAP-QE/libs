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

package di

import (
    "database/sql"
    "log"
    "os"
    "testing"
    "time"

    _ "github.com/go-sql-driver/mysql"
)

var issueDB *sql.DB

func init() {
    dsn := os.Getenv("GITHUB_DSN")
    var err error
    issueDB, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }
    issueDB.SetConnMaxLifetime(10 * time.Minute)
}

func TestGetLabels(t *testing.T) {
    issue := Issue{ID: 1}
    labels, err := getLabels(issueDB, issue)
    must(t, err, nil, "err")
    must(t, len(labels), 1, "len(labels)")
}

func TestCalculateDi(t *testing.T) {
    minorIssue := Issue{Label: map[string][]string{"severity": {"minor"}}}
    moderateIssue := Issue{Label: map[string][]string{"severity": {"moderate"}}}
    majorIssue := Issue{Label: map[string][]string{"severity": {"major"}}}
    criticalIssue := Issue{Label: map[string][]string{"severity": {"critical"}}}
    badIssue := Issue{Label: map[string][]string{"severity": {"unknown"}}}

    di := calculateDI([]Issue{minorIssue, moderateIssue, majorIssue, criticalIssue})
    must(t, di, minorDI+moderateDI+majorDI+criticalDI, "di")

    di = calculateDI([]Issue{badIssue, criticalIssue})
    must(t, di, criticalDI, "di")

    di = calculateDI([]Issue{badIssue})
    must(t, di, 0.0, "di")

}
