package spotify

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func randomHex(size int) string {
	if size <= 0 {
		return ""
	}
	buf := make([]byte, (size+1)/2)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	out := hex.EncodeToString(buf)
	return out[:size]
}

func encodeJSON(payload any) *strings.Reader {
	data, _ := json.Marshal(payload)
	return strings.NewReader(string(data))
}

func mapPlayOriginID(player map[string]any) string {
	if player == nil {
		return ""
	}
	if origin, ok := player["play_origin"].(map[string]any); ok {
		if id, ok := origin["device_identifier"].(string); ok {
			return id
		}
	}
	return ""
}

func connectVersion(auth connectAuth) string {
	if auth.ConnectVersion != "" {
		return auth.ConnectVersion
	}
	return auth.ClientVersion
}
