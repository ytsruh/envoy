package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerateUUID_BasicFunctionality(t *testing.T) {
	result := GenerateUUID()

	if result == "" {
		t.Error("GenerateUUID() returned empty string")
	}

	if result == "string" {
		t.Error("GenerateUUID() should return a UUID, not literal 'string'")
	}
}

func TestGenerateUUID_ValidFormat(t *testing.T) {
	result := GenerateUUID()

	parsed, err := uuid.Parse(result)
	if err != nil {
		t.Errorf("GenerateUUID() returned invalid UUID format: %v", err)
	}

	if parsed.Version() != 4 {
		t.Errorf("Expected UUID v4, got v%d", parsed.Version())
	}
}

func TestGenerateUUID_Uniqueness(t *testing.T) {
	uuids := make(map[string]bool)
	count := 100

	for i := 0; i < count; i++ {
		id := GenerateUUID()
		if uuids[id] {
			t.Errorf("Duplicate UUID generated: %s", id)
		}
		uuids[id] = true
	}

	if len(uuids) != count {
		t.Errorf("Expected %d unique UUIDs, got %d", count, len(uuids))
	}
}

func TestGenerateUUID_DifferentOnEachCall(t *testing.T) {
	id1 := GenerateUUID()
	id2 := GenerateUUID()

	if id1 == id2 {
		t.Errorf("GenerateUUID() returned same UUID twice: %s", id1)
	}
}
