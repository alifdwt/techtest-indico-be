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

// GetVoucher godoc
// @Summary Get voucher by ID
// @Description Get a specific voucher by its ID
// @Tags vouchers
// @Produce json
// @Param id path string true "Voucher ID"
// @Success 200 {object} util.Response{data=dto.VoucherResponse}
// @Failure 400 {object} util.Response
// @Failure 404 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /vouchers/{id} [get]
// @Security BearerAuth
func (vh *VoucherHandler) GetVoucher(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		util.ErrorResponse(ctx, http.StatusBadRequest, "Voucher ID is required")
		return
	}

	res, err := vh.voucherService.GetVoucherByID(ctx, id)
	if err != nil {
		if err.Error() == "voucher not found" {
			util.ErrorResponse(ctx, http.StatusNotFound, "Voucher not found")
			return
		}
		util.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get voucher: "+err.Error())
		return
	}

	util.SuccessResponse(ctx, http.StatusOK, "Voucher retrieved", res)
}

// UpdateVoucher godoc
// @Summary Update a voucher
// @Description Update an existing voucher with new details
// @Tags vouchers
// @Accept json
// @Produce json
// @Param id path string true "Voucher ID"
// @Param voucher body dto.UpdateVoucherRequest true "Updated voucher data"
// @Success 200 {object} util.Response{data=dto.VoucherResponse}
// @Failure 400 {object} util.Response
// @Failure 404 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /vouchers/{id} [put]
// @Security BearerAuth
func (vh *VoucherHandler) UpdateVoucher(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		util.ErrorResponse(ctx, http.StatusBadRequest, "Voucher ID is required")
		return
	}

	var req dto.UpdateVoucherRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		util.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := dto.ValidateStruct(&req); err != nil {
		util.ErrorResponse(ctx, http.StatusBadRequest, "Validation error: "+err.Error())
		return
	}

	res, err := vh.voucherService.UpdateVoucher(ctx, id, &req)
	if err != nil {
		if err.Error() == "voucher not found" {
			util.ErrorResponse(ctx, http.StatusNotFound, "Voucher not found")
			return
		}
		util.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update voucher: "+err.Error())
		return
	}

	util.SuccessResponse(ctx, http.StatusOK, "Voucher updated", res)
}

// DeleteVoucher godoc
// @Summary Delete a voucher
// @Description Delete a voucher by its ID
// @Tags vouchers
// @Produce json
// @Param id path string true "Voucher ID"
// @Success 200 {object} util.Response
// @Failure 400 {object} util.Response
// @Failure 404 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /vouchers/{id} [delete]
// @Security BearerAuth
func (vh *VoucherHandler) DeleteVoucher(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		util.ErrorResponse(ctx, http.StatusBadRequest, "Voucher ID is required")
		return
	}

	err := vh.voucherService.DeleteVoucher(ctx, id)
	if err != nil {
		if err.Error() == "voucher not found" {
			util.ErrorResponse(ctx, http.StatusNotFound, "Voucher not found")
			return
		}
		util.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete voucher: "+err.Error())
		return
	}

	util.SuccessResponse(ctx, http.StatusOK, "Voucher deleted", nil)
}
