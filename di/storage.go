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
)

func insertIntervalDI(tx *sql.Tx, table string, repo string, di IntervalDI) error {
    _, err := tx.Exec(`INSERT INTO `+table+`(REPO, START_TIME, END_TIME, DI) VALUES(?, ?, ?, ?)`, repo, di.StartTime, di.EndTime, di.Value)
    return err
}

func insertInstantDI(tx *sql.Tx, table string, repo string, di InstantDI) error {
    _, err := tx.Exec(`INSERT INTO `+table+`(REPO, TIME, DI) VALUES(?, ?, ?)`, repo, di.Time, di.Value)
    return err
}

func storeIntervalDI(db *sql.DB, table string, repo string, dis []IntervalDI) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    for _, di := range dis {
        if err := insertIntervalDI(tx, table, repo, di); err != nil {
            return err
        }
    }
    tx.Commit()
    return nil
}

func storeInstantDI(db *sql.DB, table string, repo string, dis []InstantDI) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    for _, di := range dis {
        if err := insertInstantDI(tx, table, repo, di); err != nil {
            return err
        }
    }
    tx.Commit()
    return nil
}
