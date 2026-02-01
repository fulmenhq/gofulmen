package httpauth

import "testing"

func TestBearerTokenMatches(t *testing.T) {
	expected := "secret-token"

	tests := []struct {
		name   string
		header string
		want   bool
	}{
		{name: "empty header", header: "", want: false},
		{name: "wrong scheme", header: "Basic secret-token", want: false},
		{name: "missing token", header: "Bearer ", want: false},
		{name: "too many fields", header: "Bearer secret-token extra", want: false},
		{name: "tab separated", header: "Bearer\tsecret-token", want: true},
		{name: "wrong token", header: "Bearer nope", want: false},
		{name: "match exact", header: "Bearer secret-token", want: true},
		{name: "match scheme casing", header: "bearer secret-token", want: true},
		{name: "match with whitespace", header: "  Bearer   secret-token  ", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BearerTokenMatches(tt.header, expected); got != tt.want {
				t.Fatalf("BearerTokenMatches()=%v, want %v", got, tt.want)
			}
		})
	}
}
