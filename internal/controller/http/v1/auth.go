package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/scaliann/withdrawal_service_test/internal/domain"
	"github.com/scaliann/withdrawal_service_test/internal/dto"
	"github.com/scaliann/withdrawal_service_test/pkg/render"
)

func (h *Handlers) IssueToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input := dto.IssueTokenInput{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&input)
	if err != nil {
		render.Error(w, err, http.StatusBadRequest, "invalid request body")

		return
	}

	output, err := h.usecase.IssueToken(ctx, input)
	if err != nil {
		apiErr := mapRequestError(err)
		render.Error(w, err, apiErr.status, apiErr.message)

		return
	}

	render.JSON(w, output, http.StatusOK)
}

func (h *Handlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input := dto.RefreshTokenInput{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&input)
	if err != nil {
		render.Error(w, err, http.StatusBadRequest, "invalid request body")

		return
	}

	output, err := h.usecase.RefreshToken(ctx, input)
	if err != nil {
		apiErr := mapRequestError(err)
		render.Error(w, err, apiErr.status, apiErr.message)

		return
	}

	render.JSON(w, output, http.StatusOK)
}

func (h *Handlers) VerifyToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	accessToken, err := parseBearerToken(r)
	if err != nil {
		apiErr := mapRequestError(err)
		render.Error(w, err, apiErr.status, apiErr.message)

		return
	}

	user, err := h.usecase.VerifyAccessToken(ctx, accessToken)
	if err != nil {
		apiErr := mapRequestError(err)
		render.Error(w, err, apiErr.status, apiErr.message)

		return
	}

	render.JSON(w, dto.VerifyTokenOutput{
		Valid:   true,
		Subject: user.Subject,
	}, http.StatusOK)
}

func parseBearerToken(r *http.Request) (string, error) {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if authorization == "" {
		return "", domain.ErrAccessTokenRequired
	}

	scheme, token, found := strings.Cut(authorization, " ")
	if !found {
		return "", domain.ErrInvalidTokenFormat
	}
	if !strings.EqualFold(strings.TrimSpace(scheme), "Bearer") {
		return "", domain.ErrInvalidTokenFormat
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", domain.ErrAccessTokenRequired
	}

	return token, nil
}
