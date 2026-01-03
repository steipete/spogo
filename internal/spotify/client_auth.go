package spotify

import (
	"context"
	"time"
)

func (c *Client) token(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lastToken.AccessToken != "" && time.Until(c.lastToken.ExpiresAt) > time.Minute {
		return c.lastToken.AccessToken, nil
	}
	newToken, err := c.provider.Token(ctx)
	if err != nil {
		return "", err
	}
	c.lastToken = newToken
	return newToken.AccessToken, nil
}

func defaultUserAgent() string {
	return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
}
