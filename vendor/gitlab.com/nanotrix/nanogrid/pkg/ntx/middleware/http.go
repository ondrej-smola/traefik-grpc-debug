package middleware

import (
	"net/http"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/httputil"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/handlers"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth/perm"
)

const (
	ServerTimeHeader = "server-time"
)

var allowedCorsMethods = []string{"GET", "HEAD", "POST", "PUT", "DELETE", "LOCK", "UNLOCK", "LINK", "UNLINK", "PATCH", "PURGE"}
var allowedCorsHeaders = []string{"Content-Type", "Authorization", auth.NtxTokenHeaderName}
var exposedCorsHeaders = []string{ServerTimeHeader, "Transfer-Encoding"}

type (
	Http func(http.Handler) http.Handler
)

func Combine(middleware ...Http) Http {
	return Http(func(next http.Handler) http.Handler {
		res := next
		for i := len(middleware) - 1; i >= 0; i-- {
			res = middleware[i](res)
		}
		return res
	})
}

func CORS() Http {
	return Http(
		handlers.CORS(
			handlers.AllowedMethods(allowedCorsMethods),
			handlers.AllowedHeaders(allowedCorsHeaders),
			handlers.ExposedHeaders(exposedCorsHeaders)))
}

func None() Http {
	return Combine()
}

type logWriter struct {
	logFn func(msg string)
}

func (g *logWriter) Write(p []byte) (int, error) {
	g.logFn(string(p))
	return len(p), nil
}

func Logging(l log.Logger) Http {
	return Http(func(next http.Handler) http.Handler {
		return handlers.LoggingHandler(&logWriter{logFn: func(msg string) {
			l.Log("msg", msg, "debug", true)
		}}, next)
	})
}

func NtxTokenPermissions(permFn func(permissions perm.Permissions) error) Http {
	return Http(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tkn, ok := auth.GetNtxTokenFromContext(r.Context())
			if !ok {
				httputil.WriteError(w, auth.ErrMissingNtxToken, r.Header.Get(httputil.AcceptHeader), http.StatusUnauthorized)
			} else if err := permFn(perm.Parse(tkn.Permissions...)); err != nil {
				httputil.WriteError(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusUnauthorized)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	})
}

func NtxTokenIsSystemAdmin() Http {
	return NtxTokenPermissions(func(p perm.Permissions) error {
		if !p.SystemAdmin() {
			return errors.New("permission required: " + perm.SystemAdmin())
		} else {
			return nil
		}
	})
}

func NtxTokenUseService(id string) Http {
	return NtxTokenPermissions(func(p perm.Permissions) error {
		if !p.UseService(id) {
			return errors.New("not allowed to use: " + id)
		} else {
			return nil
		}
	})
}

func NtxTokenIsSystemMonitor() Http {
	return NtxTokenPermissions(func(p perm.Permissions) error {
		if !p.SystemMonitor() {
			return errors.New("role required: " + perm.SystemMonitor())
		} else {
			return nil
		}
	})
}

func NtxTokenFromHttpHeaders(ntxDec auth.Decoder) Http {
	return Http(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ntxTokenString := r.Header.Get(auth.NtxTokenHeaderName)

			if ntxTokenString == "" {
				if authHeader := r.Header.Get("Authorization"); authHeader != "" {
					if tknS, err := auth.TokenStringFromAuthorizationHeader(authHeader); err != nil {
						httputil.WriteError(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusUnauthorized)
					} else {
						ntxTokenString = tknS
					}
				}
			}

			if ntxTokenString == "" {
				httputil.WriteError(w, auth.ErrMissingNtxToken, r.Header.Get(httputil.AcceptHeader), http.StatusUnauthorized)
			} else if ntxToken, err := ntxTokenFromString(ntxTokenString, ntxDec); err != nil {
				httputil.WriteError(w, errors.Wrap(err, "ntx-token"), r.Header.Get(httputil.AcceptHeader), http.StatusUnauthorized)
			} else {
				next.ServeHTTP(w, r.WithContext(auth.SetNtxTokenForContext(r.Context(), ntxToken)))
			}
		})
	})
}

func ntxTokenFromString(ntxTokenString string, ntxDec auth.Decoder) (*auth.NtxToken, error) {
	ntxToken, err := ntxDec.Decode(ntxTokenString)
	if err != nil {
		return nil, err
	}

	if err := ntxToken.Valid(); err != nil {
		return nil, err
	}

	return ntxToken, nil
}
