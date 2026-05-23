package response

import (
	"errors"
	"exchangeapp/internal/apperrors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code    int `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, Body{
		Code: 0,
		Message: "success",
		Data: data,
	})
}

func Created(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusCreated, Body{
		Code: 0,
		Message: "success",
		Data: data,
	})
}

func Error(ctx *gin.Context, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		ctx.JSON(appErr.HTTPStatus, Body{
			Code: appErr.Code,
			Message: appErr.Message,
		})
		return
	}

	appErr = apperrors.ErrInternal
	ctx.JSON(appErr.HTTPStatus, Body{
		Code: appErr.Code,
		Message: appErr.Message,
	})
}