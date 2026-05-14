package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func TestDoJSON_MethodPassthrough(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			var receivedMethod string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedMethod = r.Method
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			_, err := DoJSON(context.Background(), RequestOpts{
				Method: method,
				URL:    ts.URL,
			}, nil)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if receivedMethod != method {
				t.Errorf("expected method %s, got %s", method, receivedMethod)
			}
		})
	}
}

func TestDoJSON_POST_WithBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", contentType)
		}
		var reqBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if reqBody["key"] != "value" {
			t.Errorf("expected key=value, got %v", reqBody)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(testResponse{Code: 0, Message: "ok"})
	}))
	defer ts.Close()

	body := map[string]string{"key": "value"}
	var resp testResponse
	statusCode, err := DoJSON(context.Background(), RequestOpts{
		Method: http.MethodPost,
		URL:    ts.URL,
		Body:   body,
	}, &resp)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", statusCode)
	}
	if resp.Code != 0 || resp.Message != "ok" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestDoJSON_GET_WithoutBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Body != nil {
			bodyLen := r.ContentLength
			if bodyLen > 0 {
				t.Errorf("expected no body for GET, but got content length %d", bodyLen)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(testResponse{Code: 0, Message: "ok"})
	}))
	defer ts.Close()

	var resp testResponse
	statusCode, err := DoJSON(context.Background(), RequestOpts{
		Method: http.MethodGet,
		URL:    ts.URL,
	}, &resp)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", statusCode)
	}
	if resp.Code != 0 || resp.Message != "ok" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestDoJSON_TraceIDPropagation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-Id")
		if traceID != "trace-001" {
			t.Errorf("expected X-Trace-Id trace-001, got %s", traceID)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx := context.WithValue(context.Background(), "trace_id", "trace-001")
	_, err := DoJSON(ctx, RequestOpts{
		Method: http.MethodGet,
		URL:    ts.URL,
	}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoJSON_TenantIDPropagation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-Id")
		if tenantID != "tenant-001" {
			t.Errorf("expected X-Tenant-Id tenant-001, got %s", tenantID)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx := context.WithValue(context.Background(), "tenant_id", "tenant-001")
	_, err := DoJSON(ctx, RequestOpts{
		Method: http.MethodGet,
		URL:    ts.URL,
	}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoJSON_NilTarget(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":0,"message":"ok"}`))
	}))
	defer ts.Close()

	statusCode, err := DoJSON(context.Background(), RequestOpts{
		Method: http.MethodGet,
		URL:    ts.URL,
	}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", statusCode)
	}
}

func TestDoJSON_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer ts.Close()

	statusCode, err := DoJSON(context.Background(), RequestOpts{
		Method: http.MethodGet,
		URL:    ts.URL,
	}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", statusCode)
	}
}

func TestDoJSON_EmptyContext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Trace-Id") != "" {
			t.Errorf("expected empty X-Trace-Id, got %s", r.Header.Get("X-Trace-Id"))
		}
		if r.Header.Get("X-Tenant-Id") != "" {
			t.Errorf("expected empty X-Tenant-Id, got %s", r.Header.Get("X-Tenant-Id"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, err := DoJSON(context.Background(), RequestOpts{
		Method: http.MethodGet,
		URL:    ts.URL,
	}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
