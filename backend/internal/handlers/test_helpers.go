package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yourusername/folda-finances/internal/middleware"
)

// setUserIDContext sets the user ID in the request context for testing
func setUserIDContext(r *http.Request, userID uuid.UUID) context.Context {
	return context.WithValue(r.Context(), middleware.UserIDKey, userID)
}

// setRouteContext sets the chi route context for testing
func setRouteContext(r *http.Request, rctx *chi.Context) context.Context {
	return context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
}
