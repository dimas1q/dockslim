package registry

import "testing"

func TestBuildRegistryURL(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		endpoint string
		wantURL  string
	}{
		{
			name:     "base no slash v2",
			base:     "https://example.com",
			endpoint: "/v2/",
			wantURL:  "https://example.com/v2/",
		},
		{
			name:     "base slash v2",
			base:     "https://example.com/",
			endpoint: "/v2/",
			wantURL:  "https://example.com/v2/",
		},
		{
			name:     "basepath v2",
			base:     "https://example.com/basepath",
			endpoint: "/v2/",
			wantURL:  "https://example.com/basepath/v2/",
		},
		{
			name:     "basepath slash v2",
			base:     "https://example.com/basepath/",
			endpoint: "/v2/",
			wantURL:  "https://example.com/basepath/v2/",
		},
		{
			name:     "basepath manifest",
			base:     "https://example.com/basepath",
			endpoint: "/v2/library/nginx/manifests/latest",
			wantURL:  "https://example.com/basepath/v2/library/nginx/manifests/latest",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := buildRegistryURL(test.base, test.endpoint)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != test.wantURL {
				t.Fatalf("expected %q, got %q", test.wantURL, got)
			}
		})
	}
}
