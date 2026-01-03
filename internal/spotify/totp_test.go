package spotify

import (
	"context"
	"testing"
	"time"
)

func TestTotpRFCVector(t *testing.T) {
	key := []byte("12345678901234567890")
	cases := []struct {
		ts   int64
		want string
	}{
		{59, "287082"},
		{1111111109, "081804"},
		{1111111111, "050471"},
		{1234567890, "005924"},
		{2000000000, "279037"},
		{20000000000, "353130"},
	}
	for _, tc := range cases {
		got := totp(key, time.Unix(tc.ts, 0))
		if got != tc.want {
			t.Fatalf("totp(%d)=%s want %s", tc.ts, got, tc.want)
		}
	}
}

func TestGenerateTOTPUsesFetcher(t *testing.T) {
	cachedTotp = totpCache{}
	restore := SetTotpSecretFetcher(func(ctx context.Context) (int, []byte, error) {
		return 42, []byte{1, 2, 3, 4}, nil
	})
	t.Cleanup(restore)
	now := time.Unix(123456, 0)
	got, version, err := generateTOTP(context.Background(), now)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if version != 42 {
		t.Fatalf("version mismatch")
	}
	if len(got) != 6 {
		t.Fatalf("expected 6-digit code")
	}
}

func TestGenerateTOTPFallback(t *testing.T) {
	cachedTotp = totpCache{}
	restore := SetTotpSecretFetcher(func(ctx context.Context) (int, []byte, error) {
		return 0, nil, context.Canceled
	})
	t.Cleanup(restore)
	got, version, err := generateTOTP(context.Background(), time.Unix(1, 0))
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if version != fallbackTotpVer {
		t.Fatalf("expected fallback version")
	}
	if len(got) != 6 {
		t.Fatalf("expected 6-digit code")
	}
}
