package endpoints

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/kennep/timelapse/domain"
)

type (
	// ApplicationContextHandler is the actual handler used for setting up the application context
	ApplicationContextHandler struct {
		handler http.Handler
	}

	appContextKeyType struct{}
)

var appContextKey appContextKeyType

// ApplicationContext creates the application context handler
func ApplicationContext() func(http.Handler) http.Handler {
	return applicationContextHandler
}

func applicationContextHandler(next http.Handler) http.Handler {
	return &ApplicationContextHandler{
		handler: next,
	}
}

// ServeHTTP sets up the application context and forward to the next handler
func (h *ApplicationContextHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	appctx := domain.ApplicationContext{}
	appctx.RequestID = uuid.New().String()

	ctx := context.WithValue(r.Context(), appContextKey, &appctx)
	h.handler.ServeHTTP(rw, r.WithContext(ctx))
}

// ApplicationContextFromRequest returns the application context assoicated with the request
// or nil if no application context has been set up.
func ApplicationContextFromRequest(r *http.Request) *domain.ApplicationContext {
	if appctx, ok := r.Context().Value(appContextKey).(*domain.ApplicationContext); ok {
		return appctx
	}
	return nil
}

// RequestFields returns a fields structure populated with fields from the current application context
func RequestFields(r *http.Request) log.Fields {
	fields := log.Fields{}

	PopulateRequestFields(fields, r)
	return fields
}

// PopulateRequestFields populates the given fields structure with fields from the current application context
func PopulateRequestFields(fields log.Fields, r *http.Request) {
	context := ApplicationContextFromRequest(r)

	if context != nil {
		fields["requestid"] = context.RequestID
		if context.User.SubjectID != "" {
			fields["user"] = fmt.Sprintf("%s/%s", context.User.SubjectID, context.User.Email)
		}
	}
}
