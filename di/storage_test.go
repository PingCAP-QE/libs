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
)

var diDB *sql.DB

func init() {
    dsn := os.Getenv("DI_DSN")
    var err error
    diDB, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }
    diDB.SetConnMaxLifetime(10 * time.Minute)
}

func clearDB(db *sql.DB) {
    tx, err := db.Begin()
    if err != nil {
        log.Fatal(err)
    }
    if _, err := tx.Exec(`DELETE FROM DI WHERE REPO = 'test'`); err != nil {
        log.Fatal(err)
    }
    if _, err := tx.Exec(`DELETE FROM CREATED_DI WHERE REPO = 'test'`); err != nil {
        log.Fatal(err)
    }
    if _, err := tx.Exec(`DELETE FROM CLOSED_DI WHERE REPO = 'test'`); err != nil {
        log.Fatal(err)
    }
    tx.Commit()
    clearDB(diDB)
}

func TestInsertIntervalDI(t *testing.T) {
    tx, err := diDB.Begin()
    must(t, err, nil, "err")
    err = insertIntervalDI(tx, "CREATED_DI", "test", IntervalDI{
        StartTime: time.Now().AddDate(0, 0, -7),
        EndTime:   time.Now(),
        Value:     10,
    })
    must(t, err, nil, "err")
    tx.Commit()
    clearDB(diDB)
}

func TestInsertInstantDI(t *testing.T) {
    tx, err := diDB.Begin()
    must(t, err, nil, "err")
    err = insertInstantDI(tx, "DI", "test", InstantDI{
        Time:  time.Now(),
        Value: 20,
    })
    must(t, err, nil, "err")
    tx.Commit()
    clearDB(diDB)
}

func TestStoreIntervalDI(t *testing.T) {
    dis := []IntervalDI{{time.Now(), time.Now().AddDate(0, 0, -7), 10},
        {time.Now().AddDate(0, 0, -7), time.Now().AddDate(0, 0, -14), 20}}
    err := storeIntervalDI(diDB, "CREATED_DI", "test", dis)
    must(t, err, nil, "err")
    clearDB(diDB)
}

func TestStoreInstantDI(t *testing.T) {
    dis := []InstantDI{{time.Now(), 10},
        {time.Now().AddDate(0, 0, -7), 20}}
    err := storeInstantDI(diDB, "CREATED_DI", "test", dis)
    must(t, err, nil, "err")
    clearDB(diDB)
}
