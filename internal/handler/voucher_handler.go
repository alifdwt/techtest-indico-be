package handler

import (
	"net/http"

	"github.com/alifdwt/techtest-indico-be/internal/dto"
	"github.com/alifdwt/techtest-indico-be/internal/service"
	"github.com/alifdwt/techtest-indico-be/internal/util"
	"github.com/gin-gonic/gin"
)

type VoucherHandler struct {
	voucherService *service.VoucherService
}

func NewVoucherHandler(voucherService *service.VoucherService) *VoucherHandler {
	return &VoucherHandler{
		voucherService: voucherService,
	}
}

// CreateVoucher godoc
// @Summary Create a new voucher
// @Description Create a new voucher with the provided details
// @Tags vouchers
// @Accept json
// @Produce json
// @Param voucher body dto.CreateVoucherRequest true "Voucher data"
// @Success 201 {object} util.Response{data=dto.VoucherResponse}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /vouchers [post]
// @Security BearerAuth
func (vh *VoucherHandler) CreateVoucher(ctx *gin.Context) {
	var req dto.CreateVoucherRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := dto.ValidateStruct(&req); err != nil {
		util.ErrorResponse(ctx, http.StatusBadRequest, "Validation error: "+err.Error())
		return
	}

	res, err := vh.voucherService.CreateVoucher(ctx, &req)
	if err != nil {
		util.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create voucher: "+err.Error())
		return
	}

	util.SuccessResponse(ctx, http.StatusCreated, "Voucher created", res)
}

// ListVouchers godoc
// @Summary List vouchers
// @Description Retrieve a list of vouchers
// @Tags vouchers
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param search query string false "Search term"
// @Param sort_by query string false "Sort by field" default(expiry_date)
// @Param sort_order query string false "Sort order (asc or desc)" default(asc)
// @Success 200 {object} util.Response{data=[]dto.VoucherResponse}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /vouchers [get]
// @Security BearerAuth
func (vh *VoucherHandler) ListVouchers(ctx *gin.Context) {
	var req dto.VoucherListQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		util.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := dto.ValidateStruct(&req); err != nil {
		util.ErrorResponse(ctx, http.StatusBadRequest, "Validation error: "+err.Error())
		return
	}

	res, total, err := vh.voucherService.ListVouchers(ctx, &req)
	if err != nil {
		util.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list vouchers: "+err.Error())
		return
	}

	util.SuccessResponse(ctx, http.StatusOK, "Vouchers listed", gin.H{
		"vouchers": res,
		"total":    total,
	})
}
