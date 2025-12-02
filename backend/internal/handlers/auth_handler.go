package handlers

import (
	"dfs-backend/internal/database"
	"dfs-backend/internal/dto"
	"dfs-backend/internal/services"
	"dfs-backend/utils/response"
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(db *database.DB) *AuthHandler {
	return &AuthHandler{
		service: services.NewAuthService(db),
	}
}

func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var user dto.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.RegisterUser(&user); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	response.JSON(w, http.StatusCreated, response.SuccessResponse{
		Success: true,
		Message: "User registered successfully",
	})
}

func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var user dto.LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	token, err := h.service.LoginUser(&user)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to login user")
		return
	}

	response.JSON(w, http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "User logged in successfully",
		Data:    token,
	})
}
