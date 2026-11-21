package handler

import (
	"fmt"
	"log"
	"sync"

	"zeinai-backend/internal/config"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/middleware"
	"zeinai-backend/internal/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

type WhiteboardHandler struct {
	service domain.WhiteboardService
}

func NewWhiteboardHandler(service domain.WhiteboardService) *WhiteboardHandler {
	return &WhiteboardHandler{
		service: service,
	}
}

// In-memory hub for WebSockets
var (
	hubMutex sync.RWMutex
	// map[classID]map[connID]*websocket.Conn
	hub = make(map[int64]map[string]*websocket.Conn)
)

func (h *WhiteboardHandler) RegisterRoutes(api fiber.Router, cfg interface{}) {
	whiteboard := api.Group("/whiteboard", middleware.Protected(cfg.(*config.Config)))

	// Session management
	whiteboard.Post("/start", h.StartSession)
	whiteboard.Post("/end", h.EndSession)
	whiteboard.Get("/active", h.GetActiveSession)
	whiteboard.Get("/session/:sessionId", h.GetSession)
	whiteboard.Get("/history/:classId", h.GetHistory)
	whiteboard.Get("/student/history", h.GetStudentHistory)
	whiteboard.Get("/summary/:sessionId", h.GetSummary)

	// WebSocket for real-time sync
	api.Get("/ws/whiteboard/:classId", websocket.New(h.WebSocketHandler))
}

func (h *WhiteboardHandler) StartSession(c *fiber.Ctx) error {
	role, ok := c.Locals("role").(domain.Role)
	if !ok || (role != domain.RoleTeacher && role != domain.RoleAdmin) {
		return response.Error(c, fiber.StatusForbidden, "FORBIDDEN", "only teachers can start whiteboard sessions")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	var req struct {
		ClassID int64  `json:"class_id"`
		Title   string `json:"title"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	session, err := h.service.StartSession(c.Context(), userID, req.ClassID, req.Title)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_FAILED", err.Error())
	}

	return response.Success(c, session)
}

func (h *WhiteboardHandler) EndSession(c *fiber.Ctx) error {
	role, ok := c.Locals("role").(domain.Role)
	if !ok || (role != domain.RoleTeacher && role != domain.RoleAdmin) {
		return response.Error(c, fiber.StatusForbidden, "FORBIDDEN", "only teachers can end whiteboard sessions")
	}

	var req struct {
		SessionID uuid.UUID   `json:"session_id"`
		Snapshot  interface{} `json:"snapshot"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	err := h.service.EndSession(c.Context(), req.SessionID, req.Snapshot)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UPDATE_FAILED", err.Error())
	}

	// Broadcast session end to all students in the class
	session, _ := h.service.GetSession(c.Context(), req.SessionID)
	if session != nil {
		h.broadcast(session.ClassID, "", websocket.TextMessage, []byte(`{"type":"SESSION_ENDED"}`))
	}

	return response.Success(c, fiber.Map{"status": "success"})
}

func (h *WhiteboardHandler) GetActiveSession(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}
	
	role, _ := c.Locals("role").(domain.Role)
	
	session, err := h.service.GetActiveSession(c.Context(), userID, role)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error())
	}
	if session == nil {
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "no active session found")
	}

	return response.Success(c, session)
}

func (h *WhiteboardHandler) GetSession(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid session id")
	}

	session, err := h.service.GetSession(c.Context(), sessionID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error())
	}
	if session == nil {
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "session not found")
	}

	return response.Success(c, session)
}

func (h *WhiteboardHandler) GetHistory(c *fiber.Ctx) error {
	classID, _ := c.ParamsInt("classId")
	
	sessions, err := h.service.GetHistory(c.Context(), int64(classID))
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error())
	}

	return response.Success(c, sessions)
}

func (h *WhiteboardHandler) GetStudentHistory(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	sessions, err := h.service.GetStudentHistory(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error())
	}

	return response.Success(c, sessions)
}

func (h *WhiteboardHandler) GetSummary(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session id"})
	}

	summary, err := h.service.GetSummary(c.Context(), sessionID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error())
	}
	if summary == nil {
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "summary not found or still generating")
	}

	return response.Success(c, summary)
}

func (h *WhiteboardHandler) WebSocketHandler(c *websocket.Conn) {
	classIDStr := c.Params("classId")
	var classID int64
	fmt.Sscanf(classIDStr, "%d", &classID)
	
	connID := uuid.New().String()

	// Register connection in hub
	hubMutex.Lock()
	if hub[classID] == nil {
		hub[classID] = make(map[string]*websocket.Conn)
	}
	hub[classID][connID] = c
	hubMutex.Unlock()

	log.Printf("[WS] Connection %s joined class %d", connID, classID)

	defer func() {
		hubMutex.Lock()
		delete(hub[classID], connID)
		if len(hub[classID]) == 0 {
			delete(hub, classID)
		}
		hubMutex.Unlock()
		c.Close()
		log.Printf("[WS] Connection %s left class %d", connID, classID)
	}()

	for {
		messageType, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		// Broadcast to all other connections in the same class
		h.broadcast(classID, connID, messageType, msg)
	}
}

func (h *WhiteboardHandler) broadcast(classID int64, senderID string, messageType int, message []byte) {
	hubMutex.RLock()
	defer hubMutex.RUnlock()

	for id, conn := range hub[classID] {
		if id == senderID {
			continue // Don't send back to sender
		}
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("[WS] Error broadcasting to %s: %v", id, err)
		}
	}
}
