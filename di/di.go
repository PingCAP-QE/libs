package di

import (
    "database/sql"
    "time"
)

func ProcessCreatedDI(issueDB, diDB *sql.DB, repo, sig string, startTime, endTime time.Time) error {

}

func ProcessClosedDI(issueDB, diDB *sql.DB, repo, sig string, startTime, endTime time.Time) error {

}

func ProcessDI(issueDB, diDB *sql.DB, repo, sig string, time time.Time) error {

}

func ProcessCreatedDIs(issueDB, diDB *sql.DB, repo, sig string, startTime, endTime time.Time, frequency time.Duration) error {

}

func ProcessClosedDIs(issueDB, diDB *sql.DB, repo, sig string, startTime, endTime time.Time, frequency time.Duration) error {

}

func ProcessDIs(issueDB, diDB *sql.DB, repo, sig string, startTime, endTime time.Time, frequency time.Duration) error {

}
