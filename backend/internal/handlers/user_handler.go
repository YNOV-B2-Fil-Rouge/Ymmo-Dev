package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ymmo/internal/middleware"
	"ymmo/internal/services"
)

type UserHandler struct {
	svc *services.UserService
}

func NewUserHandler(svc *services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// @Summary      List all users (IT)
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /management/users [get]
func (h *UserHandler) ListAll(c *gin.Context) {
	items, err := h.svc.ListAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// @Summary      Delete my account
// @Tags         users
// @Security     BearerAuth
// @Success      204  "No Content"
// @Failure      401  {object}  map[string]string
// @Router       /me [delete]
func (h *UserHandler) DeleteMe(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	if err := h.svc.DeleteAccount(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete account"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary      Delete a user account (IT/HQ)
// @Tags         users
// @Param        id   path  int  true  "User ID"
// @Security     BearerAuth
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /management/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	// Staff cannot delete their own account through the admin route.
	if uint(id) == middleware.CurrentUserID(c) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "use the profile page to delete your own account"})
		return
	}
	if err := h.svc.DeleteAccount(uint(id)); err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete user"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary      List collaborators
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /management/collaborators [get]
func (h *UserHandler) ListCollaborators(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	role := middleware.CurrentRole(c)
	items, err := h.svc.ListCollaborators(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch collaborators"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}
