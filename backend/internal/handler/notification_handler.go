package handler

import (
	"log"
	"strconv"

	"github.com/go-playground/validator/v10"
	"zeinai-backend/internal/pkg/validation"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/service"
)

type NotificationHandler struct {
	service   *service.NotificationService
	repo      domain.NotificationRepository
	validator *validator.Validate
}

func NewNotificationHandler(s *service.NotificationService, r domain.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{
		service:   s,
		repo:      r,
		validator: validation.NewValidator(),
	}
}

// GetNotifications godoc
// @Summary Get user notifications
// @Description Get paginated list of user's notifications
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} response.APIResponse
// @Router /notifications [get]
func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	// Parse pagination
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	unreadOnly := c.Query("unread") == "true"

	log.Printf("[DEBUG] GetNotifications: userID=%s, limit=%d, offset=%d, unreadOnly=%v", userID, limit, offset, unreadOnly)

	if limit > 100 {
		limit = 100
	}

	var isRead *bool
	if unreadOnly {
		f := false
		isRead = &f
	}

	notifications, err := h.repo.GetByUserID(c.Context(), userID, limit, offset, isRead)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error())
	}

	unreadCount, err := h.repo.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "COUNT_FAILED", err.Error())
	}

	totalCount, err := h.repo.GetTotalCount(c.Context(), userID, isRead)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "COUNT_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{
		"notifications": notifications,
		"unread_count":  unreadCount,
		"total_count":   totalCount,
		"limit":         limit,
		"offset":        offset,
	})
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid notification ID")
	}

	// TODO: Verify notification belongs to user
	_ = userID

	err = h.repo.MarkAsRead(c.Context(), notificationID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UPDATE_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "Notification marked as read"})
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all user's notifications as read
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	err := h.repo.MarkAllAsRead(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UPDATE_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "All notifications marked as read"})
}

// RegisterPushToken godoc
// @Summary Register push notification token
// @Description Register device token for push notifications (FCM/APNS/Web Push)
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.RegisterPushTokenRequest true "Push Token"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /notifications/register-token [post]
func (h *NotificationHandler) RegisterPushToken(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	var req domain.RegisterPushTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	pushToken := &domain.PushToken{
		UserID:     userID.String(),
		Token:      req.Token,
		Platform:   req.Platform,
		DeviceName: req.DeviceName,
	}

	err := h.repo.RegisterPushToken(c.Context(), pushToken)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "REGISTRATION_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{
		"message": "Push token registered successfully",
		"token":   pushToken,
	})
}

// GetPushTokens godoc
// @Summary Get registered push tokens
// @Description Get all active push tokens for current user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=[]domain.PushToken}
// @Failure 401 {object} response.APIResponse
// @Router /notifications/tokens [get]
func (h *NotificationHandler) GetPushTokens(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	tokens, err := h.repo.GetActivePushTokens(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error())
	}

	return response.Success(c, tokens)
}

// DeactivatePushToken godoc
// @Summary Deactivate push token
// @Description Deactivate a specific push token
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Token ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /notifications/tokens/{id} [delete]
func (h *NotificationHandler) DeactivatePushToken(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	tokenID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid token ID")
	}

	// TODO: Verify token belongs to user
	_ = userID

	err = h.repo.DeactivatePushToken(c.Context(), tokenID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "DEACTIVATION_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "Push token deactivated"})
}

// DeleteNotification godoc
// @Summary Delete notification
// @Description Delete a specific notification
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid notification ID")
	}

	// TODO: Verify notification belongs to user
	_ = userID

	err = h.repo.Delete(c.Context(), notificationID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "DELETE_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "Notification deleted"})
}
