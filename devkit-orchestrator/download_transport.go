package main

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type downloadResult struct {
	Size         int64
	SHA256       string
	SHA512       string
	SHA1         string
	ResolvedURL  string
	Attempts     int
	Resumed      bool
	FallbackUsed bool
}

func devKitUserAgent() string {
	return "MinecraftDevKitOrchestrator/" + version
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

func modrinthCanonicalCDNURL(c Candidate) string {
	if c.ProjectID == "" || c.VersionID == "" || c.Filename == "" {
		return ""
	}
	return "https://cdn.modrinth.com/data/" + url.PathEscape(c.ProjectID) + "/versions/" + url.PathEscape(c.VersionID) + "/" + url.PathEscape(c.Filename)
}

func modrinthMavenURL(c Candidate) string {
	if c.ProjectID == "" || c.VersionID == "" || !strings.EqualFold(filepath.Ext(c.Filename), ".jar") {
		return ""
	}
	project := url.PathEscape(c.ProjectID)
	ver := url.PathEscape(c.VersionID)
	return "https://api.modrinth.com/maven/maven/modrinth/" + project + "/" + ver + "/" + project + "-" + ver + ".jar"
}

func curseForgeCDNURL(fileID, filename string) string {
	id, err := strconv.ParseInt(strings.TrimSpace(fileID), 10, 64)
	if err != nil || id <= 0 || strings.TrimSpace(filename) == "" {
		return ""
	}
	major := id / 1000
	minor := id % 1000
	return fmt.Sprintf("https://edge.forgecdn.net/files/%d/%03d/%s", major, minor, url.PathEscape(filename))
}

func appendUniqueURL(dst []string, raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return dst
	}
	for _, x := range dst {
		if x == raw {
			return dst
		}
	}
	return append(dst, raw)
}

func candidateDownloadURLs(c Candidate) []string {
	out := []string{}
	out = appendUniqueURL(out, c.URL)
	for _, raw := range c.AlternateURLs {
		out = appendUniqueURL(out, raw)
	}
	switch strings.ToLower(c.Provider) {
	case "modrinth":
		out = appendUniqueURL(out, modrinthCanonicalCDNURL(c))
		out = appendUniqueURL(out, modrinthMavenURL(c))
	case "curseforge":
		out = appendUniqueURL(out, curseForgeCDNURL(c.VersionID, c.Filename))
	}
	return out
}

func (e *engine) candidateHeaders(c Candidate) (map[string]string, error) {
	h := map[string]string{
		"User-Agent":      devKitUserAgent(),
		"Accept":          "application/octet-stream,*/*;q=0.8",
		"Accept-Encoding": "identity",
	}
	if strings.EqualFold(c.Provider, "curseforge") {
		key := strings.TrimSpace(e.p.cfKey)
		if key == "" {
			return nil, fmt.Errorf("CurseForge download requires CURSEFORGE_API_KEY (direct CDN authentication has been mandatory since 2026-07-16)")
		}
		h["x-api-key"] = key
	}
	return h, nil
}

func cloneClientWithRedirectHeaders(base *http.Client, headers map[string]string) *http.Client {
	if base == nil {
		base = http.DefaultClient
	}
	cloned := *base
	prior := base.CheckRedirect
	cloned.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		if prior != nil {
			return prior(req, via)
		}
		return nil
	}
	return &cloned
}

func retryDelay(resp *http.Response, attempt int) time.Duration {
	if resp != nil {
		if raw := strings.TrimSpace(resp.Header.Get("Retry-After")); raw != "" {
			if sec, err := strconv.Atoi(raw); err == nil && sec >= 0 && sec <= 60 {
				return time.Duration(sec) * time.Second
			}
			if when, err := http.ParseTime(raw); err == nil {
				d := time.Until(when)
				if d > 0 && d <= time.Minute {
					return d
				}
			}
		}
	}
	return time.Duration(attempt*attempt) * 350 * time.Millisecond
}

