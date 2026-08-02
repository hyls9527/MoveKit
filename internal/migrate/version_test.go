package migrate

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.2.3", "1.2.3", 0},
		{"1.2.3", "1.2.4", -1},
		{"1.2.3", "1.2", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.0", "1.0.0", 0},
		{"0.2.1", "0.2.10", -1},
		{"bad", "1.0", -1},
	}
	for _, tt := range tests {
		if got := compareVersions(tt.a, tt.b); got != tt.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCheckUpdate_NewerVersionAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v9.9.9","html_url":"https://example.com/releases/tag/v9.9.9"}`))
	}))
	defer srv.Close()

	origURL, origClient := githubReleasesLatestURL, updateCheckClient
	githubReleasesLatestURL = srv.URL
	updateCheckClient = srv.Client()
	defer func() {
		githubReleasesLatestURL = origURL
		updateCheckClient = origClient
	}()

	info, err := CheckUpdate()
	if err != nil {
		t.Fatalf("CheckUpdate: %v", err)
	}
	if !info.HasUpdate {
		t.Error("expected HasUpdate=true for newer release")
	}
	if info.NewVersion != "9.9.9" {
		t.Errorf("NewVersion = %q, want 9.9.9", info.NewVersion)
	}
	if info.DownloadURL == "" {
		t.Error("DownloadURL should not be empty")
	}
}

func TestCheckUpdate_NoNewerVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v0.2.1","html_url":"https://example.com/releases/tag/v0.2.1"}`))
	}))
	defer srv.Close()

	origURL, origClient := githubReleasesLatestURL, updateCheckClient
	githubReleasesLatestURL = srv.URL
	updateCheckClient = srv.Client()
	defer func() {
		githubReleasesLatestURL = origURL
		updateCheckClient = origClient
	}()

	info, err := CheckUpdate()
	if err != nil {
		t.Fatalf("CheckUpdate: %v", err)
	}
	if info.HasUpdate {
		t.Errorf("expected HasUpdate=false for same version, got %+v", info)
	}
}

func TestCheckUpdate_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	origURL, origClient := githubReleasesLatestURL, updateCheckClient
	githubReleasesLatestURL = srv.URL
	updateCheckClient = srv.Client()
	defer func() {
		githubReleasesLatestURL = origURL
		updateCheckClient = origClient
	}()

	if _, err := CheckUpdate(); err == nil {
		t.Error("expected error for server error response")
	}
}
