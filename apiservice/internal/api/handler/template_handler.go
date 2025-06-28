package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/contextkeys"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/models"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// TemplateHandler handles HTTP requests related to message templates.
// It validates input, delegates logic to the TemplateService, and writes JSON responses.
type TemplateHandler struct {
	service        domain.TemplateService
	logger         *zap.Logger
	contextTimeout time.Duration
}

// NewTemplateHandler creates a new TemplateHandler with the provided service, logger, and timeout.
func NewTemplateHandler(s domain.TemplateService, logger *zap.Logger, timeout time.Duration) *TemplateHandler {
	return &TemplateHandler{
		service:        s,
		logger:         logger,
		contextTimeout: timeout,
	}
}

func (th *TemplateHandler) logError(msg string, r *http.Request, fields ...zap.Field) {
	cid := r.Header.Get("X-Correlation-ID")

	allFields := []zap.Field{
		zap.String("correlation_id", cid),
		zap.String("uri", r.RequestURI),
		zap.String("client_ip", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
	}
	allFields = append(allFields, fields...)

	th.logger.Error(msg, allFields...)
}

// Get retrieves all message templates for the authenticated user.
// Responds with JSON-encoded list of templates or a 500 error.
func (th *TemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), th.contextTimeout)
	defer cancel()

	rawUserID := ctx.Value(contextkeys.UserID)
	userID, ok := rawUserID.(int)
	if !ok {
		th.logError("userID context value is not int", r, zap.Any("user_id", rawUserID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	templates, err := th.service.GetTemplatesByUserID(ctx, userID)
	if err != nil {
		th.logError("internal server error", r, zap.Int("user_id", userID), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&templates)
	if err != nil {
		th.logError("failed to write json to client", r, zap.Int("user_id", userID), zap.Error(err))
	}
}

// GetByID retrieves a single message template by its ID for the authenticated user.
// Responds with JSON-encoded template or 404/500 if not found or error occurs.
func (th *TemplateHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), th.contextTimeout)
	defer cancel()

	rawUserID := ctx.Value(contextkeys.UserID)
	userID, ok := rawUserID.(int)
	if !ok {
		th.logError("userID context value is not int", r, zap.Any("user_id", rawUserID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	tmplIDStr := vars["id"]
	tmplID, err := strconv.Atoi(tmplIDStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	tmpl, err := th.service.GetTemplateByID(ctx, userID, tmplID)
	if err != nil {
		if errors.Is(err, domain.ErrTemplateNotExists) {
			http.Error(w, "template not exists", http.StatusNotFound)
		} else {
			th.logError("internal server error", r, zap.Int("user_id", userID), zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(tmpl)
	if err != nil {
		th.logError("failed to write json to client", r, zap.Int("user_id", userID), zap.Error(err))
	}
}

// Post creates a new message template for the authenticated user.
// Validates the body, responds with 201 and the new template, or 422/500 on error.
func (th *TemplateHandler) Post(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), th.contextTimeout)
	defer cancel()

	rawUserID := ctx.Value(contextkeys.UserID)
	userID, ok := rawUserID.(int)
	if !ok {
		th.logError("userID context value is not int", r, zap.Any("user_id", rawUserID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var req domain.PostTemplateRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	newTmpl := &models.Template{
		UserID: userID,
		Name:   req.Name,
		Body:   req.Body,
	}

	newTmpl, err = th.service.CreateTemplate(ctx, newTmpl)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidTemplate):
			http.Error(w, "invalid template", http.StatusUnprocessableEntity)
		case errors.Is(err, domain.ErrTemplateAlreadyExists):
			http.Error(w, "template already exists", http.StatusConflict)
		default:
			th.logError("internal server error", r, zap.String("template_body", req.Body), zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newTmpl)
	if err != nil {
		th.logError("failed to write json to client", r, zap.Error(err))
	}
}

// Put updates an existing message template by ID for the authenticated user.
// Validates input, responds with 200 and updated template or 400/422/404/500.
func (th *TemplateHandler) Put(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), th.contextTimeout)
	defer cancel()

	rawUserID := ctx.Value(contextkeys.UserID)
	userID, ok := rawUserID.(int)
	if !ok {
		th.logError("userID context value is not int", r, zap.Any("user_id", rawUserID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	tmplIDStr := vars["id"]
	tmplID, err := strconv.Atoi(tmplIDStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req domain.PutTemplateRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	updatedTmpl := &models.Template{
		UserID: userID,
		Name:   req.Name,
		Body:   req.Body,
	}

	updatedTmpl, err = th.service.UpdateTemplate(ctx, userID, tmplID, updatedTmpl)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidTemplate):
			http.Error(w, "invalid template", http.StatusUnprocessableEntity)
		case errors.Is(err, domain.ErrTemplateNotExists):
			http.Error(w, "template not exists", http.StatusNotFound)
		case errors.Is(err, domain.ErrTemplateAlreadyExists):
			http.Error(w, "template already exists", http.StatusConflict)
		default:
			th.logError("internal server error", r, zap.String("template_body", req.Body), zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(updatedTmpl)
	if err != nil {
		th.logError("failed to write json to client", r, zap.Error(err))
	}
}

// Delete removes a message template by ID for the authenticated user.
// Responds with 204 on success, 404 if not found, or 500 on error.
func (th *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), th.contextTimeout)
	defer cancel()

	rawUserID := ctx.Value(contextkeys.UserID)
	userID, ok := rawUserID.(int)
	if !ok {
		th.logError("userID context value is not int", r, zap.Any("user_id", rawUserID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	tmplIDStr := vars["id"]
	tmplID, err := strconv.Atoi(tmplIDStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = th.service.DeleteTemplate(ctx, userID, tmplID)
	if err != nil {
		if errors.Is(err, domain.ErrTemplateNotExists) {
			http.Error(w, "template not exists", http.StatusNotFound)
		} else {
			th.logError("internal server error", r, zap.Int("user_id", userID), zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
