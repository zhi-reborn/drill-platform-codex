package service

import (
	"strings"
	"testing"
)

func TestBuildCASLoginURL(t *testing.T) {
	got, err := BuildCASLoginURL("https://cas.example.com/cas", "http://app.example.com/api/v1/auth/cas/callback?redirect=http%3A%2F%2Fapp.example.com%2Fcas%2Fcallback")
	if err != nil {
		t.Fatalf("BuildCASLoginURL: %v", err)
	}
	if !strings.HasPrefix(got, "https://cas.example.com/cas/login?") {
		t.Fatalf("unexpected login URL: %s", got)
	}
	if !strings.Contains(got, "service=") {
		t.Fatalf("expected service query, got %s", got)
	}
}

func TestParseCASServiceResponseSuccess(t *testing.T) {
	xmlBody := `<?xml version="1.0" encoding="UTF-8"?>
<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
  <cas:authenticationSuccess>
    <cas:user>zhangsan</cas:user>
  </cas:authenticationSuccess>
</cas:serviceResponse>`

	username, err := ParseCASServiceResponse([]byte(xmlBody))
	if err != nil {
		t.Fatalf("ParseCASServiceResponse: %v", err)
	}
	if username != "zhangsan" {
		t.Fatalf("expected zhangsan, got %s", username)
	}
}

func TestParseCASServiceResponseFailure(t *testing.T) {
	xmlBody := `<?xml version="1.0" encoding="UTF-8"?>
<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
  <cas:authenticationFailure code="INVALID_TICKET">ticket not recognized</cas:authenticationFailure>
</cas:serviceResponse>`

	if _, err := ParseCASServiceResponse([]byte(xmlBody)); err == nil {
		t.Fatal("expected CAS failure error")
	}
}
