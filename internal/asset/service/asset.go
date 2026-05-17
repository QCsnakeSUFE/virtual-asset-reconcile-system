package service

import (
	"context"
	"errors"
	"virtual-asset-reconcile-system/internal/asset/model"
	"virtual-asset-reconcile-system/pkg/idgen"
	"virtual-asset-reconcile-system/pkg/logger"
	"virtual-asset-reconcile-system/pkg/metrics"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GrantAssetRequest struct {
	TenantID      string
	UserID        string
	ItemCode      string
	Quantity      int64
	SourceOrderNo string
	TraceID       string
}

type GrantAssetResult struct {
	AssetID      int64 `json:"asset_id"`
	Balance      int64 `json:"balance"`
	TotalGranted int64 `json:"total_granted"`
	LedgerID     int64 `json:"ledger_id"`
	Duplicate    bool  `json:"duplicate"`
}

type UserAssetItem struct {
	ItemCode      string `json:"item_code"`
	Quantity      int64  `json:"quantity"`
	Frozen        int64  `json:"frozen"`
	TotalGranted  int64  `json:"total_granted"`
	TotalConsumed int64  `json:"total_consumed"`
}

type GrantStatusResult struct {
	Processed bool   `json:"processed"`
	ItemCode  string `json:"item_code,omitempty"`
	Amount    int64  `json:"amount,omitempty"`
	Status    string `json:"status,omitempty"`
	LedgerID  int64  `json:"ledger_id,omitempty"`
}

type modifyAssetResult struct {
	asset     model.Asset
	ledger    model.AssetLedger
	duplicate bool
}

type modifyAssetConfig struct {
	changeType   string
	allowCreate  bool
	checkBalance bool
}

// modifyAsset handles both grant and deduct with a unified logic path.
//   - For "GRANT": allowCreate=true, checkBalance=false, automatically creates asset if not found.
//   - For "CONSUME": allowCreate=false, checkBalance=true, returns error if insufficient balance.
func modifyAsset(ctx context.Context, db *gorm.DB, req GrantAssetRequest, cfg modifyAssetConfig) (*modifyAssetResult, error) {
	var ledger model.AssetLedger
	err := db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND item_code = ? AND source_order_no = ?",
		req.TenantID, req.UserID, req.ItemCode, req.SourceOrderNo).First(&ledger).Error
	if err == nil {
		var asset model.Asset
		db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND item_code = ?",
			req.TenantID, req.UserID, req.ItemCode).First(&asset)
		return &modifyAssetResult{asset: asset, ledger: ledger, duplicate: true}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var res *modifyAssetResult
	transactionErr := db.Transaction(func(tx *gorm.DB) error {
		var asset model.Asset
		err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("tenant_id = ? AND user_id = ? AND item_code = ?", req.TenantID, req.UserID, req.ItemCode).First(&asset).Error

		var balanceBefore int64
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) && cfg.allowCreate {
				newAsset := model.Asset{
					ID:           idgen.NextID(),
					TenantID:     req.TenantID,
					UserID:       req.UserID,
					ItemCode:     req.ItemCode,
					Quantity:     req.Quantity,
					TotalGranted: req.Quantity,
				}
				if err := tx.Create(&newAsset).Error; err != nil {
					logger.L.Info("create new asset failed")
					return err
				}
				asset = newAsset
				balanceBefore = 0
			} else {
				logger.L.Info("asset lookup failed", zap.Error(err))
				return err
			}
		} else {
			if cfg.checkBalance && asset.Quantity < req.Quantity {
				return errors.New("insufficient asset balance")
			}
			balanceBefore = asset.Quantity
			if cfg.changeType == "GRANT" {
				asset.Quantity += req.Quantity
				asset.TotalGranted += req.Quantity
			} else {
				asset.Quantity -= req.Quantity
				asset.TotalConsumed += req.Quantity
			}
			if err := tx.Model(&asset).Updates(map[string]any{
				"quantity":       asset.Quantity,
				"total_granted":  asset.TotalGranted,
				"total_consumed": asset.TotalConsumed,
			}).Error; err != nil {
				return err
			}
		}

		var changeAmount int64
		if cfg.changeType == "GRANT" {
			changeAmount = req.Quantity
		} else {
			changeAmount = -req.Quantity
		}
		remark := ""
		if cfg.changeType == "CONSUME" {
			remark = "gift deduction"
		}
		newLedger := model.AssetLedger{
			ID:            idgen.NextID(),
			TenantID:      req.TenantID,
			UserID:        req.UserID,
			ItemCode:      req.ItemCode,
			ChangeType:    cfg.changeType,
			ChangeAmount:  changeAmount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  asset.Quantity,
			SourceOrderNo: req.SourceOrderNo,
			Remark:        remark,
			TraceID:       req.TraceID,
		}
		if err := tx.Create(&newLedger).Error; err != nil {
			return err
		}

		res = &modifyAssetResult{asset: asset, ledger: newLedger, duplicate: false}
		return nil
	})
	if transactionErr != nil {
		return nil, transactionErr
	}
	return res, nil
}