func hashFile(path string) (size int64, sha256Hex, sha512Hex, sha1Hex string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, "", "", "", err
	}
	defer f.Close()
	h256 := sha256.New()
	h512 := sha512.New()
	h1 := sha1.New()
	n, err := io.Copy(io.MultiWriter(h256, h512, h1), f)
	if err != nil {
		return 0, "", "", "", err
	}
	return n, hex.EncodeToString(h256.Sum(nil)), hex.EncodeToString(h512.Sum(nil)), hex.EncodeToString(h1.Sum(nil)), nil
}

func verifyCandidateFile(path string, c Candidate) (downloadResult, error) {
	size, h256, h512, h1, err := hashFile(path)
	if err != nil {
		return downloadResult{}, err
	}
	if c.Size > 0 && size != c.Size {
		return downloadResult{}, fmt.Errorf("size mismatch: provider=%d downloaded=%d", c.Size, size)
	}
	if c.SHA256 != "" && !strings.EqualFold(strings.TrimSpace(c.SHA256), h256) {
		return downloadResult{}, fmt.Errorf("sha256 mismatch: provider=%s downloaded=%s", c.SHA256, h256)
	}
	if c.SHA512 != "" && !strings.EqualFold(strings.TrimSpace(c.SHA512), h512) {
		return downloadResult{}, fmt.Errorf("sha512 mismatch: provider=%s downloaded=%s", c.SHA512, h512)
	}
	if c.SHA1 != "" && !strings.EqualFold(strings.TrimSpace(c.SHA1), h1) {
		return downloadResult{}, fmt.Errorf("sha1 mismatch: provider=%s downloaded=%s", c.SHA1, h1)
	}
	return downloadResult{Size: size, SHA256: h256, SHA512: h512, SHA1: h1}, nil
}

func replaceVerifiedFile(part, out string) error {
	if err := os.Chmod(part, 0644); err != nil {
		return err
	}
	if _, err := os.Stat(out); os.IsNotExist(err) {
		return os.Rename(part, out)
	} else if err != nil {
		return err
	}
	backup := out + ".mmv-old"
	_ = os.Remove(backup)
	if err := os.Rename(out, backup); err != nil {
		return err
	}
	if err := os.Rename(part, out); err != nil {
		_ = os.Rename(backup, out)
		return err
	}
	_ = os.Remove(backup)
	return nil
}

func shouldRetryStatus(code int) bool {
	return code == http.StatusRequestTimeout || code == http.StatusTooManyRequests || code >= 500
}

func readErrorBody(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return strings.TrimSpace(string(b))
}

