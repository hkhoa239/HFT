package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"quantalpha/internal/database"
	"quantalpha/internal/models"
	"quantalpha/internal/validator"
)

type UserHandler struct {
	db database.Querier
}

func NewUserHandler(db database.Querier) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.db.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "failed to list users",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    users,
	})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
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

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "failed to hash password",
		})
		return
	}

	user, err := h.db.CreateUser(c.Request.Context(), req.Username, string(hash), req.Role, req.FullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "failed to create user",
		})
		return
	}

	currentUser, _ := c.Get("user")
	if u, ok := currentUser.(*models.User); ok {
		h.db.CreateAuditLog(c.Request.Context(), u.ID, models.AuditCreate, models.EntityUser, user.ID, "create user")
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    user,
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "invalid user ID",
		})
		return
	}

	var req models.UpdateUserRequest
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

	user, err := h.db.UpdateUser(c.Request.Context(), id, req.Username, req.FullName, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "failed to update user",
		})
		return
	}

	currentUser, _ := c.Get("user")
	if u, ok := currentUser.(*models.User); ok {
		h.db.CreateAuditLog(c.Request.Context(), u.ID, models.AuditUpdate, models.EntityUser, id, "update user")
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    user,
	})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "invalid user ID",
		})
		return
	}

	if err := h.db.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "failed to delete user",
		})
		return
	}

	currentUser, _ := c.Get("user")
	if u, ok := currentUser.(*models.User); ok {
		h.db.CreateAuditLog(c.Request.Context(), u.ID, models.AuditDelete, models.EntityUser, id, "delete user")
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
	})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "unauthorized",
		})
		return
	}

	u := user.(*models.User)
	fetched, err := h.db.GetUserByID(c.Request.Context(), u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "failed to get profile",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    fetched,
	})
}
