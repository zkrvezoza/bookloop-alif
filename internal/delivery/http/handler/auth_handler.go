package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bookloop-alif/internal/domain/model"
	"github.com/bookloop-alif/internal/response"
	"github.com/bookloop-alif/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"` // "user" | "librarian"
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, model.ErrInvalid)
		return
	}

	u, err := h.auth.Register(c.Request.Context(), req.Login, req.Password, model.Role(req.Role))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, gin.H{"id": u.ID, "login": u.Login, "role": u.Role})
}

type loginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, model.ErrInvalid)
		return
	}

	pair, err := h.auth.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, model.ErrInvalid)
		return
	}

	pair, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, model.ErrInvalid)
		return
	}

	if err := h.auth.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
