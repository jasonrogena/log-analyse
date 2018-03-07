package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	uuid "github.com/satori/go.uuid"
)

func Insert(db *sql.DB, table string, idColumn string, columns []string, values []string, cache bool) (string, error) {
	if len(columns) == len(values) {
		columnsString := idColumn + ", " + strings.Join(columns[:], ", ")
		valueString := "?" + strings.Repeat(", ?", len(values))
		query := "INSERT INTO " + table + " (" + columnsString + ") VALUES (" + valueString + ")"

		insertedUUID, err2 := genUUID()
		if err2 != nil {
			return "", err2
		}
		allValues := append([]string{insertedUUID}, values...)
		allArgs := make([]interface{}, len(values)+1)
		for i := range allValues {
			allArgs[i] = allValues[i]
		}

		if cache {
			return cacheInsert(db, table, query, allArgs, 0)
		}

		return runInsertQuery(db, query, allArgs, 0)
	}

	err3 := fmt.Errorf("Column count %d doesn't match value count %d", len(columns), len(values))
	return "", err3
}

func runInsertQuery(db *sql.DB, query string, args []interface{}, idArg int) (string, error) {
	stmt, err1 := db.Prepare(query)
	defer closeStmt(stmt)

	if err1 != nil {
		return "", err1
	}
	_, err3 := stmt.Exec(args...)
	insertedUUID, ok := args[idArg].(string)
	if ok {
		return insertedUUID, err3
	}

	return "", errors.New("Could not get the inserted UUID")
}

func closeStmt(stmt *sql.Stmt) {
	if stmt != nil {
		stmt.Close()
	}
}

func genUUID() (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return u.String(), err
}
