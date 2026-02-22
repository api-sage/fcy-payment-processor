package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/api-sage/fcy-payment-processor/src/internal/adapter/http/models"
	"github.com/api-sage/fcy-payment-processor/src/internal/commons"
	"github.com/api-sage/fcy-payment-processor/src/internal/logger"
	"github.com/api-sage/fcy-payment-processor/src/internal/usecase/service_interfaces"
)

const (
	createUserPath   = "/create-user"
	verifyUserPinPath = "/verify-pin"
)

type UserController struct {
	service service_interfaces.UserService
}

func NewUserController(service service_interfaces.UserService) *UserController {
	return &UserController{service: service}
}

func (c *UserController) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	createUserHandler := http.HandlerFunc(c.createUser)
	verifyPinHandler := http.HandlerFunc(c.verifyUserPin)

	if authMiddleware != nil {
		createUserHandler = authMiddleware(createUserHandler)
		verifyPinHandler = authMiddleware(verifyPinHandler)
	}

	mux.Handle(createUserPath, createUserHandler)
	mux.Handle(verifyUserPinPath, verifyPinHandler)
}

func (c *UserController) createUser(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		response := commons.ErrorResponse[models.CreateUserResponse]("method not allowed")
		c.respondError(w, http.StatusMethodNotAllowed, response, r, start)
		return
	}

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.CreateUserResponse]("invalid request body", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	if err := req.Validate(); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.CreateUserResponse]("validation failed", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	logRequest(r, req)
	response, err := c.service.CreateUser(r.Context(), req)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := mapUserResponseToStatus(response.Message)
		c.respondError(w, status, response, r, start)
		return
	}

	c.respondSuccess(w, http.StatusCreated, response, r, start)
}

func (c *UserController) verifyUserPin(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		response := commons.ErrorResponse[models.VerifyUserPinResponse]("method not allowed")
		c.respondError(w, http.StatusMethodNotAllowed, response, r, start)
		return
	}

	var req models.VerifyUserPinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.VerifyUserPinResponse]("invalid request body", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	if err := req.Validate(); err != nil {
		logError(r, err, nil)
		response := commons.ErrorResponse[models.VerifyUserPinResponse]("validation failed", err.Error())
		c.respondError(w, http.StatusBadRequest, response, r, start)
		return
	}

	logRequest(r, req)
	response, err := c.service.VerifyUserPin(r.Context(), req.CustomerID, req.Pin)
	if err != nil {
		logError(r, err, logger.Fields{"message": response.Message})
		status := mapUserResponseToStatus(response.Message)
		c.respondError(w, status, response, r, start)
		return
	}

	c.respondSuccess(w, http.StatusOK, response, r, start)
}

// mapUserResponseToStatus maps user response messages to appropriate HTTP status codes
func mapUserResponseToStatus(message string) int {
	switch message {
	case "validation failed", "invalid pin":
		return http.StatusBadRequest
	case "User not found":
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// respondSuccess sends a successful JSON response with logging
func (c *UserController) respondSuccess(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write response"})
	}
	logResponse(r, status, payload, start)
}

// respondError sends an error JSON response with logging
func (c *UserController) respondError(w http.ResponseWriter, status int, payload any, r *http.Request, start time.Time) {
	if err := writeJSON(w, status, payload); err != nil {
		logError(r, err, logger.Fields{"action": "write error response"})
	}
	logResponse(r, status, payload, start)
}
