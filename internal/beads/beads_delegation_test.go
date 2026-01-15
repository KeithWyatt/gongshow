package beads

import (
	"encoding/json"
	"testing"
)

func TestDelegationStructFields(t *testing.T) {
	t.Run("basic delegation", func(t *testing.T) {
		d := Delegation{
			Parent:      "go-parent",
			Child:       "go-child",
			DelegatedBy: "gongshow/crew/marge",
			DelegatedTo: "gongshow/polecats/Toast",
		}

		if d.Parent != "go-parent" {
			t.Errorf("Parent = %q, want %q", d.Parent, "go-parent")
		}
		if d.Child != "go-child" {
			t.Errorf("Child = %q, want %q", d.Child, "go-child")
		}
		if d.DelegatedBy != "gongshow/crew/marge" {
			t.Errorf("DelegatedBy = %q, want %q", d.DelegatedBy, "gongshow/crew/marge")
		}
		if d.DelegatedTo != "gongshow/polecats/Toast" {
			t.Errorf("DelegatedTo = %q, want %q", d.DelegatedTo, "gongshow/polecats/Toast")
		}
	})

	t.Run("delegation with terms", func(t *testing.T) {
		d := Delegation{
			Parent:      "go-parent",
			Child:       "go-child",
			DelegatedBy: "mayor",
			DelegatedTo: "gongshow/crew/joe",
			Terms: &DelegationTerms{
				Portion:            "authentication module",
				Deadline:           "2024-01-20",
				AcceptanceCriteria: "Tests pass, code reviewed",
				CreditShare:        80,
			},
		}

		if d.Terms == nil {
			t.Fatal("Terms should not be nil")
		}
		if d.Terms.Portion != "authentication module" {
			t.Errorf("Portion = %q, want %q", d.Terms.Portion, "authentication module")
		}
		if d.Terms.Deadline != "2024-01-20" {
			t.Errorf("Deadline = %q, want %q", d.Terms.Deadline, "2024-01-20")
		}
		if d.Terms.AcceptanceCriteria != "Tests pass, code reviewed" {
			t.Errorf("AcceptanceCriteria = %q, want %q", d.Terms.AcceptanceCriteria, "Tests pass, code reviewed")
		}
		if d.Terms.CreditShare != 80 {
			t.Errorf("CreditShare = %d, want %d", d.Terms.CreditShare, 80)
		}
	})
}

func TestDelegationJSONMarshal(t *testing.T) {
	t.Run("basic delegation", func(t *testing.T) {
		d := Delegation{
			Parent:      "go-parent",
			Child:       "go-child",
			DelegatedBy: "gongshow/crew/marge",
			DelegatedTo: "gongshow/polecats/Toast",
			CreatedAt:   "2024-01-15T10:00:00Z",
		}

		data, err := json.Marshal(d)
		if err != nil {
			t.Fatalf("json.Marshal error: %v", err)
		}

		var decoded Delegation
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("json.Unmarshal error: %v", err)
		}

		if decoded.Parent != d.Parent {
			t.Errorf("Parent = %q, want %q", decoded.Parent, d.Parent)
		}
		if decoded.Child != d.Child {
			t.Errorf("Child = %q, want %q", decoded.Child, d.Child)
		}
		if decoded.DelegatedBy != d.DelegatedBy {
			t.Errorf("DelegatedBy = %q, want %q", decoded.DelegatedBy, d.DelegatedBy)
		}
		if decoded.DelegatedTo != d.DelegatedTo {
			t.Errorf("DelegatedTo = %q, want %q", decoded.DelegatedTo, d.DelegatedTo)
		}
		if decoded.CreatedAt != d.CreatedAt {
			t.Errorf("CreatedAt = %q, want %q", decoded.CreatedAt, d.CreatedAt)
		}
	})

	t.Run("delegation with terms", func(t *testing.T) {
		d := Delegation{
			Parent:      "go-parent",
			Child:       "go-child",
			DelegatedBy: "mayor",
			DelegatedTo: "gongshow/crew/joe",
			Terms: &DelegationTerms{
				Portion:            "auth module",
				Deadline:           "2024-01-20",
				AcceptanceCriteria: "Tests pass",
				CreditShare:        75,
			},
		}

		data, err := json.Marshal(d)
		if err != nil {
			t.Fatalf("json.Marshal error: %v", err)
		}

		var decoded Delegation
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("json.Unmarshal error: %v", err)
		}

		if decoded.Terms == nil {
			t.Fatal("Terms should not be nil")
		}
		if decoded.Terms.Portion != d.Terms.Portion {
			t.Errorf("Portion = %q, want %q", decoded.Terms.Portion, d.Terms.Portion)
		}
		if decoded.Terms.CreditShare != d.Terms.CreditShare {
			t.Errorf("CreditShare = %d, want %d", decoded.Terms.CreditShare, d.Terms.CreditShare)
		}
	})

	t.Run("terms omitted when nil", func(t *testing.T) {
		d := Delegation{
			Parent:      "go-parent",
			Child:       "go-child",
			DelegatedBy: "mayor",
			DelegatedTo: "crew",
			Terms:       nil,
		}

		data, err := json.Marshal(d)
		if err != nil {
			t.Fatalf("json.Marshal error: %v", err)
		}

		// Check that "terms" is not in the JSON (omitempty)
		jsonStr := string(data)
		// The key should not be present when terms is nil
		var decoded map[string]interface{}
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("json.Unmarshal error: %v", err)
		}

		if _, exists := decoded["terms"]; exists {
			t.Errorf("terms should be omitted when nil, got: %s", jsonStr)
		}
	})
}

