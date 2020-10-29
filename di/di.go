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
    "context"
    "database/sql"
    "errors"
    "log"
    "strings"
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

// getLabels returns all labels of an issue, saved in map.
func getLabels(db *sql.DB, issue Issue) (map[string][]string, error) {
    if db == nil {
        return nil, errors.New("db is nil")
    }

    labels := make(map[string][]string)

    ctx, cancel := context.WithTimeout(context.Background(), mysqlQueryTimeout)
    defer cancel()
    rows, err := db.QueryContext(ctx, `SELECT NAME 
                                        	  FROM LABEL_ISSUE_RELATIONSHIP, LABEL 
                                        	  WHERE LABEL_ISSUE_RELATIONSHIP.ISSUE_ID = ? 
                                       		  AND LABEL_ISSUE_RELATIONSHIP.LABEL_ID = LABEL.ID`, issue.ID)

    if err != nil {
        return nil, err
    }

    for rows.Next() {
        var label string
        err := rows.Scan(&label)

        if err != nil {
            return nil, err
        }
        parts := strings.Split(label, "/")
        switch len(parts) {
        case 1:
            labels[parts[0]] = append(labels[parts[0]], "")
        case 2:
            labels[parts[0]] = append(labels[parts[0]], parts[1])
        default:
            log.Printf("Issue %v has unsupported label %s", issue.Number, label)
        }
    }

    return labels, nil
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

// GetCreatedDIBetweenTime returns total DI of issues created between startTime and endTime
func GetCreatedDIBetweenTime(db *sql.DB, startTime, endTime time.Time) (float64, error) {
    if db == nil {
        return 0, errors.New("db is nil")
    }

    if startTime.After(endTime) {
        return 0, errors.New("startTime > endTime")
    }

    ctx, cancel := context.WithTimeout(context.Background(), mysqlQueryTimeout)
    defer cancel()
    rows, err := db.QueryContext(ctx, "SELECT ID, NUMBER FROM ISSUE WHERE CREATED_AT BETWEEN ? AND ?", startTime, endTime)
    if err != nil {
        return 0, err
    }

    issues, err := parseIssues(rows)
    if err != nil {
        return 0, err
    }

    for i, _ := range issues {
        issues[i].Label, err = getLabels(db, issues[i])
        if err != nil {
            return 0, err
        }
    }

    di := calculateDI(issues)

    return di, nil
}

// GetClosedDIBetweenTime returns total DI of issues closed between startTime and endTime
func GetClosedDIBetweenTime(db *sql.DB, startTime, endTime time.Time) (float64, error) {
    if db == nil {
        return 0, errors.New("db is nil")
    }

    if startTime.After(endTime) {
        return 0, errors.New("startTime > endTime")
    }

    ctx, cancel := context.WithTimeout(context.Background(), mysqlQueryTimeout)
    defer cancel()
    rows, err := db.QueryContext(ctx, "SELECT ID, NUMBER FROM ISSUE WHERE CLOSED_AT BETWEEN ? AND ?", startTime, endTime)
    if err != nil {
        return 0, err
    }

    issues, err := parseIssues(rows)
    if err != nil {
        return 0, err
    }

    for i, _ := range issues {
        issues[i].Label, err = getLabels(db, issues[i])
        if err != nil {
            return 0, err
        }
    }

    di := calculateDI(issues)

    return di, nil
}

