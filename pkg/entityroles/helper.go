package entityroles

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func DuplicateEntityIDMiddleware(next http.Handler, copyEntityIDKey string) http.Handler {
	var entityIDKey any = "entity_id"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entityID := chi.URLParam(r, copyEntityIDKey)
		ctx := context.WithValue(r.Context(), entityIDKey, entityID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
