package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

type NullInt32 sql.NullInt32

func (i NullInt32) MarshalJSON() ([]byte, error) {
	if i.Valid {
		return json.Marshal(i.Int32)
	}
	return json.Marshal(nil)
}

func (i NullInt32) Value() (driver.Value, error) {
	if !i.Valid {
		return nil, nil
	}
	return int64(i.Int32), nil
}
