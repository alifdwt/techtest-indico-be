package dto

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateVoucherRequest struct {
	// voucher_code, discount_percent, expiry_date
	VoucherCode     string  `json:"voucher_code" binding:"required" validate:"max=255"`
	DiscountPercent float64 `json:"discount_percent" binding:"required" validate:"min=0,max=100"`
	ExpiryDate      string  `json:"expiry_date" binding:"required"`
}

type UpdateVoucherRequest struct {
	// voucher_code, discount_percent, expiry_date
	VoucherCode     string  `json:"voucher_code" binding:"required" validate:"max=255"`
	DiscountPercent float64 `json:"discount_percent" binding:"required" validate:"min=0,max=100"`
	ExpiryDate      string  `json:"expiry_date" binding:"required"`
}

type VoucherResponse struct {
	ID              pgtype.UUID `json:"id"`
	VoucherCode     string      `json:"voucher_code"`
	DiscountPercent int         `json:"discount_percent"`
	ExpiryDate      time.Time   `json:"expiry_date"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type VoucherListQuery struct {
	// search, sort_by, sort_order, page, limit
	Search    string `form:"search"`
	SortBy    string `form:"sort_by" validate:"oneof=expiry_date discount_percent"`
	SortOrder string `form:"sort_order"`
	Page      int    `form:"page,default=1" validate:"min=1"`
	Limit     int    `form:"limit" validate:"min=1,max=100"`
}

type CSVUploadResponse struct {
	SuccessCount int `json:"success_count"`
	FailedCount  int `json:"failed_count"`
}
