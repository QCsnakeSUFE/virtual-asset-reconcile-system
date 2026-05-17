package outbox

import (
	"context"
	"time"
	"virtual-asset-reconcile-system/internal/transaction/model"
	"virtual-asset-reconcile-system/pkg/logger"
	"virtual-asset-reconcile-system/pkg/metrics"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Consumer struct {
	db       *gorm.DB
	interval time.Duration
	handlers map[string]Handler
}

type Handler func(ctx context.Context, db *gorm.DB, msg model.OutboxMessage) error

func NewConsumer(db *gorm.DB, interval time.Duration) *Consumer {
	return &Consumer{
		db:       db,
		interval: interval,
		handlers: map[string]Handler{},
	}
}

// To invoke the consumer, run `go consumer.Start(ctx)`
func (c Consumer) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.processMessages(ctx)
		}
	}
}

func (c Consumer) processMessages(ctx context.Context) {
	var err error
	var pendingCount int64
	c.db.WithContext(ctx).Model(&model.OutboxMessage{}).Where("status = ?", "PENDING").Count(&pendingCount)
	metrics.OutboxPendingTotal.Set(float64(pendingCount))

	var messages []model.OutboxMessage
	err = c.db.WithContext(ctx).Where("status = ? AND next_retry_at <= ?", "PENDING", time.Now()).Limit(10).Find(&messages).Error
	if err != nil {
		logger.L.Warn("message not found")
		return
	}
	if len(messages) == 0 {
		return
	}
	for _, msg := range messages {
		handler, ok := c.handlers[msg.EventType]
		if !ok {
			logger.L.Warn("not handler not registered for event type")
			continue
		}
		// update trace_id in context 
		traceCtx := ctx 
		if msg.TraceID != "" {
			traceCtx = context.WithValue(ctx, "trace_id", msg.TraceID)
		}
		handlerErr := handler(traceCtx, c.db, msg)
		if handlerErr != nil {
			logger.L.Warn("handler failed", zap.String("event_type", msg.EventType), zap.Error(handlerErr))
			c.retryOrFail(ctx, &msg)
		} else {
			c.markSent(ctx, &msg)
		}
	}
}

func (c Consumer) retryOrFail(ctx context.Context, msg *model.OutboxMessage) {
	if msg.RetryCount < msg.MaxRetries {
		msg.RetryCount++
		now := time.Now()
		c.db.WithContext(ctx).Model(msg).Updates(map[string]any{
			"retry_count":   msg.RetryCount,
			"last_retry_at": now,
			"next_retry_at": now.Add(time.Duration(msg.RetryCount) * 10 * time.Second),
		})
		return
	}
	c.db.WithContext(ctx).Model(msg).Updates(map[string]any{
		"status": "FAILED",
	})
}

func (c Consumer) markSent(ctx context.Context, msg *model.OutboxMessage) {
	now := time.Now()
	msg.Status = "SENT"
	msg.SentAt = &now
	c.db.WithContext(ctx).Model(msg).Updates(map[string]any{
		"status":  "SENT",
		"sent_at": time.Now(),
	})
}

func (c *Consumer) RegisterHandler(eventType string, handler Handler) {
	c.handlers[eventType] = handler
}
