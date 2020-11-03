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
    "time"
)

func ProcessCreatedDI(issueDB, diDB *sql.DB, repo, sig string, startTime, endTime time.Time) error {
    di, err := getCreatedDI(issueDB, repo, sig, startTime, endTime)
    if err != nil {
        return err
    }
    err = storeIntervalDI(diDB, "CREATED_DI", repo, sig, []IntervalDI{{StartTime: startTime, EndTime: endTime, Value: di}})
    return err
}

func ProcessClosedDI(issueDB, diDB *sql.DB, repo, sig string, startTime, endTime time.Time) error {
    di, err := getClosedDI(issueDB, repo, sig, startTime, endTime)
    if err != nil {
        return err
    }
    err = storeIntervalDI(diDB, "CLOSED_DI", repo, sig, []IntervalDI{{StartTime: startTime, EndTime: endTime, Value: di}})
    return err
}

func ProcessDI(issueDB, diDB *sql.DB, repo, sig string, time time.Time) error {
    di, err := getDI(issueDB, repo, sig, time)
    if err != nil {
        return err
    }
    err = storeInstantDI(diDB, "DI", repo, sig, []InstantDI{{Time: time, Value: di}})
    return err
}

func ProcessCreatedDIs(issueDB, diDB *sql.DB, repo, sig string, startTime, endTime time.Time, frequency time.Duration) error {
    dis := make([]IntervalDI, 0)

    for startTime.Before(endTime) {
        endTime := startTime.Add(frequency)

        di, err := getCreatedDI(issueDB, repo, sig, startTime, endTime)
        if err != nil {
            return err
        }

        dis = append(dis, IntervalDI{
            StartTime: startTime,
            EndTime:   endTime,
            Value:     di,
        })

        startTime = endTime
    }

    err := storeIntervalDI(diDB, "CREATED_DI", repo, sig, dis)

    return err
}

func ProcessClosedDIs(issueDB, diDB *sql.DB, repo, sig string, startTime, endTime time.Time, frequency time.Duration) error {
    dis := make([]IntervalDI, 0)

    for startTime.Before(endTime) {
        endTime := startTime.Add(frequency)

        di, err := getClosedDI(issueDB, repo, sig, startTime, endTime)
        if err != nil {
            return err
        }

        dis = append(dis, IntervalDI{
            StartTime: startTime,
            EndTime:   endTime,
            Value:     di,
        })

        startTime = endTime
    }

    err := storeIntervalDI(diDB, "CLOSED_DI", repo, sig, dis)

    return err
}

func ProcessDIs(issueDB, diDB *sql.DB, repo, sig string, startTime, endTime time.Time, frequency time.Duration) error {
    dis := make([]InstantDI, 0)
    insDI, err := getDI(issueDB, repo, sig, startTime)
    if err != nil {
        return err
    }
    dis = append(dis, InstantDI{
        Time:  startTime,
        Value: insDI,
    })

    for startTime.Before(endTime) {

        createdDI, err := getCreatedDI(issueDB, repo, sig, startTime, startTime.Add(frequency))
        if err != nil {
            return err
        }

        closedDI, err := getClosedDI(issueDB, repo, sig, startTime, startTime.Add(frequency))
        if err != nil {
            return err
        }

        insDI += createdDI - closedDI

        dis = append(dis, InstantDI{
            Time:  startTime.Add(frequency),
            Value: insDI,
        })

        startTime = startTime.Add(frequency)
    }

    err = storeInstantDI(diDB, "DI", repo, sig, dis)

    return err
}
