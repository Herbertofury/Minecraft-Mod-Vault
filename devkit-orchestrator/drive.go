package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type driveClient struct {
	http               *http.Client
	api, upload, token string
}

func newDriveClient(c DriveConfig, h *http.Client) (*driveClient, error) {
	env := first(c.AccessTokenEnv, "MMV_GOOGLE_DRIVE_TOKEN")
	tok := strings.TrimSpace(os.Getenv(env))
	if tok == "" {
		return nil, fmt.Errorf("Google Drive token missing: set %s", env)
	}
	if h == nil {
		h = &http.Client{}
	}
	return &driveClient{http: h, api: first(c.APIBase, "https://www.googleapis.com/drive/v3"), upload: first(c.UploadBase, "https://www.googleapis.com/upload/drive/v3"), token: tok}, nil
}
func (d *driveClient) do(ctx context.Context, method, raw, ctype string, body io.Reader, out any) error {
	req, e := http.NewRequestWithContext(ctx, method, raw, body)
	if e != nil {
		return e
	}
	req.Header.Set("Authorization", "Bearer "+d.token)
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	resp, e := d.http.Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("drive http %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
func (d *driveClient) replace(ctx context.Context, fileID, name string, data []byte) (string, error) {
	var meta struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	raw := d.upload + "/files/" + url.PathEscape(fileID) + "?uploadType=media&fields=id,name"
	if e := d.do(ctx, "PATCH", raw, "application/octet-stream", bytes.NewReader(data), &meta); e != nil {
		return "", e
	}
	if name != "" && name != meta.Name {
		p, _ := json.Marshal(map[string]string{"name": name})
		if e := d.do(ctx, "PATCH", d.api+"/files/"+url.PathEscape(fileID)+"?fields=id,name", "application/json", bytes.NewReader(p), &meta); e != nil {
			return "", e
		}
	}
	return fileID, nil
}
func (d *driveClient) uploadFile(ctx context.Context, parentID, name string, data []byte) (string, error) {
	boundary := "mmvdevkitboundary"
	var b bytes.Buffer
	fmt.Fprintf(&b, "--%s\r\nContent-Type: application/json; charset=UTF-8\r\n\r\n", boundary)
	m, _ := json.Marshal(map[string]any{"name": name, "parents": []string{parentID}})
	b.Write(m)
	fmt.Fprintf(&b, "\r\n--%s\r\nContent-Type: application/octet-stream\r\n\r\n", boundary)
	b.Write(data)
	fmt.Fprintf(&b, "\r\n--%s--\r\n", boundary)
	var out struct {
		ID string `json:"id"`
	}
	e := d.do(ctx, "POST", d.upload+"/files?uploadType=multipart&fields=id", "multipart/related; boundary="+boundary, &b, &out)
	return out.ID, e
}
