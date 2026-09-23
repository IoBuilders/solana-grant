package utils

import "testing"

func TestEnsureTimezoneUTC(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		inputURL string
		wantURL  string
	}{
		{
			name:     "adds timezone when query is missing",
			inputURL: "postgres://user:pass@localhost:5432/dbname",
			wantURL:  "postgres://user:pass@localhost:5432/dbname?timezone=UTC",
		},
		{
			name:     "adds timezone when query exists",
			inputURL: "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
			wantURL:  "postgres://user:pass@localhost:5432/dbname?sslmode=disable&timezone=UTC",
		},
		{
			name:     "keeps UTC timezone value",
			inputURL: "postgres://user:pass@localhost:5432/dbname?sslmode=disable&timezone=UTC",
			wantURL:  "postgres://user:pass@localhost:5432/dbname?sslmode=disable&timezone=UTC",
		},
		{
			name:     "overrides non UTC timezone value",
			inputURL: "postgres://user:pass@localhost:5432/dbname?sslmode=disable&timezone=Europe%2FMadrid",
			wantURL:  "postgres://user:pass@localhost:5432/dbname?sslmode=disable&timezone=UTC",
		},
		{
			name:     "overrides timezone key with different casing",
			inputURL: "postgres://user:pass@localhost:5432/dbname?sslmode=disable&TimeZone=Asia%2FTokyo",
			wantURL:  "postgres://user:pass@localhost:5432/dbname?sslmode=disable&timezone=UTC",
		},
		{
			name:     "deduplicates multiple timezone params",
			inputURL: "postgres://user:pass@localhost:5432/dbname?timezone=Europe%2FMadrid&sslmode=disable&timezone=UTC",
			wantURL:  "postgres://user:pass@localhost:5432/dbname?timezone=UTC&sslmode=disable",
		},
		{
			name:     "handles fragment",
			inputURL: "postgres://user:pass@localhost:5432/dbname?sslmode=disable#ignored",
			wantURL:  "postgres://user:pass@localhost:5432/dbname?sslmode=disable&timezone=UTC#ignored",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotURL := ensureTimezoneUTC(tc.inputURL)
			if gotURL != tc.wantURL {
				t.Fatalf("ensureTimezoneUTC() = %q, want %q", gotURL, tc.wantURL)
			}
		})
	}
}
