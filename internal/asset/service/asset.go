package service

import (
	"context"
	"errors"
	"virtual-asset-reconcile-system/internal/asset/model"
	"virtual-asset-reconcile-system/pkg/idgen"
	"virtual-asset-reconcile-system/pkg/logger"

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

func GrantAsset(ctx context.Context, db *gorm.DB, req GrantAssetRequest) (*GrantAssetResult, error) {
	var ledger model.AssetLedger
	// err for idempotency check
	err := db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND item_code = ? AND source_order_no = ?", req.TenantID, req.UserID, req.ItemCode, req.SourceOrderNo).First(&ledger).Error
	// branch: does not exist
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.L.Info("Itempotency check passed. No repeated grant found", zap.String("source_order_no", req.SourceOrderNo), zap.String("item_code", req.ItemCode))
			// no duplicate grant found, grant asset
			var res *GrantAssetResult
			transactionErr := db.Transaction(func(tx *gorm.DB) error {
				var asset model.Asset
				// err for asset lookup
				err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id = ? AND user_id = ? AND item_code = ?", req.TenantID, req.UserID, req.ItemCode).First(&asset).Error

				var balanceBefore int64
				// asset not found, create new asset
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						newAsset := model.Asset{
							ID:           idgen.NextID(),
							TenantID:     req.TenantID,
							UserID:       req.UserID,
							ItemCode:     req.ItemCode,
							Quantity:     req.Quantity,
							TotalGranted: req.Quantity,
						}
						// err for creating new asset
						if err := tx.Create(&newAsset).Error; err != nil {
							logger.L.Info("Create new asset failed.")
							return err
						}
						asset = newAsset
						balanceBefore = 0
					} else {
						logger.L.Info("asset grant failed, waiting for retry...")
						return err
					}
				} else {
					// asset found, update balance, quantity, total_granted
					balanceBefore = asset.Quantity
					asset.Quantity += req.Quantity
					asset.TotalGranted += req.Quantity
					// err for inserting
					if err := tx.Model(&asset).Updates(map[string]any{
						"quantity":      asset.Quantity,
						"total_granted": asset.TotalGranted,
					}).Error; err != nil {
						return err
					}
				}

				// ledge
				newLedger := model.AssetLedger{
					ID:            idgen.NextID(),
					TenantID:      req.TenantID,
					UserID:        req.UserID,
					ItemCode:      req.ItemCode,
					ChangeType:    "GRANT",
					ChangeAmount:  req.Quantity,
					BalanceBefore: balanceBefore,
					BalanceAfter:  asset.Quantity,
					SourceOrderNo: req.SourceOrderNo,
					TraceID:       req.TraceID,
				}
				if err := tx.Create(&newLedger).Error; err != nil {
					return err
				}
				res = &GrantAssetResult{
					AssetID:      asset.ID,
					Balance:      asset.Quantity,
					TotalGranted: asset.TotalGranted,
					LedgerID:     newLedger.ID,
					Duplicate:    false,
				}
				return nil
			})
			if transactionErr != nil {
				logger.L.Error("asset grant transaction failed", zap.String("source_order_no", req.SourceOrderNo), zap.Error(transactionErr))
				return nil, transactionErr
			}
			return res, nil
		}
		return nil, err
	}
	// branch: exists
	var asset model.Asset
	err = db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND item_code = ?", req.TenantID, req.UserID, req.ItemCode).First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &GrantAssetResult{
		AssetID:      asset.ID,
		Balance:      asset.Quantity,
		TotalGranted: asset.TotalGranted,
		LedgerID:     ledger.ID,
		Duplicate:    true,
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
