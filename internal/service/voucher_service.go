package service

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
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

	if responses == nil {
		responses = []*dto.VoucherResponse{}
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

func (s *VoucherService) UploadCSV(ctx context.Context, file io.Reader) (*dto.CSVUploadResponse, error) {
	requiredHeaders := []string{"voucher_code", "discount_percent", "expiry_date"}
	reader := csv.NewReader(file)
	var successCount int
	var failedRows []dto.FailedRow

	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("csv file is empty")
		}
		return nil, fmt.Errorf("failed to read csv headers: %w", err)
	}

	headerMap := make(map[string]int)
	for i, header := range headers {
		normalizedHeader := strings.TrimSpace(strings.ToLower(header))
		headerMap[normalizedHeader] = i
	}

	for _, required := range requiredHeaders {
		if _, ok := headerMap[required]; !ok {
			return nil, fmt.Errorf("header '%s' not found in the csv header", required)
		}
	}

	lineNumber := 1

	for {
		record, err := reader.Read()
		lineNumber++

		if err == io.EOF {
			break
		}
		if err != nil {
			failedRows = append(failedRows, dto.FailedRow{
				RowNumber: lineNumber,
				Reason:    "csv row format is not vlaid",
			})
			continue
		}

		voucherCode := strings.TrimSpace(record[headerMap["voucher_code"]])
		discountPercentStr := strings.TrimSpace(record[headerMap["discount_percent"]])
		expiryDateStr := strings.TrimSpace(record[headerMap["expiry_date"]])

		recordFailed := func(reason string) {
			failedRows = append(failedRows, dto.FailedRow{
				RowNumber:   lineNumber,
				VoucherCode: voucherCode,
				Reason:      reason,
			})
		}

		if voucherCode == "" || discountPercentStr == "" || expiryDateStr == "" {
			recordFailed("voucher_code, discount_percent, or expiry_date are empty.")
			continue
		}

		var discountPercent int
		discountPercent, err = strconv.Atoi(discountPercentStr)
		if err != nil {
			recordFailed(fmt.Sprintf("Discount percent must be a number: %s", err.Error()))
			continue
		}
		if discountPercent < 0 || discountPercent > 100 {
			recordFailed("Discount percent must be between 0 and 100.")
			continue
		}

		expiryDate, err := time.Parse("2006-01-02", expiryDateStr)
		if err != nil {
			expiryDate, err = time.Parse("2006-01-02 15:04:05", expiryDateStr)
			if err != nil {

				recordFailed("expiry_date format is not valid. Use YYYY-MM-DD or YYYY-MM-DD HH:MM:SS.")
				continue
			}
		}

		obj := repository.CreateVoucherParams{
			VoucherCode:     voucherCode,
			DiscountPercent: int32(discountPercent),
			ExpiryDate:      pgtype.Timestamptz{Time: expiryDate, Valid: true},
		}
		_, err = s.repo.CreateVoucher(ctx, obj)
		if err != nil {
			recordFailed("Failed to save to database (Possibly duplicate voucher_code)")
			continue
		}

		successCount++
	}

	return &dto.CSVUploadResponse{
		SuccessCount: successCount,
		FailedCount:  len(failedRows),
		FailedRows:   failedRows,
	}, nil
}

func (s *VoucherService) ExportCSV(ctx context.Context) ([][]string, error) {
	vouchers, err := s.repo.GetAllVouchersForExport(ctx)
	if err != nil {
		return nil, err
	}

	var records [][]string
	records = append(records, []string{"ID", "Voucher Code", "Discount Percent", "Expiry Date", "Created At", "Updated At"})

	for _, voucher := range vouchers {
		record := []string{
			voucher.ID.String(),
			voucher.VoucherCode,
			strconv.Itoa(int(voucher.DiscountPercent)),
			voucher.ExpiryDate.Time.Format("2006-01-02 15:04:05"),
			voucher.CreatedAt.Time.Format("2006-01-02 15:04:05"),
			voucher.UpdatedAt.Time.Format("2006-01-02 15:04:05"),
		}
		records = append(records, record)
	}

	return records, nil
}
