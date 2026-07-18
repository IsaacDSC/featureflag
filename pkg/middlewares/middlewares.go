package middlewares

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
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
	EMAIL_KEY        = "current_user_email"

	// SessionCookieName é o cookie de sessão do login humano (distinto do
	// cookie legado "token" usado pelo fluxo SERVICE_CLIENT/SDK_CLIENT em
	// POST /auth), lido tanto aqui quanto por internal/auth ao emitir o JWT.
	SessionCookieName = "ff_session"

	serviceAccountEmail = "service-account"
	loginRedirectPath   = "/auth/login"
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

// extractEmail lê o claim "email" de dentro do payload genérico devolvido por
// authutils.GetDataJWT — depois do round-trip por JSON, um map[string]string
// passado a CreateToken sempre volta como map[string]any.
func extractEmail(raw any) (string, bool) {
	data, ok := raw.(map[string]any)
	if !ok {
		return "", false
	}

	email, ok := data["email"].(string)
	return email, ok
}

// redirectOrUnauthorized responde 401 para chamadas de API (fetch/XHR) e 302
// para /auth/login em navegação de página (o browser pedindo text/html).
func redirectOrUnauthorized(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		http.Redirect(w, r, loginRedirectPath, http.StatusFound)
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
}

// RequireLogin exige uma sessão de login humano válida (cookie ff_session
// contendo um JWT com claim "email") e propaga o email autenticado no
// contexto sob EMAIL_KEY.
func RequireLogin(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil {
			redirectOrUnauthorized(w, r)
			return
		}

		if err := authutils.VerifyToken(cookie.Value); err != nil {
			redirectOrUnauthorized(w, r)
			return
		}

		raw, err := authutils.GetDataJWT(cookie.Value)
		if err != nil {
			redirectOrUnauthorized(w, r)
			return
		}

		email, ok := extractEmail(raw)
		if !ok {
			redirectOrUnauthorized(w, r)
			return
		}

		ctx := ctxutils.SetContext(r.Context(), EMAIL_KEY, email)
		h.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequireServiceOrLogin aceita ou o token estático de serviço (automação,
// grava EMAIL_KEY="service-account" na trilha de auditoria) ou uma sessão de
// login humano válida (RequireLogin). Usado nas rotas administrativas que
// tanto automação quanto o dashboard precisam chamar.
func RequireServiceOrLogin(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if authorization := r.Header.Get("Authorization"); authorization != "" && authorization == env.Get().ServiceClientAT {
			ctx := ctxutils.SetContext(r.Context(), EMAIL_KEY, serviceAccountEmail)
			h.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		RequireLogin(h).ServeHTTP(w, r)
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

// noLogPrefixes lista os prefixos de rota cujas requisições não devem gerar
// o log "HTTP Request" — health check e o frontend do dashboard (assets
// estáticos, poll de navegação) geram ruído sem valor de observabilidade.
var noLogPrefixes = []string{"/ping", "/dashboard"}

func shouldSkipLog(path string) bool {
	for _, prefix := range noLogPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// Logger middleware adiciona logger ao contexto e loga todas as responses
func Logger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if shouldSkipLog(r.URL.Path) {
			h.ServeHTTP(w, r.WithContext(ctxlog.SetLogger(r.Context(), ctxlog.NewLogger(r.Context()))))
			return
		}

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
			"request_body", redactSensitiveFields(requestBody),
			"response_body", redactSensitiveFields(rw.body.String()),
		)
	}
}

var sensitiveJSONFields = []string{"password", "password_hash"}

// redactSensitiveFields tenta decodificar rawBody como um objeto JSON e substitui
// o valor de qualquer campo sensível de primeiro nível por um placeholder, antes
// de ir para o log. Corpos que não são um objeto JSON válido (ex.: vazio, ou uma
// rota sem payload sensível) voltam inalterados.
func redactSensitiveFields(rawBody string) string {
	if rawBody == "" {
		return rawBody
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(rawBody), &payload); err != nil {
		return rawBody
	}

	redacted := false
	for _, field := range sensitiveJSONFields {
		if _, ok := payload[field]; ok {
			payload[field] = "***REDACTED***"
			redacted = true
		}
	}
	if !redacted {
		return rawBody
	}

	out, err := json.Marshal(payload)
	if err != nil {
		return rawBody // fallback seguro: nunca panica por causa de log
	}

	return string(out)
}
