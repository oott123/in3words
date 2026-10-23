// Package logging 封装 logrus，并把请求 ID 放进 context，方便请求相关的日志带上同一个 request_id。
package logging

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

// RequestIDHeader 是透传与回写请求 ID 所用的 HTTP 头。
const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

// Setup 按配置设置全局 logger 的级别与输出格式。format 为 "json" 时输出 JSON，否则输出文本。
func Setup(level, format string) error {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("parse log level: %w", err)
	}
	logrus.SetLevel(lvl)
	if format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	}
	return nil
}

// WithRequestID 返回携带请求 ID 的 context。
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestID 取出 context 中的请求 ID，没有时返回空字符串。
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// FromContext 返回带有 request_id 字段（如果有）的日志条目。
func FromContext(ctx context.Context) *logrus.Entry {
	entry := logrus.NewEntry(logrus.StandardLogger())
	if id := RequestID(ctx); id != "" {
		entry = entry.WithField("request_id", id)
	}
	return entry
}

// Middleware 为每个请求分配请求 ID（优先沿用上游传入的），并在请求结束后记录访问日志。
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" || len(id) > 64 {
			id = xid.New().String()
		}
		w.Header().Set(RequestIDHeader, id)
		ctx := WithRequestID(r.Context(), id)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r.WithContext(ctx))

		FromContext(ctx).WithFields(logrus.Fields{
			"method":   r.Method,
			"uri":      r.URL.RequestURI(),
			"status":   rec.status,
			"duration": time.Since(start).String(),
		}).Info("request")
	})
}

type statusRecorder struct {
	http.ResponseWriter

	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
