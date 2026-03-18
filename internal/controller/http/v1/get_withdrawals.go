package v1

import (
	"net/http"

	"github.com/scaliann/withdrawal_service_test/pkg/render"
)

func (h *Handlers) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok := UserFromContext(ctx)
	if !ok {
		render.Error(w, nil, http.StatusUnauthorized, "Unauthorized")
		return
	}

	output, err := h.usecase.GetWithdrawals(ctx, user.UserID)
	if err != nil {
		apiErr := mapRequestError(err)
		render.Error(w, err, apiErr.status, apiErr.message)
		return
	}

	render.JSON(w, output.Withdrawals, http.StatusOK)
}
