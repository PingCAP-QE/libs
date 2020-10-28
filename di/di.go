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
	"fmt"
	"log"
	"strings"
	"time"
)

// DI constants
const (
	CRITICAL_DI = 10.0
	MAJOR_DI    = 3.0
	MODERATE_DI = 1.0
	MINOR_DI    = 0.1
)

// MySQL config
const (
	MYSQL_QUERY_TIMEOUT = 5 * time.Second
	MYSQL_LIFE_TIME     = 5 * time.Minute
)

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

// OpenDB opens a mysql database by dsn, returns a handler
func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	db.SetConnMaxLifetime(MYSQL_LIFE_TIME)
	return db, err
}

// calculateDi returns total DI of specified issues
func calculateDi(issues []Issue) float64 {
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
			di += CRITICAL_DI
		case "major":
			di += MAJOR_DI
		case "moderate":
			di += MODERATE_DI
		case "minor":
			di += MINOR_DI
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

	ctx, cancel := context.WithTimeout(context.Background(), MYSQL_QUERY_TIMEOUT)
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

// GetCreatedDiBetweenTime returns total DI of issues created between startTime and endTime
func GetCreatedDiBetweenTime(db *sql.DB, startTime, endTime time.Time) (float64, error) {
	if db == nil {
		return 0, errors.New("db is nil")
	}

	if startTime.After(endTime) {
		return 0, errors.New("startTime > endTime")
	}

	issues := make([]Issue, 0)

	ctx, cancel := context.WithTimeout(context.Background(), MYSQL_QUERY_TIMEOUT)
	defer cancel()
	rows, err := db.QueryContext(ctx, "SELECT ID, NUMBER FROM ISSUE WHERE CLOSED = 0 AND CREATED_AT BETWEEN ? AND ?", startTime, endTime)

	if err != nil {
		return 0, err
	}

	for rows.Next() {
		var issue Issue
		err := rows.Scan(&issue.ID, &issue.Number)
		if err != nil {
			return 0, err
		}
		issue.Label, err = getLabels(db, issue)
		if err != nil {
			return 0, err
		}
		issues = append(issues, issue)
	}

	fmt.Println(issues)

	di := calculateDi(issues)

	return di, nil
}
