package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
)

type UserService interface {
	CreateUser(ctx context.Context, req models.CreateUserRequest) (commons.Response[models.CreateUserResponse], error)
	GetUser(ctx context.Context, id string) (commons.Response[models.GetUserResponse], error)
	VerifyUserPin(ctx context.Context, customerID string, pin string) (commons.Response[models.VerifyUserPinResponse], error)
}

type UserController struct {
	service UserService
}

func NewUserController(service UserService) *UserController {
	return &UserController{service: service}
}

func (c *UserController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	handler := http.HandlerFunc(c.createUser)
	verifyPinHandler := http.HandlerFunc(c.verifyUserPin)
	if authMiddleware != nil {
		handler = authMiddleware(handler).ServeHTTP
		verifyPinHandler = authMiddleware(verifyPinHandler).ServeHTTP
	}
	mux.Handle("/create-user", http.HandlerFunc(handler))
	mux.Handle("/verify-pin", http.HandlerFunc(verifyPinHandler))
}

func (c *UserController) createUser(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logRequest(r, nil)

	if r.Method != http.MethodPost {
		response := commons.ErrorResponse[models.CreateUserResponse]("method not allowed")
		writeJSON(w, http.StatusMethodNotAllowed, response)
		logResponse(r, http.StatusMethodNotAllowed, response, start)
		return
	}

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.CreateUserResponse]("invalid request body", err.Error())
		writeJSON(w, http.StatusBadRequest, response)
		logResponse(r, http.StatusBadRequest, response, start)
		return
	}
	logRequest(r, req)

	if err := req.Validate(); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.CreateUserResponse]("validation failed", err.Error())
		writeJSON(w, http.StatusBadRequest, response)
		logResponse(r, http.StatusBadRequest, response, start)
		return
	}

	response, err := c.service.CreateUser(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := http.StatusInternalServerError
		if response.Message == "validation failed" {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, response)
		logResponse(r, status, response, start)
		return
	}

	writeJSON(w, http.StatusCreated, response)
	logResponse(r, http.StatusCreated, response, start)
}

func (c *UserController) verifyUserPin(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logRequest(r, nil)

	if r.Method != http.MethodPost {
		response := commons.ErrorResponse[models.VerifyUserPinResponse]("method not allowed")
		writeJSON(w, http.StatusMethodNotAllowed, response)
		logResponse(r, http.StatusMethodNotAllowed, response, start)
		return
	}

	var req models.VerifyUserPinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.VerifyUserPinResponse]("invalid request body", err.Error())
		writeJSON(w, http.StatusBadRequest, response)
		logResponse(r, http.StatusBadRequest, response, start)
		return
	}
	logRequest(r, req)

	if err := req.Validate(); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.VerifyUserPinResponse]("validation failed", err.Error())
		writeJSON(w, http.StatusBadRequest, response)
		logResponse(r, http.StatusBadRequest, response, start)
		return
	}

	response, err := c.service.VerifyUserPin(r.Context(), req.CustomerID, req.Pin)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := http.StatusInternalServerError
		if response.Message == "validation failed" || response.Message == "invalid pin" {
			status = http.StatusBadRequest
		}
		if response.Message == "User not found" {
			status = http.StatusNotFound
		}
		writeJSON(w, status, response)
		logResponse(r, status, response, start)
		return
	}

	writeJSON(w, http.StatusOK, response)
	logResponse(r, http.StatusOK, response, start)
}
