package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"quantalpha/internal/database"
	"quantalpha/internal/models"
	"quantalpha/internal/utils"
	"quantalpha/internal/validator"
)

type AuthHandler struct {
	db         *database.Queries
	jwtSec     string
	expireHour int
}

func NewAuthHandler(db *database.Queries, jwtSec string, expireHour int) *AuthHandler {
	return &AuthHandler{db: db, jwtSec: jwtSec, expireHour: expireHour}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	if errs := validator.Validate(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   validator.GetFirstError(errs),
		})
		return
	}

	user, passwordHash, err := h.db.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "invalid credentials",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "invalid credentials",
		})
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Username, user.Role, h.jwtSec, h.expireHour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "failed to generate token",
		})
		return
	}

	h.db.CreateAuditLog(c.Request.Context(), user.ID, models.AuditCreate, models.EntityUser, user.ID, "user login")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.LoginResponse{
			Token: token,
			User:  user,
		},
	})
}