func GrantAsset(ctx context.Context, db *gorm.DB, req GrantAssetRequest) (*GrantAssetResult, error) {
	result, err := modifyAsset(ctx, db, req, modifyAssetConfig{
		changeType:   "GRANT",
		allowCreate:  true,
		checkBalance: false,
	})
	if err != nil {
		metrics.AssetGrantFailedTotal.Inc()
		logger.L.Error("asset grant failed", zap.String("source_order_no", req.SourceOrderNo), zap.Error(err))
		return nil, err
	}
	if result.duplicate {
		return &GrantAssetResult{
			AssetID:      result.asset.ID,
			Balance:      result.asset.Quantity,
			TotalGranted: result.asset.TotalGranted,
			LedgerID:     result.ledger.ID,
			Duplicate:    true,
		}, nil
	}
	metrics.AssetGrantTotal.Inc()
	return &GrantAssetResult{
		AssetID:      result.asset.ID,
		Balance:      result.asset.Quantity,
		TotalGranted: result.asset.TotalGranted,
		LedgerID:     result.ledger.ID,
		Duplicate:    false,
	}, nil
}

type DeductAssetResult struct {
	AssetID       int64 `json:"asset_id"`
	Balance       int64 `json:"balance"`
	TotalConsumed int64 `json:"total_consumed"`
	LedgerID      int64 `json:"ledger_id"`
	Duplicate     bool  `json:"duplicate"`
}

func DeductAsset(ctx context.Context, db *gorm.DB, req GrantAssetRequest) (*DeductAssetResult, error) {
	result, err := modifyAsset(ctx, db, req, modifyAssetConfig{
		changeType:   "CONSUME",
		allowCreate:  false,
		checkBalance: true,
	})
	if err != nil {
		return nil, err
	}
	return &DeductAssetResult{
		AssetID:       result.asset.ID,
		Balance:       result.asset.Quantity,
		TotalConsumed: result.asset.TotalConsumed,
		LedgerID:      result.ledger.ID,
		Duplicate:     result.duplicate,
	}, nil
}

func GetUserAssets(ctx context.Context, db *gorm.DB, tenantID, userID string) ([]UserAssetItem, error) {
	var assets []model.Asset
	err := db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).Find(&assets).Error
	if err != nil {
		return nil, err
	}
	items := make([]UserAssetItem, len(assets))
	for i, asset := range assets {
		items[i] = UserAssetItem{
			ItemCode:      asset.ItemCode,
			Quantity:      asset.Quantity,
			Frozen:        asset.Frozen,
			TotalGranted:  asset.TotalGranted,
			TotalConsumed: asset.TotalConsumed,
		}
	}
	return items, nil
}

func GetGrantStatus(ctx context.Context, db *gorm.DB, tenantID, sourceOrderNo string) (*GrantStatusResult, error) {
	var ledger model.AssetLedger
	err := db.WithContext(ctx).Where("tenant_id = ? AND source_order_no = ?", tenantID, sourceOrderNo).First(&ledger).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &GrantStatusResult{Processed: false}, nil
		}
		return nil, err
	}
	return &GrantStatusResult{
		Processed: true,
		ItemCode:  ledger.ItemCode,
		Amount:    ledger.ChangeAmount,
		Status:    "SUCCESS",
		LedgerID:  ledger.ID,
	}, nil
}
