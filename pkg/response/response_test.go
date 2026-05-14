package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSuccessResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("trace_id", "test-trace-123")

	Success(c, map[string]string{"order_no": "ORD20250101001"})

	var resp APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("expected message 'success', got '%s'", resp.Message)
	}
	if resp.TraceID != "test-trace-123" {
		t.Errorf("expected trace_id 'test-trace-123', got '%s'", resp.TraceID)
	}
	if resp.Data == nil {
		t.Errorf("expected non-nil data")
	}
}

func TestErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("trace_id", "test-trace-456")

	Error(c, http.StatusBadRequest, 1001, "invalid order id")

	var resp APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Code != 1001 {
		t.Errorf("expected code 1001, got %d", resp.Code)
	}
	if resp.Message != "invalid order id" {
		t.Errorf("expected message 'invalid order id', got '%s'", resp.Message)
	}
	if resp.TraceID != "test-trace-456" {
		t.Errorf("expected trace_id 'test-trace-456', got '%s'", resp.TraceID)
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestSuccessResponseTraceIDFromContext(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// No trace_id set — this should panic or return empty string
	c.Set("trace_id", "")

	Success(c, nil)

	var resp APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
}
