package db

import (
	"database/sql"
	"encoding/json"
)

type NullInt32 struct {
	sql.NullInt32
}

func NewNullInt32(_int32 int32, valid bool) NullInt32 {
	return NullInt32{
		sql.NullInt32{
			Int32: _int32,
			Valid: valid,
		},
	}
}

func (i NullInt32) MarshalJSON() ([]byte, error) {
	if i.Valid {
		return json.Marshal(i.Int32)
	}

	return json.Marshal(nil)
}
