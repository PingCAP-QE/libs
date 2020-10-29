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

var db *sql.DB

func init() {
	dsn := os.Getenv("GITHUB_DSN")
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)
}

func TestGetLabels(t *testing.T) {
	issue := Issue{ID: 1}
	labels, err := getLabels(db, issue)
	if err != nil {
		t.Fatal(err)
	}
	if len(labels) != 1 {
		t.Fatal(labels)
	}
}

func TestCalculateDi(t *testing.T) {
	minorIssue := Issue{Label: map[string][]string{"severity": {"minor"}}}
	moderateIssue := Issue{Label: map[string][]string{"severity": {"moderate"}}}
	majorIssue := Issue{Label: map[string][]string{"severity": {"major"}}}
	criticalIssue := Issue{Label: map[string][]string{"severity": {"critical"}}}
	badIssue := Issue{Label: map[string][]string{"severity": {"unknown"}}}

	if di := calculateDI([]Issue{minorIssue, moderateIssue, majorIssue, criticalIssue}); di != minorDI+moderateDI+majorDI+criticalDI {
		t.Fatal(di)
	}

	if di := calculateDI([]Issue{badIssue, criticalIssue}); di != criticalDI {
		t.Fatal(di)
	}

	if di := calculateDI([]Issue{badIssue}); di != 0 {
		t.Fatal(di)
	}

}

func TestGetCreatedDiBetweenTime(t *testing.T) {
	startTime := time.Date(2015, 10, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2015, 11, 1, 0, 0, 0, 0, time.UTC)
	di, err := GetCreatedDIBetweenTime(db, startTime, endTime)
	if err != nil {
		t.Fatal(err)
	}
	if di != 0.1 {
		t.Fatal(di)
	}

	startTime = time.Date(2020, 10, 27, 0, 0, 0, 0, time.UTC)
	endTime = time.Date(2020, 10, 26, 0, 0, 0, 0, time.UTC)
	di, err = GetCreatedDIBetweenTime(db, startTime, endTime)
	if err == nil {
		t.Fatal(err)
	}
}
func TestGetClosedDiBetweenTime(t *testing.T) {
	startTime := time.Date(2020, 10, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2020, 11, 1, 0, 0, 0, 0, time.UTC)
	di, err := GetClosedDIBetweenTime(db, startTime, endTime)
	if err != nil {
		t.Fatal(err)
	}
	if di != 139.1 {
		t.Fatal(di)
	}

	startTime = time.Date(2020, 10, 27, 0, 0, 0, 0, time.UTC)
	endTime = time.Date(2020, 10, 26, 0, 0, 0, 0, time.UTC)
	di, err = GetClosedDIBetweenTime(db, startTime, endTime)
	if err == nil {
		t.Fatal(err)
	}
}

