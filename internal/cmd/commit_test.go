package cmd

import "testing"

func TestIdentityToEmail(t *testing.T) {
	tests := []struct {
		name     string
		identity string
		domain   string
		want     string
	}{
		{
			name:     "crew member",
			identity: "gongshow/crew/jack",
			domain:   "gongshow.local",
			want:     "gongshow.crew.jack@gongshow.local",
		},
		{
			name:     "polecat",
			identity: "gongshow/polecats/max",
			domain:   "gongshow.local",
			want:     "gongshow.polecats.max@gongshow.local",
		},
		{
			name:     "witness",
			identity: "gongshow/witness",
			domain:   "gongshow.local",
			want:     "gongshow.witness@gongshow.local",
		},
		{
			name:     "refinery",
			identity: "gongshow/refinery",
			domain:   "gongshow.local",
			want:     "gongshow.refinery@gongshow.local",
		},
		{
			name:     "mayor with trailing slash",
			identity: "mayor/",
			domain:   "gongshow.local",
			want:     "mayor@gongshow.local",
		},
		{
			name:     "deacon with trailing slash",
			identity: "deacon/",
			domain:   "gongshow.local",
			want:     "deacon@gongshow.local",
		},
		{
			name:     "custom domain",
			identity: "myrig/crew/alice",
			domain:   "example.com",
			want:     "myrig.crew.alice@example.com",
		},
		{
			name:     "deeply nested",
			identity: "rig/polecats/nested/deep",
			domain:   "test.io",
			want:     "rig.polecats.nested.deep@test.io",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := identityToEmail(tt.identity, tt.domain)
			if got != tt.want {
				t.Errorf("identityToEmail(%q, %q) = %q, want %q",
					tt.identity, tt.domain, got, tt.want)
			}
		})
	}
}
