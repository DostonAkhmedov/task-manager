package util

import (
	"context"
	"net/http"
)

type claimsContextKey struct{}
type userIDContextKey struct{}

var (
	claimsKey = claimsContextKey{}
	userIDKey = userIDContextKey{}
)

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(r *http.Request) string {
	claims, ok := r.Context().Value(claimsKey).(*Claims)
	if !ok {
		return ""
	}
	return claims.UserID
}

// AddClaimsToContext adds claims to the context
func AddClaimsToContext(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// AddClaimsAndUserIDToContext adds both claims and user_id to the context
// This is used in auth middleware to prepare the context for authenticated requests
func AddClaimsAndUserIDToContext(ctx context.Context, claims *Claims) context.Context {
	ctx = context.WithValue(ctx, claimsKey, claims)
	ctx = context.WithValue(ctx, userIDKey, claims.UserID)
	return ctx
}

// GetClaimsFromContext extracts claims from request context
func GetClaimsFromContext(r *http.Request) *Claims {
	claims, ok := r.Context().Value(claimsKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}
