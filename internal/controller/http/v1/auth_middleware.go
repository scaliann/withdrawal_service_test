package v1

import (
	"context"
	"net/http"

	"github.com/scaliann/withdrawal_service_test/internal/usecase"
	"github.com/scaliann/withdrawal_service_test/pkg/render"
)

type AuthUserKey struct{}

func UserFromContext(ctx context.Context) (usecase.JwtClaims, bool) {
	user, ok := ctx.Value(AuthUserKey{}).(usecase.JwtClaims)
	return user, ok
}

func (h *Handlers) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := parseBearerToken(r)
		if err != nil {
			apiErr := mapRequestError(err)
			render.Error(w, err, apiErr.status, apiErr.message)

			return
		}

		user, err := h.usecase.VerifyAccessToken(r.Context(), accessToken)
		if err != nil {
			apiErr := mapRequestError(err)
			render.Error(w, err, apiErr.status, apiErr.message)

			return
		}

		ctx := context.WithValue(r.Context(), AuthUserKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
