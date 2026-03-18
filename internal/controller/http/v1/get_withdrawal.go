package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/scaliann/withdrawal_service_test/internal/dto"
	"github.com/scaliann/withdrawal_service_test/pkg/render"
)

func (h *Handlers) GetWithdrawal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input := dto.GetWithdrawalInput{
		ID: chi.URLParam(r, "id"),
	}

	output, err := h.usecase.GetWithdrawal(ctx, input)
	if err != nil {
		apiErr := mapRequestError(err)
		render.Error(w, err, apiErr.status, apiErr.message)
		return
	}

	render.JSON(w, output.Withdrawal, http.StatusOK)
}
