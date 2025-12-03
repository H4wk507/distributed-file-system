package handlers

import (
	"dfs-backend/internal/database"
	"dfs-backend/internal/dto"
	"dfs-backend/internal/middleware"
	"dfs-backend/internal/services"
	"dfs-backend/utils/response"
	"encoding/json"
	"errors"
	"net/http"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(db *database.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		service: services.NewAuthService(db, jwtSecret),
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
		if errors.Is(err, services.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to login user")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 hours
	})

	response.JSON(w, http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "User logged in successfully",
	})
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	user, err := h.service.GetUserByID(claims.UserID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "User not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	response.JSON(w, http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    user,
	})
}
