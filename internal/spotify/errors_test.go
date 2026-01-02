package spotify

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAPIErrorFromResponse(t *testing.T) {
	body := io.NopCloser(strings.NewReader(`{"error":{"status":401,"message":"bad"}}`))
	resp := &http.Response{StatusCode: 401, Status: "401", Body: body}
	err := apiErrorFromResponse(resp)
	apiErr, ok := err.(APIError)
	if !ok {
		t.Fatalf("expected APIError")
	}
	if apiErr.Status != 401 || apiErr.Message != "bad" {
		t.Fatalf("unexpected: %#v", apiErr)
	}
}

func TestAPIErrorFromResponseNil(t *testing.T) {
	err := apiErrorFromResponse(nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestAPIErrorError(t *testing.T) {
	err := APIError{Status: 400, Message: "bad"}
	if err.Error() == "" {
		t.Fatalf("expected error string")
	}
	err = APIError{Status: 400}
	if err.Error() == "" {
		t.Fatalf("expected error string")
	}
	err = APIError{}
	if err.Error() == "" {
		t.Fatalf("expected error string")
	}
}
