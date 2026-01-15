package connection

import (
	"testing"
)

func TestParseAddress(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Address
		wantErr bool
	}{
		{
			name:  "rig/polecat",
			input: "gongshow/rictus",
			want:  &Address{Rig: "gongshow", Polecat: "rictus"},
		},
		{
			name:  "rig/ broadcast",
			input: "gongshow/",
			want:  &Address{Rig: "gongshow"},
		},
		{
			name:  "machine:rig/polecat",
			input: "vm:gongshow/rictus",
			want:  &Address{Machine: "vm", Rig: "gongshow", Polecat: "rictus"},
		},
		{
			name:  "machine:rig/ broadcast",
			input: "vm:gongshow/",
			want:  &Address{Machine: "vm", Rig: "gongshow"},
		},
		{
			name:  "rig only (no slash)",
			input: "gongshow",
			want:  &Address{Rig: "gongshow"},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "empty machine",
			input:   ":gongshow/rictus",
			wantErr: true,
		},
		{
			name:    "empty rig",
			input:   "vm:/rictus",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAddress(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseAddress(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseAddress(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got.Machine != tt.want.Machine {
				t.Errorf("Machine = %q, want %q", got.Machine, tt.want.Machine)
			}
			if got.Rig != tt.want.Rig {
				t.Errorf("Rig = %q, want %q", got.Rig, tt.want.Rig)
			}
			if got.Polecat != tt.want.Polecat {
				t.Errorf("Polecat = %q, want %q", got.Polecat, tt.want.Polecat)
			}
		})
	}
}

func TestAddressString(t *testing.T) {
	tests := []struct {
		addr *Address
		want string
	}{
		{
			addr: &Address{Rig: "gongshow", Polecat: "rictus"},
			want: "gongshow/rictus",
		},
		{
			addr: &Address{Rig: "gongshow"},
			want: "gongshow/",
		},
		{
			addr: &Address{Machine: "vm", Rig: "gongshow", Polecat: "rictus"},
			want: "vm:gongshow/rictus",
		},
		{
			addr: &Address{Machine: "vm", Rig: "gongshow"},
			want: "vm:gongshow/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.addr.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAddressIsLocal(t *testing.T) {
	tests := []struct {
		addr *Address
		want bool
	}{
		{&Address{Rig: "gongshow"}, true},
		{&Address{Machine: "", Rig: "gongshow"}, true},
		{&Address{Machine: "local", Rig: "gongshow"}, true},
		{&Address{Machine: "vm", Rig: "gongshow"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.addr.String(), func(t *testing.T) {
			if got := tt.addr.IsLocal(); got != tt.want {
				t.Errorf("IsLocal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddressIsBroadcast(t *testing.T) {
	tests := []struct {
		addr *Address
		want bool
	}{
		{&Address{Rig: "gongshow"}, true},
		{&Address{Rig: "gongshow", Polecat: ""}, true},
		{&Address{Rig: "gongshow", Polecat: "rictus"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.addr.String(), func(t *testing.T) {
			if got := tt.addr.IsBroadcast(); got != tt.want {
				t.Errorf("IsBroadcast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddressEqual(t *testing.T) {
	tests := []struct {
		a, b *Address
		want bool
	}{
		{
			&Address{Rig: "gongshow", Polecat: "rictus"},
			&Address{Rig: "gongshow", Polecat: "rictus"},
			true,
		},
		{
			&Address{Machine: "", Rig: "gongshow"},
			&Address{Machine: "local", Rig: "gongshow"},
			true,
		},
		{
			&Address{Rig: "gongshow", Polecat: "rictus"},
			&Address{Rig: "gongshow", Polecat: "nux"},
			false,
		},
		{
			&Address{Rig: "gongshow"},
			nil,
			false,
		},
	}

	for _, tt := range tests {
		name := "equal"
		if !tt.want {
			name = "not equal"
		}
		t.Run(name, func(t *testing.T) {
			if got := tt.a.Equal(tt.b); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAddress_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Address
		wantErr bool
	}{
		// Malformed: empty/whitespace variations
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			want:    &Address{Rig: "   "},
			wantErr: false, // whitespace-only rig is technically parsed
		},
		{
			name:    "just slash",
			input:   "/",
			wantErr: true,
		},
		{
			name:    "double slash",
			input:   "//",
			wantErr: true,
		},
		{
			name:    "triple slash",
			input:   "///",
			wantErr: true,
		},

		// Malformed: leading/trailing issues
		{
			name:    "leading slash",
			input:   "/polecat",
			wantErr: true,
		},
		{
			name:    "leading slash with rig",
			input:   "/rig/polecat",
			wantErr: true,
		},
		{
			name:  "trailing slash is broadcast",
			input: "rig/",
			want:  &Address{Rig: "rig"},
		},

		// Machine prefix edge cases
		{
			name:    "colon only",
			input:   ":",
			wantErr: true,
		},
		{
			name:    "colon with trailing slash",
			input:   ":/",
			wantErr: true,
		},
		{
			name:    "empty machine with colon",
			input:   ":rig/polecat",
			wantErr: true,
		},
		{
			name:  "multiple colons in machine",
			input: "host:8080:rig/polecat",
			want:  &Address{Machine: "host", Rig: "8080:rig", Polecat: "polecat"},
		},
		{
			name:  "colon in rig name",
			input: "machine:rig:port/polecat",
			want:  &Address{Machine: "machine", Rig: "rig:port", Polecat: "polecat"},
		},

		// Multiple slash handling (SplitN behavior)
		{
			name:  "extra slashes in polecat",
			input: "rig/pole/cat/extra",
			want:  &Address{Rig: "rig", Polecat: "pole/cat/extra"},
		},
		{
			name:  "many path components",
			input: "a/b/c/d/e",
			want:  &Address{Rig: "a", Polecat: "b/c/d/e"},
		},

		// Unicode handling
		{
			name:  "unicode rig name",
			input: "日本語/polecat",
			want:  &Address{Rig: "日本語", Polecat: "polecat"},
		},
		{
			name:  "unicode polecat name",
			input: "rig/工作者",
			want:  &Address{Rig: "rig", Polecat: "工作者"},
		},
		{
			name:  "emoji in address",
			input: "🔧/🐱",
			want:  &Address{Rig: "🔧", Polecat: "🐱"},
		},
		{
			name:  "unicode machine name",
			input: "マシン:rig/polecat",
			want:  &Address{Machine: "マシン", Rig: "rig", Polecat: "polecat"},
		},

		// Long addresses
		{
			name:  "very long rig name",
			input: "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789/polecat",
			want:  &Address{Rig: "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789", Polecat: "polecat"},
		},
		{
			name:  "very long polecat name",
			input: "rig/abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789",
			want:  &Address{Rig: "rig", Polecat: "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789"},
		},

		// Special characters
		{
			name:  "hyphen in names",
			input: "my-rig/my-polecat",
			want:  &Address{Rig: "my-rig", Polecat: "my-polecat"},
		},
		{
			name:  "underscore in names",
			input: "my_rig/my_polecat",
			want:  &Address{Rig: "my_rig", Polecat: "my_polecat"},
		},
		{
			name:  "dots in names",
			input: "my.rig/my.polecat",
			want:  &Address{Rig: "my.rig", Polecat: "my.polecat"},
		},
		{
			name:  "mixed special chars",
			input: "rig-1_v2.0/polecat-alpha_1.0",
			want:  &Address{Rig: "rig-1_v2.0", Polecat: "polecat-alpha_1.0"},
		},

		// Whitespace in components
		{
			name:  "space in rig name",
			input: "my rig/polecat",
			want:  &Address{Rig: "my rig", Polecat: "polecat"},
		},
		{
			name:  "space in polecat name",
			input: "rig/my polecat",
			want:  &Address{Rig: "rig", Polecat: "my polecat"},
		},
		{
			name:  "leading space in rig",
			input: " rig/polecat",
			want:  &Address{Rig: " rig", Polecat: "polecat"},
		},
		{
			name:  "trailing space in polecat",
			input: "rig/polecat ",
			want:  &Address{Rig: "rig", Polecat: "polecat "},
		},

		// Edge case: machine with no rig after colon
		{
			name:    "machine colon nothing",
			input:   "machine:",
			wantErr: true,
		},
		{
			name:    "machine colon slash",
			input:   "machine:/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAddress(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseAddress(%q) expected error, got %+v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseAddress(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got.Machine != tt.want.Machine {
				t.Errorf("Machine = %q, want %q", got.Machine, tt.want.Machine)
			}
			if got.Rig != tt.want.Rig {
				t.Errorf("Rig = %q, want %q", got.Rig, tt.want.Rig)
			}
			if got.Polecat != tt.want.Polecat {
				t.Errorf("Polecat = %q, want %q", got.Polecat, tt.want.Polecat)
			}
		})
	}
}

func TestMustParseAddress_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseAddress with empty string should panic")
		}
	}()
	MustParseAddress("")
}

func TestMustParseAddress_Valid(t *testing.T) {
	// Should not panic
	addr := MustParseAddress("rig/polecat")
	if addr.Rig != "rig" || addr.Polecat != "polecat" {
		t.Errorf("MustParseAddress returned wrong address: %+v", addr)
	}
}

func TestAddressRigPath(t *testing.T) {
	tests := []struct {
		addr *Address
		want string
	}{
		{
			addr: &Address{Rig: "gongshow", Polecat: "rictus"},
			want: "gongshow/rictus",
		},
		{
			addr: &Address{Rig: "gongshow"},
			want: "gongshow/",
		},
		{
			addr: &Address{Machine: "vm", Rig: "gongshow", Polecat: "rictus"},
			want: "gongshow/rictus",
		},
		{
			addr: &Address{Rig: "a", Polecat: "b/c/d"},
			want: "a/b/c/d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.addr.RigPath()
			if got != tt.want {
				t.Errorf("RigPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
