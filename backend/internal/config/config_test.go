package config

import "testing"

func TestNormalizeDatabaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "plain postgres scheme",
			raw:  "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
			want: "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
		},
		{
			name: "plain postgresql scheme",
			raw:  "postgresql://user:pass@localhost:5432/dbname?sslmode=require",
			want: "postgresql://user:pass@localhost:5432/dbname?sslmode=require",
		},
		{
			name: "pasted psql command with single quotes",
			raw:  "psql 'postgresql://user:pass@host/db?sslmode=require&channel_binding=require'",
			want: "postgresql://user:pass@host/db?sslmode=require&channel_binding=require",
		},
		{
			name: "pasted psql command with options",
			raw:  "psql --set=sslmode=require \"postgres://user:pass@host:5432/db\"",
			want: "postgres://user:pass@host:5432/db",
		},
		{
			name: "trailing semicolon",
			raw:  "postgresql://user:pass@host/db?sslmode=require;",
			want: "postgresql://user:pass@host/db?sslmode=require",
		},
		{
			name: "empty value",
			raw:  "   ",
			want: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeDatabaseURL(tc.raw)
			if got != tc.want {
				t.Fatalf("normalizeDatabaseURL(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}
