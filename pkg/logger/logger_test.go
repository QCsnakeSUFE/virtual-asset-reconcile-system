package logger

import (
	"context"
	"testing"
)

func TestInit(t *testing.T) {
	Init("test-service")
	if L == nil {
		t.Fatal("expected L to be non-nil after Init")
	}
}

func TestFromContext(t *testing.T) {
	Init("test-service")

	ctx := context.WithValue(context.Background(), "trace_id", "trace-001")
	ctx = context.WithValue(ctx, "tenant_id", "tenant-001")
	ctx = context.WithValue(ctx, "user_id", "user-001")

	l := FromContext(ctx)
	if l == nil {
		t.Fatal("expected non-nil logger from context")
	}

	// 能用 Info/Warn/Error 不 panic 就算通过
	l.Info("test log from context")
}

func TestFromContextEmptyContext(t *testing.T) {
	Init("test-service")

	l := FromContext(context.Background())
	if l == nil {
		t.Fatal("expected non-nil logger from empty context")
	}

	l.Info("test log from empty context")
}

func TestFromContextWithoutInit(t *testing.T) {
	oldL := L
	L = nil
	defer func() { L = oldL }()

	l := FromContext(context.Background())
	if l == nil {
		t.Fatal("expected non-nil logger even without Init")
	}
}

func TestInitDevEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	Init("test-dev")
	if L == nil {
		t.Fatal("expected L to be non-nil")
	}
	L.Info("dev mode log should be console format")
}

func TestInitLogLevelDebug(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	Init("test-debug")
	if L == nil {
		t.Fatal("expected L to be non-nil")
	}
	L.Debug("debug log should appear")
}
