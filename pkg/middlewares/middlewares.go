package middlewares

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/IsaacDSC/featureflag/pkg/authutils"
	"github.com/IsaacDSC/featureflag/pkg/ctxlog"
	"github.com/IsaacDSC/featureflag/pkg/ctxutils"
)

const (
	KEY              = "client"
	USERNAME_SERVICE = "SERVICE_CLIENT"
	USERNAME_SDK     = "SDK_CLIENT"
)

func getClientPermission(key string) (string, error) {
	cfg := env.Get()
	if client, ok := map[string]string{
		cfg.ServiceClientAT: USERNAME_SERVICE,
		cfg.SDKClientAT:     USERNAME_SDK,
	}[key]; ok {
		return client, nil
	}

	return "", errors.New("client not found")
}

func CheckPermission(h http.HandlerFunc, permission string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := ctxutils.GetValueCtx(r.Context(), KEY)
		if key == permission {
			h.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusForbidden)
	}
}

func Authorization(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		clientName, err := getClientPermission(authorization)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		request := r.WithContext(ctxutils.SetContext(r.Context(), KEY, clientName))

		w.Header().Set("Content-Type", "application/json")
		h.ServeHTTP(w, request)
	}
}

func Authentication(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := authutils.VerifyToken(cookie.Value); err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	}
}

// responseWriter é um wrapper para capturar o status code e o body da response
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// Flush implementa http.Flusher
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implementa http.Hijacker
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("http.Hijacker not supported")
}

// Push implementa http.Pusher
func (rw *responseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := rw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return errors.New("http.Pusher not supported")
}

// Logger middleware adiciona logger ao contexto e loga todas as responses
func Logger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Cria logger e adiciona ao contexto
		logger := ctxlog.NewLogger(r.Context())
		ctx := ctxlog.SetLogger(r.Context(), logger)

		// Adiciona informações da request ao logger
		logger = logger.With(
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)

		// Atualiza o contexto com o logger enriquecido
		ctx = ctxlog.SetLogger(ctx, logger)

		// Captura a response
		rw := newResponseWriter(w)

		// Lê o body da request se existir
		var requestBody string
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				// Restaura o body para ser lido novamente pelo handler
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Executa o handler com o contexto atualizado
		h.ServeHTTP(rw, r.WithContext(ctx))

		// Calcula duração
		duration := time.Since(start)

		// Loga a response
		logger.Info("HTTP Request",
			"status_code", rw.statusCode,
			"duration_ms", duration.Milliseconds(),
			"request_body", requestBody,
			"response_body", rw.body.String(),
		)
	}
}
