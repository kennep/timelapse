package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	oidc "github.com/coreos/go-oidc"
	log "github.com/sirupsen/logrus"

	"github.com/kennep/timelapse/repository"
)

const bearerLen = len("Bearer ")

var providerURLs = [...]string{"https://id.wangpedersen.com/auth/realms/wangpedersen", "https://accounts.google.com"}

type (
	// AuthenticationHandler is the actual handler used for authentication processing
	AuthenticationHandler struct {
		handler   http.Handler
		providers map[string]*oidc.Provider
		verifiers map[string]*oidc.IDTokenVerifier
		repo      *repository.TimelapseRepository
	}

	// The parts of the claims that the code cares about
	jwtClaims struct {
		Issuer        string `json:"iss"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}

	// This is returned in error responses
	errorResponse struct {
		Message string `json:"message"`
	}
)

func (h *AuthenticationHandler) oidcVerifiers() map[string]*oidc.IDTokenVerifier {
	if h.providers == nil {
		h.providers = make(map[string]*oidc.Provider)
		h.verifiers = make(map[string]*oidc.IDTokenVerifier)
		for _, providerURL := range providerURLs {
			ctx := context.Background()
			provider, err := oidc.NewProvider(ctx, providerURL)
			if err != nil {
				log.WithFields(log.Fields{"provider": provider, "error": err}).Error("Disabling authentication provider")
				continue
			}
			h.providers[providerURL] = provider

			verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})
			h.verifiers[providerURL] = verifier
		}
	}
	return h.verifiers
}

// Authentication creates a new handler component.
func Authentication(repo *repository.TimelapseRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &AuthenticationHandler{
			handler: next,
			repo:    repo,
		}
	}
}

func getIssuerFromJWT(p string) (string, error) {
	parts := strings.Split(p, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("malformed jwt, expected 3 parts got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("malformed jwt payload: %v", err)
	}

	var claims jwtClaims
	json.Unmarshal(payload, &claims)

	return claims.Issuer, nil
}

func emitErrorResponse(rw http.ResponseWriter, statusCode int, errorMessage string) {
	errResponse := errorResponse{errorMessage}
	body, err := json.Marshal(errResponse)
	if err != nil {
		// oh dear... errors when marshaling the error response. We'll write the message as plain text instead
		rw.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		rw.WriteHeader(statusCode)
		rw.Write([]byte(errorMessage))
	} else {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(statusCode)
		rw.Write(body)
	}
}

// ServeHTTP calls the next handler after authentication processing
func (h *AuthenticationHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	authenticated, err := h.authenticate(rw, r)
	if err != nil {
		log.Errorf("Authentication error: %s", err)
		emitErrorResponse(rw, 500, "Internal server error")
		return
	}
	if authenticated {
		h.handler.ServeHTTP(rw, r)
	} else {
		rw.Header().Add("WWW-Authenticate", "Bearer")
		emitErrorResponse(rw, 401, "Authorization needed")
	}
}

func (h *AuthenticationHandler) authenticate(rw http.ResponseWriter, r *http.Request) (bool, error) {
	appcontext := ApplicationContextFromRequest(r)
	if appcontext == nil {
		return false, errors.New("Application configuration error - application context not initialized")
	}

	authheader := r.Header.Get("authorization")
	if authheader != "" {
		// Bearer
		if len(authheader) > bearerLen && strings.ToLower(authheader[0:bearerLen]) == "bearer " {
			bearertoken := authheader[bearerLen:]

			issuer, err := getIssuerFromJWT(bearertoken)
			if err != nil {
				fields := RequestFields(r)
				fields["error"] = err
				log.WithFields(fields).Warn("could not parse issuer from token - ignoring")
				return false, nil
			}
			verifiers := h.oidcVerifiers()
			verifier := verifiers[issuer]

			if verifier != nil {
				idtoken, err := verifier.Verify(r.Context(), bearertoken)
				if err != nil {
					fields := RequestFields(r)
					fields["error"] = err
					log.WithFields(fields).Warn("Ignoring invalid token")
					return false, nil
				}
				var claims jwtClaims
				if err := idtoken.Claims(&claims); err != nil {
					fields := RequestFields(r)
					fields["error"] = err
					log.WithFields(fields).Warn("Ignoring token with missing claims")
					return false, nil
				}

				appcontext.User.SubjectID = idtoken.Subject
				appcontext.User.Issuer = idtoken.Issuer
				appcontext.User.Email = claims.Email

				_, err = h.repo.CreateUserFromContext(appcontext)
				if err != nil {
					return false, err
				}
				return true, nil
			}
		}
	}

	return false, nil

}
