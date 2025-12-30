package shared

import "time"

// UserID is a unique identifier for a user.
type UserID string

// ProjectID is a unique identifier for a project.
type ProjectID string

// EnvironmentID is a unique identifier for an environment.
type EnvironmentID string

// EnvironmentVariableID is a unique identifier for an environment variable.
type EnvironmentVariableID string

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

// StringToProjectID converts a string to ProjectID.
func StringToProjectID(id string) ProjectID {
	return ProjectID(id)
}

// ProjectIDToString converts a ProjectID to string.
func ProjectIDToString(id ProjectID) string {
	return string(id)
}

// StringToEnvironmentID converts a string to EnvironmentID.
func StringToEnvironmentID(id string) EnvironmentID {
	return EnvironmentID(id)
}

// EnvironmentIDToString converts an EnvironmentID to string.
func EnvironmentIDToString(id EnvironmentID) string {
	return string(id)
}

// StringToEnvironmentVariableID converts a string to EnvironmentVariableID.
func StringToEnvironmentVariableID(id string) EnvironmentVariableID {
	return EnvironmentVariableID(id)
}

// EnvironmentVariableIDToString converts an EnvironmentVariableID to string.
func EnvironmentVariableIDToString(id EnvironmentVariableID) string {
	return string(id)
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
