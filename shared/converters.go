package shared

import (
	"database/sql"
	"database/sql/driver"
)

// NullStringToStringPtr converts a sql.NullString to a *string.
// If the NullString is not valid, returns nil.
func NullStringToStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// StringPtrToNullString converts a *string to a sql.NullString.
// If the pointer is nil, returns an invalid NullString.
func StringPtrToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// NullString is an alternative to sql.NullString that implements
// driver.Valuer and sql.Scanner for better JSON marshaling.
type NullString struct {
	String string
	Valid  bool
}

// Scan implements sql.Scanner for NullString.
func (ns *NullString) Scan(value any) error {
	if value == nil {
		ns.String, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	ns.String = string(value.([]byte))
	return nil
}

// Value implements driver.Valuer for NullString.
func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

// MarshalJSON implements json.Marshaler for NullString.
func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return []byte(`"` + ns.String + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler for NullString.
func (ns *NullString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.String, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	ns.String = string(data[1 : len(data)-1])
	return nil
}