func TestDelegationTermsStruct(t *testing.T) {
	t.Run("full terms", func(t *testing.T) {
		terms := DelegationTerms{
			Portion:            "frontend components",
			Deadline:           "2024-02-01",
			AcceptanceCriteria: "All tests pass, reviewed by senior",
			CreditShare:        90,
		}

		if terms.Portion != "frontend components" {
			t.Errorf("Portion = %q, want %q", terms.Portion, "frontend components")
		}
		if terms.Deadline != "2024-02-01" {
			t.Errorf("Deadline = %q, want %q", terms.Deadline, "2024-02-01")
		}
		if terms.AcceptanceCriteria != "All tests pass, reviewed by senior" {
			t.Errorf("AcceptanceCriteria = %q, want %q", terms.AcceptanceCriteria, "All tests pass, reviewed by senior")
		}
		if terms.CreditShare != 90 {
			t.Errorf("CreditShare = %d, want %d", terms.CreditShare, 90)
		}
	})

	t.Run("partial terms", func(t *testing.T) {
		terms := DelegationTerms{
			CreditShare: 50,
		}

		if terms.Portion != "" {
			t.Errorf("Portion should be empty, got %q", terms.Portion)
		}
		if terms.CreditShare != 50 {
			t.Errorf("CreditShare = %d, want %d", terms.CreditShare, 50)
		}
	})

	t.Run("credit share boundaries", func(t *testing.T) {
		tests := []struct {
			name        string
			creditShare int
		}{
			{"zero credit", 0},
			{"full credit", 100},
			{"partial credit", 50},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				terms := DelegationTerms{CreditShare: tc.creditShare}

				data, err := json.Marshal(terms)
				if err != nil {
					t.Fatalf("json.Marshal error: %v", err)
				}

				var decoded DelegationTerms
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal error: %v", err)
				}

				if decoded.CreditShare != tc.creditShare {
					t.Errorf("CreditShare = %d, want %d", decoded.CreditShare, tc.creditShare)
				}
			})
		}
	})
}

func TestDelegationTermsJSONOmitempty(t *testing.T) {
	// Empty terms should have minimal JSON representation
	terms := DelegationTerms{}

	data, err := json.Marshal(terms)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	// All string fields with omitempty should be absent
	if _, exists := decoded["portion"]; exists {
		t.Error("portion should be omitted when empty")
	}
	if _, exists := decoded["deadline"]; exists {
		t.Error("deadline should be omitted when empty")
	}
	if _, exists := decoded["acceptance_criteria"]; exists {
		t.Error("acceptance_criteria should be omitted when empty")
	}
	// CreditShare with omitempty will be absent when 0
	if _, exists := decoded["credit_share"]; exists {
		t.Error("credit_share should be omitted when 0")
	}
}

func TestDelegationRoundTrip(t *testing.T) {
	original := Delegation{
		Parent:      "go-epic-123",
		Child:       "go-task-456",
		DelegatedBy: "hop://town/mayor",
		DelegatedTo: "hop://town/gongshow/crew/marge",
		CreatedAt:   "2024-01-15T10:30:00Z",
		Terms: &DelegationTerms{
			Portion:            "database migration",
			Deadline:           "2024-01-25",
			AcceptanceCriteria: "Migration runs without errors, rollback tested",
			CreditShare:        85,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var decoded Delegation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	// Compare all fields
	if decoded.Parent != original.Parent {
		t.Errorf("Parent mismatch: got %q, want %q", decoded.Parent, original.Parent)
	}
	if decoded.Child != original.Child {
		t.Errorf("Child mismatch: got %q, want %q", decoded.Child, original.Child)
	}
	if decoded.DelegatedBy != original.DelegatedBy {
		t.Errorf("DelegatedBy mismatch: got %q, want %q", decoded.DelegatedBy, original.DelegatedBy)
	}
	if decoded.DelegatedTo != original.DelegatedTo {
		t.Errorf("DelegatedTo mismatch: got %q, want %q", decoded.DelegatedTo, original.DelegatedTo)
	}
	if decoded.CreatedAt != original.CreatedAt {
		t.Errorf("CreatedAt mismatch: got %q, want %q", decoded.CreatedAt, original.CreatedAt)
	}

	if decoded.Terms == nil {
		t.Fatal("Terms should not be nil")
	}
	if decoded.Terms.Portion != original.Terms.Portion {
		t.Errorf("Terms.Portion mismatch: got %q, want %q", decoded.Terms.Portion, original.Terms.Portion)
	}
	if decoded.Terms.Deadline != original.Terms.Deadline {
		t.Errorf("Terms.Deadline mismatch: got %q, want %q", decoded.Terms.Deadline, original.Terms.Deadline)
	}
	if decoded.Terms.AcceptanceCriteria != original.Terms.AcceptanceCriteria {
		t.Errorf("Terms.AcceptanceCriteria mismatch: got %q, want %q", decoded.Terms.AcceptanceCriteria, original.Terms.AcceptanceCriteria)
	}
	if decoded.Terms.CreditShare != original.Terms.CreditShare {
		t.Errorf("Terms.CreditShare mismatch: got %d, want %d", decoded.Terms.CreditShare, original.Terms.CreditShare)
	}
}
