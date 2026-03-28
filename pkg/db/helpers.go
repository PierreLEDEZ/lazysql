package db

import (
	"database/sql"
	"fmt"
	"time"
)

func ScanRows(rows *sql.Rows, elapsed time.Duration) (*Result, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := &Result{
		Columns: cols,
		Elapsed: elapsed,
	}

	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		row := make([]interface{}, len(cols))
		for i, v := range values {
			switch val := v.(type) {
			case []byte:
				row[i] = string(val)
			default:
				row[i] = val
			}
		}
		result.Rows = append(result.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result.RowCount = len(result.Rows)
	return result, nil
}

func StringVal(v interface{}) string {
	if v == nil {
		return "<NULL>"
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}
