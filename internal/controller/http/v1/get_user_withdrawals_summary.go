package v1

import (
	"net/http"

	"github.com/scaliann/withdrawal_service_test/internal/dto"
	"github.com/scaliann/withdrawal_service_test/pkg/render"
)

func (h *Handlers) GetUserWithdrawalsSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input := dto.GetUserWithdrawalsSummaryInput{
		From:   r.URL.Query().Get("from"),
		To:     r.URL.Query().Get("to"),
		Limit:  r.URL.Query().Get("limit"),
		Offset: r.URL.Query().Get("offset"),
	}

	output, err := h.usecase.GetUserWithdrawalsSummary(ctx, input)
	if err != nil {
		apiErr := mapRequestError(err)
		render.Error(w, err, apiErr.status, apiErr.message)
		return
	}

	render.JSON(w, output.Items, http.StatusOK)
}
