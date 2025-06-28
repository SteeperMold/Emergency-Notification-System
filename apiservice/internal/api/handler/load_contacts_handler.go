package handler

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/contextkeys"
	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/domain"
	"go.uber.org/zap"
)

// LoadContactsHandler handles multipart file uploads for loading contacts into S3
// and enqueuing background processing tasks via Kafka.
type LoadContactsHandler struct {
	service        domain.LoadContactsService
	logger         *zap.Logger
	contextTimeout time.Duration
}

// NewLoadContactsHandler constructs a LoadContactsHandler.
func NewLoadContactsHandler(s domain.LoadContactsService, logger *zap.Logger, contextTimeout time.Duration) *LoadContactsHandler {
	return &LoadContactsHandler{
		service:        s,
		logger:         logger,
		contextTimeout: contextTimeout,
	}
}

func (lch *LoadContactsHandler) logError(msg string, r *http.Request, fields ...zap.Field) {
	cid := r.Header.Get("X-Correlation-ID")

	allFields := []zap.Field{
		zap.String("correlation_id", cid),
		zap.String("uri", r.RequestURI),
		zap.String("client_ip", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
	}
	allFields = append(allFields, fields...)

	lch.logger.Error(msg, allFields...)
}

// LoadContactsFile handles POST /load-contacts requests with a multipart "file" field.
func (lch *LoadContactsHandler) LoadContactsFile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), lch.contextTimeout)
	defer cancel()

	rawUserID := ctx.Value(contextkeys.UserID)
	userID, ok := rawUserID.(int)
	if !ok {
		lch.logError("userID context value is not int", r, zap.Any("user_id", rawUserID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 100<<20) // 100MB

	err := r.ParseMultipartForm(0)
	if err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			lch.logError("internal server error", r, zap.Error(err))
		}
	}(file)

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".csv" && ext != ".xlsx" {
		http.Error(w, "only csv and xlsx files allowed", http.StatusUnprocessableEntity)
		return
	}

	buf := make([]byte, 10e6)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])

	switch contentType {
	case "text/plain; charset=utf-8", "text/csv":
	case "application/zip":
	default:
		http.Error(w, "invalid file type", http.StatusUnprocessableEntity)
		return
	}

	seeker, ok := file.(io.Seeker)
	if !ok {
		lch.logError("internal server error", r, zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_, err = seeker.Seek(0, io.SeekStart)
	if err != nil {
		lch.logError("internal server error", r, zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = lch.service.ProcessUpload(ctx, userID, header.Filename, file)
	if err != nil {
		lch.logError("internal server error", r, zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
