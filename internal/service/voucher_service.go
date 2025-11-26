package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alifdwt/techtest-indico-be/internal/dto"
	"github.com/alifdwt/techtest-indico-be/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type VoucherService struct {
	repo *repository.Queries
}

func NewVoucherService(repo *repository.Queries) *VoucherService {
	return &VoucherService{
		repo: repo,
	}
}

func (s *VoucherService) CreateVoucher(ctx context.Context, req *dto.CreateVoucherRequest) (*dto.VoucherResponse, error) {
	existingVoucher, err := s.repo.GetVoucherByCode(ctx, req.VoucherCode)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	if existingVoucher.ID.Valid {
		return nil, fmt.Errorf("voucher with code %s already exists", req.VoucherCode)
	}

	expiryDateTime, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		return nil, err
	}

	obj := repository.CreateVoucherParams{
		VoucherCode:     req.VoucherCode,
		DiscountPercent: int32(req.DiscountPercent),
		ExpiryDate:      pgtype.Timestamptz{Time: expiryDateTime, Valid: true},
	}
	voucher, err := s.repo.CreateVoucher(ctx, obj)
	if err != nil {
		return nil, err
	}

	return s.toVoucherResponse(&voucher), nil
}

func (s *VoucherService) toVoucherResponse(voucher *repository.Voucher) *dto.VoucherResponse {
	return &dto.VoucherResponse{
		ID:              voucher.ID,
		VoucherCode:     voucher.VoucherCode,
		DiscountPercent: int(voucher.DiscountPercent),
		ExpiryDate:      voucher.ExpiryDate.Time,
		CreatedAt:       voucher.CreatedAt.Time,
		UpdatedAt:       voucher.UpdatedAt.Time,
	}
}

func (s *VoucherService) ListVouchers(ctx context.Context, query *dto.VoucherListQuery) ([]*dto.VoucherResponse, int64, error) {
	sortOrder := query.SortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	var searchSQL pgtype.Text
	if query.Search != "" {
		searchSQL = pgtype.Text{String: query.Search, Valid: true}
	} else {
		searchSQL = pgtype.Text{Valid: false}
	}

	offset := (query.Page - 1) * query.Limit

	obj := repository.ListVouchersParams{
		Search:    searchSQL,
		SortBy:    pgtype.Text{String: query.SortBy, Valid: true},
		SortOrder: pgtype.Text{String: sortOrder, Valid: true},
		Limit:     int32(query.Limit),
		Offset:    int32(offset),
	}
	vouchers, err := s.repo.ListVouchers(ctx, obj)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountVouchers(ctx, obj.Search)
	if err != nil {
		return nil, 0, err
	}

	var responses []*dto.VoucherResponse
	for _, voucher := range vouchers {
		responses = append(responses, s.toVoucherResponse(&voucher))
	}

	return responses, total, nil
}

func (s *VoucherService) GetVoucherByID(ctx context.Context, id string) (*dto.VoucherResponse, error) {
	voucherID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid voucher id")
	}

	uuidPg := pgtype.UUID{Bytes: voucherID, Valid: true}

	voucher, err := s.repo.GetVoucherByID(ctx, uuidPg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("voucher not found")
		}
		return nil, err
	}

	return s.toVoucherResponse(&voucher), nil
}

func (s *VoucherService) UpdateVoucher(ctx context.Context, id string, req *dto.UpdateVoucherRequest) (*dto.VoucherResponse, error) {
	voucherID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid voucher id")
	}

	uuidPg := pgtype.UUID{Bytes: voucherID, Valid: true}

	_, err = s.repo.GetVoucherByID(ctx, uuidPg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("voucher not found")
		}
		return nil, err
	}

	expiryDateTime, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		return nil, err
	}

	obj := repository.UpdateVoucherParams{
		ID:              uuidPg,
		VoucherCode:     req.VoucherCode,
		DiscountPercent: int32(req.DiscountPercent),
		ExpiryDate:      pgtype.Timestamptz{Time: expiryDateTime, Valid: true},
	}
	voucher, err := s.repo.UpdateVoucher(ctx, obj)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("voucher with code %s already exists", req.VoucherCode)
		}
		return nil, err
	}

	return s.toVoucherResponse(&voucher), nil
}

func (s *VoucherService) DeleteVoucher(ctx context.Context, id string) error {
	voucherID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid voucher id")
	}

	uuidPg := pgtype.UUID{Bytes: voucherID, Valid: true}

	_, err = s.repo.GetVoucherByID(ctx, uuidPg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("voucher not found")
		}
		return err
	}

	return s.repo.DeleteVoucher(ctx, uuidPg)
}
