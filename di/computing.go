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
    "time"
)

// DI constants
const (
    criticalDI = 10.0
    majorDI    = 3.0
    moderateDI = 1.0
    minorDI    = 0.1
)

// MySQL config
const mysqlQueryTimeout = 10 * time.Second

// Issue struct
type Issue struct {
    ID           uint
    Number       int
    RepositoryID int
    Closed       bool
    ClosedAt     time.Time
    CreatedAt    time.Time
    Title        string
    Label        map[string][]string
}

// Interval DI struct
type IntervalDI struct {
    StartTime time.Time
    EndTime   time.Time
    Value     float64
}

// Instant DI struct
type InstantDI struct {
    Time  time.Time
    Value float64
}

// calculateDI returns total DI of specified issues
func calculateDI(issues []Issue) float64 {
    di := 0.0
    for _, issue := range issues {
        severity, ok := issue.Label["severity"]
        if !ok {
            log.Printf("Issue %v has no severity", issue.Number)
            continue
        }

        if len(severity) > 1 {
            log.Printf("Issue %v has multiple severities", issue.Number)
            continue
        }

        switch severity[0] {
        case "critical":
            di += criticalDI
        case "major":
            di += majorDI
        case "moderate":
            di += moderateDI
        case "minor":
            di += minorDI
        default:
            log.Printf("Issue %v has unsupported severity %s", issue.Number, severity)
        }
    }
    return di
}

// parseIssues returns issues parsed from sql.Rows
func parseIssues(rows *sql.Rows) ([]Issue, error) {
    issues := make([]Issue, 0)

    for rows.Next() {
        var issue Issue

        err := rows.Scan(&issue.ID, &issue.Number)
        if err != nil {
            return nil, err
        }

        issues = append(issues, issue)
    }

    return issues, nil
}