package db

import (
	"database/sql"
	"encoding/json"
)

type NullInt32 sql.NullInt32

func (i NullInt32) MarshalJSON() ([]byte, error) {
	if i.Valid {
		return json.Marshal(i.Int32)
	}

	return json.Marshal(nil)
}
