package shared

import "time"

// UserID is a unique identifier for a user.
type UserID string

// ProjectID is a unique identifier for a project.
type ProjectID int64

// EnvironmentID is a unique identifier for an environment.
type EnvironmentID int64

// EnvironmentVariableID is a unique identifier for an environment variable.
type EnvironmentVariableID int64

// Role represents a user's permission level within a project.
type Role string

const (
	// RoleOwner has full control over the project, can manage sharing and settings
	RoleOwner Role = "owner"
	// RoleEditor can create, read, update, and delete resources within the project
	RoleEditor Role = "editor"
	// RoleViewer can only read resources within the project
	RoleViewer Role = "viewer"
)

// Int64ToProjectID converts an int64 to ProjectID.
func Int64ToProjectID(id int64) ProjectID {
	return ProjectID(id)
}

// ProjectIDToInt64 converts a ProjectID to int64.
func ProjectIDToInt64(id ProjectID) int64 {
	return int64(id)
}

// Int64ToEnvironmentID converts an int64 to EnvironmentID.
func Int64ToEnvironmentID(id int64) EnvironmentID {
	return EnvironmentID(id)
}

// EnvironmentIDToInt64 converts an EnvironmentID to int64.
func EnvironmentIDToInt64(id EnvironmentID) int64 {
	return int64(id)
}

// Int64ToEnvironmentVariableID converts an int64 to EnvironmentVariableID.
func Int64ToEnvironmentVariableID(id int64) EnvironmentVariableID {
	return EnvironmentVariableID(id)
}

// EnvironmentVariableIDToInt64 converts an EnvironmentVariableID to int64.
func EnvironmentVariableIDToInt64(id EnvironmentVariableID) int64 {
	return int64(id)
}

// StringToUserID converts a string to UserID.
func StringToUserID(id string) UserID {
	return UserID(id)
}

// UserIDToString converts a UserID to string.
func UserIDToString(id UserID) string {
	return string(id)
}

// Timestamp is a wrapper around time.Time that marshals/unmarshals as RFC3339 strings.
// This ensures consistent JSON serialization between the CLI client and server.
type Timestamp time.Time

// UnmarshalJSON implements json.Unmarshaler for Timestamp.
// It accepts RFC3339 formatted strings.
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	str := string(data)

	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	if str == "" {
		return nil
	}

	parsed, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return err
	}
	*t = Timestamp(parsed)
	return nil
}

// MarshalJSON implements json.Marshaler for Timestamp.
// It outputs RFC3339 formatted strings.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	if time.Time(t).IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + time.Time(t).Format(time.RFC3339) + `"`), nil
}

// ToTime converts a Timestamp to a time.Time.
func (t Timestamp) ToTime() time.Time {
	return time.Time(t)
}

// FromTime creates a Timestamp from a time.Time.
func FromTime(t time.Time) Timestamp {
	return Timestamp(t)
}

// String returns the RFC3339 representation of the Timestamp.
func (t Timestamp) String() string {
	return time.Time(t).Format(time.RFC3339)
}
