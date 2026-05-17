package service

import (
	"context"
	"errors"
	"time"
	"virtual-asset-reconcile-system/internal/notify/model"
	"virtual-asset-reconcile-system/pkg/idgen"

	"gorm.io/gorm"
)

type SendNotifyRequest struct {
	TenantID string
	UserID   string
	OrderNo  string
	Title    string
	Content  string
	Channel  string
}

type SendNotifyResult struct {
	NotifyID int64  `json:"notify_id"`
	Status   string `json:"status"`
}

func SendNotify(ctx context.Context, db *gorm.DB, req SendNotifyRequest) (*SendNotifyResult, error) {
	var existing model.Notification
	err := db.WithContext(ctx).Where("tenant_id = ? AND biz_no = ? AND event_type = ?", req.TenantID, req.OrderNo, "NOTICE_SEND").First(&existing).Error
	if err == nil {
		// found existing request, return current status
		return &SendNotifyResult{
			NotifyID: existing.ID,
			Status:   existing.Status,
		}, nil
	}
	// check err type, if ErrRecordNotFound, create; else return error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// -- create --
	// 1. assemble model
	traceID, _ := ctx.Value("trace_id").(string)
	msg := &model.Notification{
		ID:        idgen.NextID(),
		TenantID:  req.TenantID,
		UserID:    req.UserID,
		BizNo:     req.OrderNo,
		EventType: "NOTICE_SEND",
		Channel:   req.Channel,
		Status:    "PENDING",
		Title:     req.Title,
		Content:   req.Content,
		Result:    "",
		TraceID:   traceID,
	}
	// 2. insert into db
	if err = db.WithContext(ctx).Create(msg).Error; err != nil {
		return nil, err
	}
	// 3. mock notification & update db
	time.Sleep(2 * time.Second)
	msg.Status = "SENT"
	msg.Result = "mock_send_success"
	now := time.Now()
	msg.SentAt = &now
	if err = db.WithContext(ctx).Model(msg).Updates(map[string]interface{}{
		"status":  "SENT",
		"result":  "mock_send_success",
		"sent_at": time.Now(),
	}).Error; err != nil {
		return nil, err
	}
	// 4. return result
	return &SendNotifyResult{
		NotifyID: msg.ID,
		Status:   msg.Status,
	}, nil
}

func GetNotifyStatus(ctx context.Context, db *gorm.DB, tenantID string, orderNo string) (*model.Notification, error) {
	var notification model.Notification
	err := db.WithContext(ctx).Where("tenant_id = ? AND biz_no = ?", tenantID, orderNo).First(&notification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &notification, nil
}
