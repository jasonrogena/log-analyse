package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	uuid "github.com/satori/go.uuid"
)

func Insert(db *sql.DB, table string, idColumn string, columns []string, values []string) (string, error) {
	if len(columns) == len(values) {
		columnsString := idColumn + ", " + strings.Join(columns[:], ", ")
		valueString := "?" + strings.Repeat(", ?", len(values))
		query := "INSERT INTO " + table + " (" + columnsString + ") VALUES (" + valueString + ")"
		fmt.Printf("Query is %q", query)
		stmt, err1 := db.Prepare(query)
		if err1 != nil {
			return "", err1
		}

		insertedUUID, err2 := genUUID()
		if err2 != nil {
			return "", err2
		}
		_, err3 := stmt.Exec(append([]string{insertedUUID}, values...))
		return insertedUUID, err3
	}

	err3 := fmt.Errorf("Column count %d doesn't match value count %d", len(columns), len(values))
	return "", err3
}

func genUUID() (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return u.String(), err
}
