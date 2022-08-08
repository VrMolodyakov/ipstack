package v1

import (
	"context"
	"ipstack/internal/domain/entity"
	"ipstack/internal/domain/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service service.UserSerivce
}

func NewUserHandler(service service.UserSerivce) UserHandler {
	return UserHandler{service: service}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user entity.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}
	_, err := h.service.Create(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := h.service.GetAll(context.Background())
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}
	c.JSON(http.StatusOK, users)
}