func (e *engine) fetchVerifiedToPath(ctx context.Context, c Candidate, out string) (downloadResult, error) {
	if strings.TrimSpace(out) == "" {
		return downloadResult{}, errors.New("download output path is empty")
	}
	urls := candidateDownloadURLs(c)
	if len(urls) == 0 {
		return downloadResult{}, errors.New("candidate has no download URL")
	}
	headers, err := e.candidateHeaders(c)
	if err != nil {
		return downloadResult{}, err
	}
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return downloadResult{}, err
	}
	part := out + ".mmv-part"
	client := cloneClientWithRedirectHeaders(e.http, headers)
	var errs []string
	totalAttempts := 0
	resumedAny := false

	for urlIndex, raw := range urls {
		if !safeDownloadURL(raw) {
			errs = append(errs, raw+": refused unsafe download URL")
			continue
		}
		for attempt := 1; attempt <= 4; attempt++ {
			totalAttempts++
			if err := ctx.Err(); err != nil {
				return downloadResult{}, err
			}
			var offset int64
			if st, statErr := os.Stat(part); statErr == nil && st.Mode().IsRegular() {
				offset = st.Size()
				if c.Size > 0 && offset > c.Size {
					_ = os.Remove(part)
					offset = 0
				}
			}
			req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
			if reqErr != nil {
				return downloadResult{}, reqErr
			}
			for k, v := range headers {
				req.Header.Set(k, v)
			}
			if offset > 0 {
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
				resumedAny = true
			}
			resp, doErr := client.Do(req)
			if doErr != nil {
				errs = append(errs, fmt.Sprintf("%s attempt %d: %v", raw, attempt, doErr))
				if attempt < 4 {
					time.Sleep(retryDelay(nil, attempt))
					continue
				}
				break
			}

			if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable && c.Size > 0 && offset == c.Size {
				_ = resp.Body.Close()
				vr, verifyErr := verifyCandidateFile(part, c)
				if verifyErr == nil {
					if err := replaceVerifiedFile(part, out); err != nil {
						return downloadResult{}, err
					}
					vr.ResolvedURL = raw
					vr.Attempts = totalAttempts
					vr.Resumed = resumedAny
					vr.FallbackUsed = urlIndex > 0
					return vr, nil
				}
				_ = os.Remove(part)
				errs = append(errs, fmt.Sprintf("%s: completed resume verification failed: %v", raw, verifyErr))
				break
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				body := readErrorBody(resp)
				_ = resp.Body.Close()
				if resp.StatusCode == http.StatusUnauthorized && strings.EqualFold(c.Provider, "curseforge") {
					return downloadResult{}, fmt.Errorf("CurseForge download rejected the API key (HTTP 401). Set a valid CURSEFORGE_API_KEY; direct CDN authentication is required. Provider response: %s", body)
				}
				errs = append(errs, fmt.Sprintf("%s attempt %d: HTTP %d %s", raw, attempt, resp.StatusCode, body))
				if shouldRetryStatus(resp.StatusCode) && attempt < 4 {
					time.Sleep(retryDelay(resp, attempt))
					continue
				}
				break
			}

			flags := os.O_CREATE | os.O_WRONLY
			if offset > 0 && resp.StatusCode == http.StatusPartialContent {
				flags |= os.O_APPEND
			} else {
				offset = 0
				flags |= os.O_TRUNC
			}
			f, openErr := os.OpenFile(part, flags, 0644)
			if openErr != nil {
				_ = resp.Body.Close()
				return downloadResult{}, openErr
			}
			_, copyErr := io.CopyBuffer(f, resp.Body, make([]byte, 1024*1024))
			closeBodyErr := resp.Body.Close()
			syncErr := f.Sync()
			closeErr := f.Close()
			if copyErr != nil || closeBodyErr != nil || syncErr != nil || closeErr != nil {
				firstErr := firstError(copyErr, closeBodyErr, syncErr, closeErr)
				errs = append(errs, fmt.Sprintf("%s attempt %d: stream failed: %v", raw, attempt, firstErr))
				if attempt < 4 {
					time.Sleep(retryDelay(resp, attempt))
					continue
				}
				break
			}

			if st, statErr := os.Stat(part); statErr == nil && c.Size > 0 && st.Size() < c.Size {
				errs = append(errs, fmt.Sprintf("%s attempt %d: short download %d/%d bytes; resuming", raw, attempt, st.Size(), c.Size))
				if attempt < 4 {
					time.Sleep(retryDelay(resp, attempt))
					continue
				}
				break
			}

			vr, verifyErr := verifyCandidateFile(part, c)
			if verifyErr != nil {
				errs = append(errs, fmt.Sprintf("%s attempt %d: verification failed: %v", raw, attempt, verifyErr))
				_ = os.Remove(part)
				break
			}
			if err := replaceVerifiedFile(part, out); err != nil {
				return downloadResult{}, err
			}
			vr.ResolvedURL = raw
			vr.Attempts = totalAttempts
			vr.Resumed = resumedAny
			vr.FallbackUsed = urlIndex > 0
			return vr, nil
		}
		_ = os.Remove(part)
	}
	_ = os.Remove(part)
	return downloadResult{}, fmt.Errorf("all download transports failed: %s", strings.Join(errs, "; "))
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
