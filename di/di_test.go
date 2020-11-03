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
    "testing"
    "time"
)

func must(t *testing.T, value interface{}, expected interface{}, name string) {
    if expected != value {
        t.Fatalf("%v = %v, expected %v", name, value, expected)
    }
}

func TestProcessCreatedDI(t *testing.T) {
    startTime := time.Date(2020, 9, 7, 0, 0, 0, 0, time.UTC)
    endTime := time.Date(2020, 9, 14, 0, 0, 0, 0, time.UTC)
    err := ProcessCreatedDI(issueDB, diDB, "tidb", "", startTime, endTime)
    must(t, err, nil, "err")
}

func TestProcessClosedDI(t *testing.T) {
    startTime := time.Date(2020, 9, 7, 0, 0, 0, 0, time.UTC)
    endTime := time.Date(2020, 9, 14, 0, 0, 0, 0, time.UTC)
    err := ProcessClosedDI(issueDB, diDB, "tidb", "", startTime, endTime)
    must(t, err, nil, "err")
}

func TestProcessDI(t *testing.T) {
    time := time.Date(2020, 9, 14, 0, 0, 0, 0, time.UTC)
    err := ProcessDI(issueDB, diDB, "tidb", "", time)
    must(t, err, nil, "err")
}

func TestProcessCreatedDIs(t *testing.T) {
    startTime := time.Date(2019, 12, 30, 0, 0, 0, 0, time.UTC)
    endTime := time.Now()
    repos := []string{"tidb", "tikv", "pd", "dm", "br", "ticdc", "tidb-lightning"}
    for _, repo := range repos {
        err := ProcessCreatedDIs(issueDB, diDB, repo, "", startTime, endTime, 7*24*time.Hour)
        must(t, err, nil, "err")
    }
}

func TestProcessClosedDIs(t *testing.T) {
    startTime := time.Date(2019, 12, 30, 0, 0, 0, 0, time.UTC)
    endTime := time.Now()
    repos := []string{"tidb", "tikv", "pd", "dm", "br", "ticdc", "tidb-lightning"}
    for _, repo := range repos {
        err := ProcessClosedDIs(issueDB, diDB, repo, "", startTime, endTime, 7*24*time.Hour)
        must(t, err, nil, "err")
    }
}

func TestProcessDIs(t *testing.T) {
    startTime := time.Date(2019, 12, 30, 0, 0, 0, 0, time.UTC)
    endTime := time.Now()
    repos := []string{"tidb", "tikv", "pd", "dm", "br", "ticdc", "tidb-lightning"}
    for _, repo := range repos {
        err := ProcessDIs(issueDB, diDB, repo, "", startTime, endTime, 7*24*time.Hour)
        must(t, err, nil, "err")
    }
}
