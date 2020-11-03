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

// generateQuery generates query from original query string with repo and sig
// only non-empty repo and sig will be involved
func generateQuery(query, repo, sig string) string {
    if len(repo) > 0 {
        query += ` AND REPOSITORY_ID = (
                                SELECT ID 
                                FROM REPOSITORY
                                WHERE REPO_NAME = '` + repo + `')`
    }
    if len(sig) > 0 {
        query += ` AND ID IN (
                                SELECT ISSUE_ID
                                FROM LABEL_ISSUE_RELATIONSHIP
                                    LEFT JOIN LABEL ON LABEL_ISSUE_RELATIONSHIP.LABEL_ID = LABEL.ID
                                WHERE LABEL.NAME = '` + sig + `')`
    }
    return query
}

// getClosedDI returns DI of issues created between startTime and endTime
// only non-empty repo and sig will be involved
func getCreatedDI(db *sql.DB, repo, sig string, startTime, endTime time.Time) (float64, error) {
    if db == nil {
        return 0, errors.New("db is nil")
    }

    if startTime.After(endTime) {
        return 0, errors.New("startTime > endTime")
    }

    ctx, cancel := context.WithTimeout(context.Background(), mysqlQueryTimeout)
    defer cancel()

    query := generateQuery("SELECT ID, NUMBER FROM ISSUE WHERE CREATED_AT BETWEEN ? AND ?", repo, sig)
    rows, err := db.QueryContext(ctx, query, startTime, endTime)

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

// getClosedDI returns DI of issues closed between startTime and endTime
// only non-empty repo and sig will be involved
func getClosedDI(db *sql.DB, repo, sig string, startTime, endTime time.Time) (float64, error) {
    if db == nil {
        return 0, errors.New("db is nil")
    }

    if startTime.After(endTime) {
        return 0, errors.New("startTime > endTime")
    }

    ctx, cancel := context.WithTimeout(context.Background(), mysqlQueryTimeout)
    defer cancel()

    query := generateQuery("SELECT ID, NUMBER FROM ISSUE WHERE CLOSED_AT BETWEEN ? AND ?", repo, sig)
    rows, err := db.QueryContext(ctx, query, startTime, endTime)

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

// getDI returns DI at a specified time
// only non-empty repo and sig will be involved
func getDI(db *sql.DB, repo, sig string, time time.Time) (float64, error) {
    if db == nil {
        return 0, errors.New("db is nil")
    }

    ctx, cancel := context.WithTimeout(context.Background(), mysqlQueryTimeout)
    defer cancel()

    query := generateQuery("SELECT ID, NUMBER FROM ISSUE WHERE CREATED_AT < ? AND (CLOSED = 0 OR CLOSED_AT > ?)", repo, sig)

    rows, err := db.QueryContext(ctx, query, time, time)
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

// insertIntervalDI inserts an IntervalDI into table (not committed)
func insertIntervalDI(tx *sql.Tx, table string, repo, sig string, di IntervalDI) error {
    _, err := tx.Exec(`INSERT INTO `+table+`(REPO, SIG, START_TIME, END_TIME, DI) VALUES(?, ?, ?, ?, ?)`, repo, sig, di.StartTime, di.EndTime, di.Value)
    return err
}

// insertIntervalDI inserts an InstantDI into table (not committed)
func insertInstantDI(tx *sql.Tx, table string, repo, sig string, di InstantDI) error {
    _, err := tx.Exec(`INSERT INTO `+table+`(REPO, SIG, TIME, DI) VALUES(?, ?, ?, ?)`, repo, sig, di.Time, di.Value)
    return err
}

// storeIntervalDI inserts an array of IntervalDI into table and commits
func storeIntervalDI(db *sql.DB, table string, repo, sig string, dis []IntervalDI) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    for _, di := range dis {
        if err := insertIntervalDI(tx, table, repo, sig, di); err != nil {
            return err
        }
    }
    tx.Commit()
    return nil
}

// storeInstantDI inserts an array of InstantDI into table and commits
func storeInstantDI(db *sql.DB, table string, repo, sig string, dis []InstantDI) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    for _, di := range dis {
        if err := insertInstantDI(tx, table, repo, sig, di); err != nil {
            return err
        }
    }
    tx.Commit()
    return nil
}