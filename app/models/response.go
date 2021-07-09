package models

import "net/http"

type ErrorResponse struct {
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
	Code    int               `json:"code"`
}

var (
	EmailAlreadyExists = ErrorResponse{
		Message: "User already exists",
		Errors:  map[string]string{"email": "User already exists"},
		Code:    http.StatusBadRequest,
	}
	UsernameAlreadyExists = ErrorResponse{
		Message: "User already exists",
		Errors:  map[string]string{"username": "User already exists"},
		Code:    http.StatusBadRequest,
	}
	UserNotFound = ErrorResponse{
		Message: "User not found",
		Code:    http.StatusNotFound,
	}
	InternalServerError = ErrorResponse{
		Message: "Internal server error",
		Code:    http.StatusInternalServerError,
	}
	Unauthorized = ErrorResponse{
		Message: "Invalid email or password",
		Code:    http.StatusUnauthorized,
	}
)
