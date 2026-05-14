package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var client = &http.Client{Timeout: 10 * time.Second}

type RequestOpts struct {
	Method string
	URL    string
	Body   interface{}
}

func DoJSON(ctx context.Context, opts RequestOpts, target interface{}) (int, error) {
	var bodyReader io.Reader
	if opts.Body != nil {
		data, err := json.Marshal(opts.Body)
		if err != nil {
			return 0, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, opts.Method, opts.URL, bodyReader)
	if err != nil {
		return 0, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if traceID, ok := ctx.Value("trace_id").(string); ok && traceID != "" {
		req.Header.Set("X-Trace-Id", traceID)
	}
	if tenantID, ok := ctx.Value("tenant_id").(string); ok && tenantID != "" {
		req.Header.Set("X-Tenant-Id", tenantID)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return resp.StatusCode, fmt.Errorf("decode response: %w", err)
		}
	}

	return resp.StatusCode, nil
}
