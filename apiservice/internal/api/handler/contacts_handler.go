package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SteeperMold/Emergency-Notification-System/internal/contextkeys"
	"github.com/SteeperMold/Emergency-Notification-System/internal/domain"
	"github.com/SteeperMold/Emergency-Notification-System/internal/models"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type ContactsHandler struct {
	service        domain.ContactsService
	logger         *zap.Logger
	contextTimeout time.Duration
}

func NewContactsHandler(s domain.ContactsService, logger *zap.Logger, timeout time.Duration) *ContactsHandler {
	return &ContactsHandler{
		service:        s,
		logger:         logger,
		contextTimeout: timeout,
	}
}

func (ch *ContactsHandler) logError(msg string, r *http.Request, fields ...zap.Field) {
	cid := r.Header.Get("X-Correlation-ID")

	allFields := []zap.Field{
		zap.String("correlation_id", cid),
		zap.String("uri", r.RequestURI),
		zap.String("client_ip", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
	}
	allFields = append(allFields, fields...)

	ch.logger.Error(msg, allFields...)
}

func (ch *ContactsHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), ch.contextTimeout)
	defer cancel()

	rawUserID := ctx.Value(contextkeys.UserID)
	userID, ok := rawUserID.(int)
	if !ok {
		ch.logError("userID context value is not int", r, zap.Any("user_id", rawUserID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	contacts, err := ch.service.GetContactsByUserID(ctx, userID)
	if err != nil {
		ch.logError("internal server error", r, zap.Int("user_id", userID), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&contacts)
	if err != nil {
		ch.logError("failed to write json to client", r, zap.Int("user_id", userID), zap.Error(err))
	}
}

func (ch *ContactsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), ch.contextTimeout)
	defer cancel()

	rawUserID := ctx.Value(contextkeys.UserID)
	userID, ok := rawUserID.(int)
	if !ok {
		ch.logError("userID context value is not int", r, zap.Any("user_id", rawUserID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	contactIDStr := vars["id"]
	contactID, err := strconv.Atoi(contactIDStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	contact, err := ch.service.GetContactByID(ctx, userID, contactID)
	if err != nil {
		if errors.Is(err, domain.ErrContactNotExists) {
			http.Error(w, "contact not exists", http.StatusNotFound)
		} else {
			ch.logError("internal server error", r, zap.Int("user_id", userID), zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(contact)
	if err != nil {
		ch.logError("failed to write json to client", r, zap.Int("user_id", userID), zap.Error(err))
	}
}

func (ch *ContactsHandler) Post(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), ch.contextTimeout)
	defer cancel()

	rawUserID := ctx.Value(contextkeys.UserID)
	userID, ok := rawUserID.(int)
	if !ok {
		ch.logError("userID context value is not int", r, zap.Any("user_id", rawUserID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var req domain.PostContactRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	newContact := &models.Contact{
		UserID: userID,
		Name:   req.Name,
		Phone:  req.Phone,
	}

	newContact, err = ch.service.CreateContact(ctx, newContact)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidContact):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case errors.Is(err, domain.ErrContactAlreadyExists):
			http.Error(w, "contact already exists", http.StatusConflict)
		default:
			ch.logError("internal server error", r, zap.String("contact_phone", req.Phone), zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newContact)
	if err != nil {
		ch.logError("failed to write json to client", r, zap.Error(err))
	}
}

func (ch *ContactsHandler) Put(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), ch.contextTimeout)
	defer cancel()

	rawUserID := ctx.Value(contextkeys.UserID)
	userID, ok := rawUserID.(int)
	if !ok {
		ch.logError("userID context value is not int", r, zap.Any("user_id", rawUserID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	contactIDStr := vars["id"]
	contactID, err := strconv.Atoi(contactIDStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req domain.PutContactRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	updatedContact := &models.Contact{
		UserID: userID,
		Name:   req.Name,
		Phone:  req.Phone,
	}

	updatedContact, err = ch.service.UpdateContact(ctx, userID, contactID, updatedContact)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidContact):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case errors.Is(err, domain.ErrContactNotExists):
			http.Error(w, "contact not exists", http.StatusNotFound)
		case errors.Is(err, domain.ErrContactAlreadyExists):
			http.Error(w, "contacts already exists", http.StatusConflict)
		default:
			ch.logError("internal server error", r, zap.String("contact_phone", req.Phone), zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(updatedContact)
	if err != nil {
		ch.logError("failed to write json to client", r, zap.Error(err))
	}
}

func (ch *ContactsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), ch.contextTimeout)
	defer cancel()

	rawUserID := ctx.Value(contextkeys.UserID)
	userID, ok := rawUserID.(int)
	if !ok {
		ch.logError("userID context value is not int", r, zap.Any("user_id", rawUserID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	contactIDStr := vars["id"]
	contactID, err := strconv.Atoi(contactIDStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = ch.service.DeleteContact(ctx, userID, contactID)
	if err != nil {
		if errors.Is(err, domain.ErrContactNotExists) {
			http.Error(w, "contact doesn't exist", http.StatusNotFound)
		} else {
			ch.logError("internal server error", r, zap.Int("user_id", userID), zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
