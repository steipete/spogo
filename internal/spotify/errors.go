package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var ErrNoContent = fmt.Errorf("no content")

type APIError struct {
	Status  int
	Message string
	Body    string
}

func (e APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("spotify api error (%d): %s", e.Status, e.Message)
	}
	if e.Status != 0 {
		return fmt.Sprintf("spotify api error (%d)", e.Status)
	}
	return "spotify api error"
}

func apiErrorFromResponse(resp *http.Response) error {
	if resp == nil {
		return APIError{Message: "nil response"}
	}
	body, _ := io.ReadAll(resp.Body)
	payload := struct {
		Error struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		} `json:"error"`
		Message string `json:"message"`
	}{
		Message: resp.Status,
	}
	_ = json.Unmarshal(body, &payload)
	status := resp.StatusCode
	message := payload.Error.Message
	if message == "" {
		message = payload.Message
	}
	return APIError{Status: status, Message: message, Body: string(body)}
}
