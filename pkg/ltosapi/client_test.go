package ltosapi

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mustWrite(t *testing.T, w http.ResponseWriter, data []byte) {
	t.Helper()
	if _, err := w.Write(data); err != nil {
		t.Errorf("failed to write mock response: %v", err)
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestTarget(t *testing.T) {
	client, err := NewClient("https://clock.example.com", "", "", false)
	if err != nil {
		t.Errorf("unexpected error calling NewClient()")
	}
	if got := client.Target(); got != "https://clock.example.com" {
		t.Errorf("Target() = %q, want %q", got, "https://clock.example.com")
	}
}

func TestInvalidTarget(t *testing.T) {
	var err error

	_, err = NewClient("", "", "", false)
	if err == nil {
		t.Errorf("expected error, got nil for empty baseURL")
	}

	_, err = NewClient("foobar", "", "", false)
	if err == nil {
		t.Errorf("expected error, got nil for baseURL 'foobar'")
	}
}

func TestNewClient_RejectsCredentialsInTarget(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
	}{
		{"user and password", "https://user:pass@clock.example.com"},
		{"password only", "https://:pass@clock.example.com"},
		{"user only", "https://user@clock.example.com"},
		{"empty userinfo", "https://@clock.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL, "", "", false)
			if err == nil {
				t.Fatalf("expected error for baseURL %q, got nil", tt.baseURL)
			}
			if client != nil {
				t.Error("expected nil client on error")
			}
			if !strings.Contains(err.Error(), "must not contain credentials") {
				t.Errorf("error = %q, want it to mention credentials", err)
			}
			if strings.Contains(err.Error(), tt.baseURL) {
				t.Errorf("error message echoes the target URL and may leak credentials: %q", err)
			}
		})
	}
}

func TestFetchStatus_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/status" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		mustWrite(t, w, []byte(`{
			"system-information": {"hostname": "clock1", "version": "7.10.008"},
			"data": {"rest-api": {"api-version": "20.05.013"}}
		}`))
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, "", "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := client.FetchStatus(ctx, testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.SystemInformation.Hostname != "clock1" {
		t.Errorf("hostname = %q, want %q", status.SystemInformation.Hostname, "clock1")
	}
	if status.Data.RestAPI.Version != "20.05.013" {
		t.Errorf("api version = %q, want %q", status.Data.RestAPI.Version, "20.05.013")
	}
}

func TestFetchStatus_BasicAuth(t *testing.T) {
	var gotUser, gotPass string
	var authPresent bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, authPresent = r.BasicAuth()
		mustWrite(t, w, []byte(`{"system-information": {}, "data": {"rest-api": {}}}`))
	}))
	defer srv.Close()

	t.Run("credentials sent when configured", func(t *testing.T) {
		client, _ := NewClient(srv.URL, "myuser", "mypass", false)
		_, err := client.FetchStatus(context.Background(), testLogger())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !authPresent {
			t.Fatal("expected Basic Auth header to be present")
		}
		if gotUser != "myuser" || gotPass != "mypass" {
			t.Errorf("auth = %q:%q, want myuser:mypass", gotUser, gotPass)
		}
	})

	t.Run("no auth header when credentials empty", func(t *testing.T) {
		client, _ := NewClient(srv.URL, "", "", false)
		_, err := client.FetchStatus(context.Background(), testLogger())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if authPresent {
			t.Error("expected no Basic Auth header when credentials are empty")
		}
	})
}

func TestFetchStatus_Non200Status(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"401 unauthorized", http.StatusUnauthorized},
		{"403 forbidden", http.StatusForbidden},
		{"404 not found", http.StatusNotFound},
		{"500 server error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer srv.Close()

			client, _ := NewClient(srv.URL, "", "", false)
			status, err := client.FetchStatus(context.Background(), testLogger())
			if err == nil {
				t.Fatal("expected error for non-200 status code")
			}
			if status != nil {
				t.Error("expected nil status on error")
			}
		})
	}
}

func TestFetchStatus_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mustWrite(t, w, []byte(`not valid json`))
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, "", "", false)
	_, err := client.FetchStatus(context.Background(), testLogger())
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestFetchStatus_ConnectionRefused(t *testing.T) {
	// Point at a closed server to simulate connection refused
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	client, _ := NewClient(url, "", "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := client.FetchStatus(ctx, testLogger())
	if err == nil {
		t.Fatal("expected error when server is unreachable")
	}
}

func TestFetchStatus_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow device — block until the context is done
		<-r.Context().Done()
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, "", "", false)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.FetchStatus(ctx, testLogger())
	if err == nil {
		t.Fatal("expected error when context is cancelled")
	}
}

func TestNewClient_ClonesDefaultTransport(t *testing.T) {
	client, _ := NewClient("https://clock.example.com", "", "", true)

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T, want *http.Transport", client.httpClient.Transport)
	}

	if transport.Proxy == nil {
		t.Fatal("expected Proxy to be preserved from default transport")
	}
	if transport.DialContext == nil {
		t.Fatal("expected DialContext to be preserved from default transport")
	}
	if !transport.ForceAttemptHTTP2 {
		t.Fatal("expected ForceAttemptHTTP2 to be preserved from default transport")
	}
	if transport.TLSHandshakeTimeout == 0 {
		t.Fatal("expected TLSHandshakeTimeout to be preserved from default transport")
	}
	if transport.TLSClientConfig == nil {
		t.Fatal("expected TLSClientConfig to be set")
	}
	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("expected InsecureSkipVerify=true")
	}
}

func TestNewClient_DisablesInsecureSkipVerifyWhenRequested(t *testing.T) {
	client, _ := NewClient("https://clock.example.com", "", "", false)

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T, want *http.Transport", client.httpClient.Transport)
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("expected TLSClientConfig to be set")
	}
	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("expected InsecureSkipVerify=false")
	}
}
