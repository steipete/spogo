package spotify

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	totpSecretEnv     = "SPOGO_TOTP_SECRET_URL"
	totpCacheTTL      = 15 * time.Minute
	fallbackTotpVer   = 18
	totpDigits        = 6
	totpStepInSeconds = 30
	totpHTTPTimeout   = 5 * time.Second
)

var totpSecretURLs = []string{
	"https://github.com/xyloflake/spot-secrets-go/blob/main/secrets/secretDict.json?raw=true",
	"https://github.com/Thereallo1026/spotify-secrets/blob/main/secrets/secretDict.json?raw=true",
	"https://code.thetadev.de/ThetaDev/spotify-secrets/raw/branch/main/secrets/secretDict.json",
}

var fallbackTotpSecret = []byte{
	70, 60, 33, 57, 92, 120, 90, 33, 32, 62, 62, 55, 126, 93, 66, 35, 108, 68,
}

type totpCache struct {
	mu      sync.Mutex
	version int
	secret  []byte
	expires time.Time
}

var cachedTotp totpCache

var totpSecretFetcher = fetchTotpSecretHTTP

// SetTotpSecretFetcher overrides the secret fetcher for tests and returns a restore func.
func SetTotpSecretFetcher(fn func(context.Context) (int, []byte, error)) func() {
	prev := totpSecretFetcher
	if fn == nil {
		totpSecretFetcher = fetchTotpSecretHTTP
	} else {
		totpSecretFetcher = fn
	}
	return func() { totpSecretFetcher = prev }
}

func fetchTotpSecretHTTP(ctx context.Context) (int, []byte, error) {
	sources := totpSecretSources()
	var lastErr error
	for _, source := range sources {
		version, secret, err := fetchTotpSecretSource(ctx, source)
		if err == nil && len(secret) > 0 {
			return version, secret, nil
		}
		if err != nil {
			lastErr = err
		}
	}
	if lastErr == nil {
		lastErr = errors.New("totp secrets missing")
	}
	return 0, nil, lastErr
}

func totpSecretSources() []string {
	if override := strings.TrimSpace(os.Getenv(totpSecretEnv)); override != "" {
		return []string{override}
	}
	sources := make([]string, len(totpSecretURLs))
	copy(sources, totpSecretURLs)
	return sources
}

func fetchTotpSecretSource(ctx context.Context, source string) (int, []byte, error) {
	if source == "" {
		return 0, nil, errors.New("totp secret source empty")
	}
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return fetchTotpSecretHTTPURL(ctx, source)
	}
	if strings.HasPrefix(source, "file://") {
		return loadTotpSecretFile(strings.TrimPrefix(source, "file://"))
	}
	return loadTotpSecretFile(source)
}

func fetchTotpSecretHTTPURL(ctx context.Context, source string) (int, []byte, error) {
	client := &http.Client{Timeout: totpHTTPTimeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return 0, nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, nil, fmt.Errorf("totp secrets status %d", resp.StatusCode)
	}
	return parseTotpSecret(resp.Body)
}

func loadTotpSecretFile(path string) (int, []byte, error) {
	if path == "" {
		return 0, nil, errors.New("totp secret path empty")
	}
	file, err := os.Open(path)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = file.Close() }()
	return parseTotpSecret(file)
}

func parseTotpSecret(reader io.Reader) (int, []byte, error) {
	var raw map[string][]int
	if err := json.NewDecoder(reader).Decode(&raw); err != nil {
		return 0, nil, err
	}
	var (
		bestVer    = -1
		bestSecret []int
	)
	for key, value := range raw {
		version, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		if version > bestVer {
			bestVer = version
			bestSecret = value
		}
	}
	if bestVer < 0 || len(bestSecret) == 0 {
		return 0, nil, errors.New("totp secrets missing")
	}
	secret := make([]byte, len(bestSecret))
	for i, value := range bestSecret {
		if value < 0 || value > 255 {
			return 0, nil, errors.New("totp secret out of range")
		}
		secret[i] = byte(value)
	}
	return bestVer, secret, nil
}

func totpSecret(ctx context.Context) (int, []byte) {
	now := time.Now()
	cachedTotp.mu.Lock()
	if now.Before(cachedTotp.expires) && len(cachedTotp.secret) > 0 {
		version := cachedTotp.version
		secret := append([]byte(nil), cachedTotp.secret...)
		cachedTotp.mu.Unlock()
		return version, secret
	}
	cachedTotp.mu.Unlock()

	version, secret, err := totpSecretFetcher(ctx)
	if err != nil || len(secret) == 0 {
		return fallbackTotpVer, append([]byte(nil), fallbackTotpSecret...)
	}

	cachedTotp.mu.Lock()
	cachedTotp.version = version
	cachedTotp.secret = append([]byte(nil), secret...)
	cachedTotp.expires = now.Add(totpCacheTTL)
	cachedTotp.mu.Unlock()

	return version, secret
}

func generateTOTP(ctx context.Context, now time.Time) (string, int, error) {
	version, secret := totpSecret(ctx)
	code, err := totpFromSecret(secret, now)
	return code, version, err
}

func totpFromSecret(secret []byte, now time.Time) (string, error) {
	if len(secret) == 0 {
		return "", errors.New("totp secret empty")
	}
	transformed := make([]byte, len(secret))
	for i, value := range secret {
		transformed[i] = value ^ byte((i%33)+9)
	}
	var joined strings.Builder
	joined.Grow(len(transformed) * 3)
	for _, value := range transformed {
		joined.WriteString(strconv.Itoa(int(value)))
	}
	return totp([]byte(joined.String()), now), nil
}

func totp(key []byte, now time.Time) string {
	counter := uint64(now.Unix() / totpStepInSeconds)
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], counter)
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(msg[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	binCode := (uint32(sum[offset])&0x7f)<<24 |
		(uint32(sum[offset+1])&0xff)<<16 |
		(uint32(sum[offset+2])&0xff)<<8 |
		(uint32(sum[offset+3]) & 0xff)
	code := int(binCode % 1000000)
	return fmt.Sprintf("%0*d", totpDigits, code)
}
