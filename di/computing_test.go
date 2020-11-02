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
    "math"
    "os"
    "testing"
    "time"

    _ "github.com/go-sql-driver/mysql"
)

var issueDB *sql.DB

func must(t *testing.T, value interface{}, expected interface{}, name string) {
    if expected != value {
        t.Fatalf("%v = %v, expected %v", name, value, expected)
    }
}

func mustNot(t *testing.T, value interface{}, unexpected interface{}, name string) {
    if unexpected == value {
        t.Fatalf("%v must not be %v", name, unexpected)
    }
}

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

func TestGetCreatedDiBetweenTime(t *testing.T) {
    startTime := time.Date(2015, 10, 1, 0, 0, 0, 0, time.UTC)
    endTime := time.Date(2015, 11, 1, 0, 0, 0, 0, time.UTC)
    di, err := getCreatedDIBetweenTime(issueDB, startTime, endTime)
    must(t, err, nil, "err")
    must(t, di, 0.1, "di")

    startTime = time.Date(2020, 10, 27, 0, 0, 0, 0, time.UTC)
    endTime = time.Date(2020, 10, 26, 0, 0, 0, 0, time.UTC)
    _, err = getCreatedDIBetweenTime(issueDB, startTime, endTime)
    mustNot(t, err, nil, "err")
}

func TestGetClosedDiBetweenTime(t *testing.T) {
    startTime := time.Date(2020, 10, 1, 0, 0, 0, 0, time.UTC)
    endTime := time.Date(2020, 11, 1, 0, 0, 0, 0, time.UTC)
    di, err := getClosedDIBetweenTime(issueDB, startTime, endTime)
    must(t, err, nil, "err")
    must(t, di, 139.1, "di")

    startTime = time.Date(2020, 10, 27, 0, 0, 0, 0, time.UTC)
    endTime = time.Date(2020, 10, 26, 0, 0, 0, 0, time.UTC)
    di, err = getClosedDIBetweenTime(issueDB, startTime, endTime)
    mustNot(t, err, nil, "err")
}

func TestGetCreatedDIsFrom(t *testing.T) {
    ti := time.Date(2020, 9, 21, 0, 0, 0, 0, time.UTC)
    dis, err := getCreatedDIsFrom(issueDB, ti, 7*24*time.Hour)
    must(t, err, nil, "err")
    must(t, len(dis), 7, `len(dis)`)
}

func TestGetClosedDIsFrom(t *testing.T) {
    ti := time.Date(2020, 9, 21, 0, 0, 0, 0, time.UTC)
    dis, err := getClosedDIsFrom(issueDB, ti, 7*24*time.Hour)
    must(t, err, nil, "err")
    must(t, len(dis), 7, `len(dis)`)
}

func TestGetDI(t *testing.T) {
    ti := time.Date(2020, 9, 21, 0, 0, 0, 0, time.UTC)
    di, err := getDI(issueDB, ti)
    must(t, err, nil, "err")
    if math.Abs(di-1229.3) > 1e-6 {
        t.Fatalf("GetDI %v returns %f", ti, di)
    }
}
