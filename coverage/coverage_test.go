package coverage

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestProcessCoverage(t *testing.T) {
	db, err := sql.Open("mysql", os.Getenv("GITHUB_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	err = ProcessCoverage(db, "pingcap", "tidb-lightning")
	if err != nil {
		t.Fatal(err)
	}
}
