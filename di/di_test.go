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
	"os"
	"testing"
	"time"
)

var db *sql.DB

func init() {
	var err error
	db, err = OpenDB(os.Getenv("MYSQL_DSN"))
	if err != nil {
		panic(err)
	}
}

func TestGetLabels(t *testing.T) {
	issue := Issue{ID: 11}
	labels, err := getLabels(db, issue)
	if err != nil {
		t.Fatal(err)
	}
	if len(labels) != 3 {
		t.Fatal()
	}

}

func TestCalculateDi(t *testing.T) {
	minorIssue := Issue{Label: map[string][]string{"severity": {"minor"}}}
	criticalIssue := Issue{Label: map[string][]string{"severity": {"critical"}}}
	badIssue := Issue{Label: map[string][]string{"severity": {"unknown"}}}

	if calculateDI([]Issue{minorIssue, criticalIssue}) != MINOR_DI+CRITICAL_DI {
		t.Fatal()
	}

	if calculateDI([]Issue{badIssue, criticalIssue}) != CRITICAL_DI {
		t.Fatal()
	}

	if calculateDI([]Issue{badIssue}) != 0 {
		t.Fatal()
	}

}

func TestGetCreatedDiBetweenTime(t *testing.T) {
	startTime := time.Date(2020, 10, 19, 0, 0, 0, 0, time.Local)
	endTime := time.Date(2020, 10, 26, 0, 0, 0, 0, time.Local)
	di, err := GetCreatedDiBetweenTime(db, startTime, endTime)
	if err != nil {
		t.Fatal()
	}
	if di != 77 {
		t.Fatal()
	}

	startTime = time.Date(2020, 10, 27, 0, 0, 0, 0, time.Local)
	endTime = time.Date(2020, 10, 26, 0, 0, 0, 0, time.Local)
	di, err = GetCreatedDiBetweenTime(db, startTime, endTime)
	if err == nil {
		t.Fatal()
	}
}
