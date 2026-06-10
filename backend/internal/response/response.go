package response

import (
	"errors"
	"log/slog"
	"net/http"

	"employee-portal/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     ErrorBody `json:"error"`
	RequestID string    `json:"requestId"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, SuccessResponse{Success: true, Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, SuccessResponse{Success: true, Data: data})
}

func Fail(c *gin.Context, logger *slog.Logger, err error) {
	status, code, message := mapError(err)
	requestID, _ := c.Get("request_id")
	if status >= http.StatusInternalServerError {
		logger.Error("request failed", "error", err, "request_id", requestID)
	}
	c.JSON(status, ErrorResponse{
		Success:   false,
		Error:     ErrorBody{Code: code, Message: message},
		RequestID: requestIDString(requestID),
	})
}

func mapError(err error) (int, string, string) {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr.Kind, domain.ErrUnauthorized):
			return http.StatusUnauthorized, appErr.Code, appErr.Message
		case errors.Is(appErr.Kind, domain.ErrForbidden):
			return http.StatusForbidden, appErr.Code, appErr.Message
		case errors.Is(appErr.Kind, domain.ErrNotFound):
			return http.StatusNotFound, appErr.Code, appErr.Message
		case errors.Is(appErr.Kind, domain.ErrValidation):
			return http.StatusBadRequest, appErr.Code, appErr.Message
		case errors.Is(appErr.Kind, domain.ErrConflict):
			return http.StatusConflict, appErr.Code, appErr.Message
		}
	}
	return http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong"
}

func requestIDString(value any) string {
	if id, ok := value.(string); ok {
		return id
	}
	return ""
}
