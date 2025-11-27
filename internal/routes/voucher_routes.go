package routes

import (
	"github.com/alifdwt/techtest-indico-be/internal/handler"
	"github.com/alifdwt/techtest-indico-be/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupVoucherRoutes(
	router *gin.Engine,
	voucherHandler *handler.VoucherHandler,
) {
	voucherGroup := router.Group("/vouchers")
	voucherGroup.Use(middleware.AuthMiddleware())
	{
		voucherGroup.POST("", voucherHandler.CreateVoucher)
		voucherGroup.GET("", voucherHandler.ListVouchers)
		voucherGroup.GET("/:id", voucherHandler.GetVoucher)
		voucherGroup.PUT("/:id", voucherHandler.UpdateVoucher)
		voucherGroup.DELETE("/:id", voucherHandler.DeleteVoucher)

		voucherGroup.POST("/upload-csv", voucherHandler.UploadCSV)
	}
}
