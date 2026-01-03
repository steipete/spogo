package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type hashResolver struct {
	client  *http.Client
	session *connectSession

	mu     sync.Mutex
	hashes map[string]string
}

func newHashResolver(client *http.Client, session *connectSession) *hashResolver {
	return &hashResolver{
		client:  client,
		session: session,
		hashes:  map[string]string{},
	}
}

func (h *hashResolver) Hash(ctx context.Context, operation string) (string, error) {
	if operation == "" {
		return "", errors.New("operation required")
	}
	h.mu.Lock()
	if hash, ok := h.hashes[operation]; ok && hash != "" {
		h.mu.Unlock()
		return hash, nil
	}
	h.mu.Unlock()
	if err := h.load(ctx, []string{operation}); err != nil {
		return "", err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	hash := h.hashes[operation]
	if hash == "" {
		return "", fmt.Errorf("hash for %s not found", operation)
	}
	return hash, nil
}

func (h *hashResolver) load(ctx context.Context, ops []string) error {
	h.mu.Lock()
	need := make([]string, 0, len(ops))
	for _, op := range ops {
		if h.hashes[op] == "" {
			need = append(need, op)
		}
	}
	h.mu.Unlock()
	if len(need) == 0 {
		return nil
	}
	html, err := h.fetchWebPlayerHTML(ctx)
	if err != nil {
		return err
	}
	mainJS, err := pickWebPlayerBundle(html)
	if err != nil {
		return err
	}
	bundleBase := bundleBaseURL(mainJS)
	mainBody, err := h.fetchText(ctx, mainJS)
	if err != nil {
		return err
	}
	nameMap, hashMap, err := parseWebpackMaps(mainBody)
	if err != nil {
		return err
	}
	chunks := combineChunkNames(nameMap, hashMap)
	if len(chunks) == 0 {
		return errors.New("no chunks found")
	}
	for _, chunk := range chunks {
		body, err := h.fetchText(ctx, bundleBase+chunk)
		if err != nil {
			continue
		}
		found := findOperationHashes(body, need)
		if len(found) > 0 {
			h.mu.Lock()
			for op, hash := range found {
				if h.hashes[op] == "" {
					h.hashes[op] = hash
				}
			}
			h.mu.Unlock()
			need = filterMissing(need, found)
			if len(need) == 0 {
				return nil
			}
		}
	}
	return fmt.Errorf("missing hashes for %s", strings.Join(need, ", "))
}

func (h *hashResolver) fetchWebPlayerHTML(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://open.spotify.com/", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", defaultUserAgent())
	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", apiErrorFromResponse(resp)
	}
	body, err := readAll(resp)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (h *hashResolver) fetchText(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", defaultUserAgent())
	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", apiErrorFromResponse(resp)
	}
	body, err := readAll(resp)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func pickWebPlayerBundle(html string) (string, error) {
	re := regexp.MustCompile(`<script[^>]+src="([^"]+)"`)
	matches := re.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		src := match[1]
		if strings.HasSuffix(src, ".js") && (strings.Contains(src, "/web-player/") || strings.Contains(src, "/mobile-web-player/")) {
			return src, nil
		}
	}
	return "", errors.New("web player bundle not found")
}

func bundleBaseURL(bundleURL string) string {
	if idx := strings.LastIndex(bundleURL, "/"); idx >= 0 {
		return bundleURL[:idx+1]
	}
	return "https://open.spotifycdn.com/cdn/build/web-player/"
}

func parseWebpackMaps(js string) (map[int]string, map[int]string, error) {
	re := regexp.MustCompile(`\{(?:\d+:"[^"]+",?)+\}`)
	matches := re.FindAllString(js, -1)
	if len(matches) == 0 {
		return nil, nil, errors.New("no maps found")
	}
	type scored struct {
		score float64
		data  map[int]string
	}
	var hashMaps []scored
	var nameMaps []scored
	for _, raw := range matches {
		parsed, err := parseMapLiteral(raw)
		if err != nil || len(parsed) == 0 {
			continue
		}
		hashScore := scoreHashMap(parsed)
		nameScore := scoreNameMap(parsed)
		if hashScore > 0.4 {
			hashMaps = append(hashMaps, scored{score: hashScore, data: parsed})
		}
		if nameScore > 0.4 {
			nameMaps = append(nameMaps, scored{score: nameScore, data: parsed})
		}
	}
	if len(hashMaps) == 0 || len(nameMaps) == 0 {
		return nil, nil, errors.New("no suitable maps found")
	}
	sort.Slice(hashMaps, func(i, j int) bool { return hashMaps[i].score > hashMaps[j].score })
	sort.Slice(nameMaps, func(i, j int) bool { return nameMaps[i].score > nameMaps[j].score })
	return nameMaps[0].data, hashMaps[0].data, nil
}

func parseMapLiteral(raw string) (map[int]string, error) {
	mapped := regexp.MustCompile(`(\d+):`).ReplaceAllString(raw, `"$1":`)
	var temp map[string]string
	if err := json.Unmarshal([]byte(mapped), &temp); err != nil {
		return nil, err
	}
	out := make(map[int]string, len(temp))
	for key, value := range temp {
		num, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		out[num] = value
	}
	return out, nil
}

func scoreHashMap(m map[int]string) float64 {
	if len(m) == 0 {
		return 0
	}
	var hits int
	for _, value := range m {
		if isHex(value) && len(value) >= 6 && len(value) <= 12 {
			hits++
		}
	}
	return float64(hits) / float64(len(m))
}

func scoreNameMap(m map[int]string) float64 {
	if len(m) == 0 {
		return 0
	}
	var hits int
	for _, value := range m {
		if strings.Contains(value, "-") || strings.Contains(value, "/") {
			hits++
		}
	}
	return float64(hits) / float64(len(m))
}

func isHex(value string) bool {
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return value != ""
}

func combineChunkNames(nameMap, hashMap map[int]string) []string {
	keys := make([]int, 0, len(nameMap))
	for key := range nameMap {
		if hashMap[key] != "" {
			keys = append(keys, key)
		}
	}
	sort.Ints(keys)
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		name := nameMap[key]
		hash := hashMap[key]
		if name == "" || hash == "" {
			continue
		}
		out = append(out, fmt.Sprintf("%s.%s.js", name, hash))
	}
	return out
}

func findOperationHashes(body string, ops []string) map[string]string {
	found := map[string]string{}
	for _, op := range ops {
		if op == "" {
			continue
		}
		escaped := regexp.QuoteMeta(op)
		pattern := regexp.MustCompile(`(?s)` + escaped + `.{0,400}?sha256Hash\":\"([a-f0-9]{64})\"`)
		match := pattern.FindStringSubmatch(body)
		if len(match) > 1 {
			found[op] = match[1]
			continue
		}
	}
	return found
}

func filterMissing(need []string, found map[string]string) []string {
	remaining := make([]string, 0, len(need))
	for _, op := range need {
		if found[op] == "" {
			remaining = append(remaining, op)
		}
	}
	return remaining
}
