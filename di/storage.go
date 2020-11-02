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
