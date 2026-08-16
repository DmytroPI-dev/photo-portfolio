package appconfig

import "testing"

func TestMediaBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "empty is allowed locally", value: "", want: ""},
		{name: "normalizes trailing slash", value: "https://media.example.test/", want: "https://media.example.test"},
		{name: "rejects insecure scheme", value: "http://media.example.test", wantErr: true},
		{name: "rejects a path", value: "https://media.example.test/derivatives", wantErr: true},
		{name: "rejects a query", value: "https://media.example.test?preview=true", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("GALLERY_MEDIA_BASE_URL", test.value)
			got, err := MediaBaseURL()
			if test.wantErr {
				if err == nil {
					t.Fatal("MediaBaseURL returned nil error")
				}
				return
			}
			if err != nil {
				t.Fatalf("MediaBaseURL returned error: %v", err)
			}
			if got != test.want {
				t.Fatalf("MediaBaseURL = %q, want %q", got, test.want)
			}
		})
	}
}
