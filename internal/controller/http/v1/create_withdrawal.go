package v1

import (
	"encoding/json"
	"net/http"

	"github.com/scaliann/withdrawal_service_test/internal/dto"
	"github.com/scaliann/withdrawal_service_test/pkg/render"
)

func (h *Handlers) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input := dto.CreateWithdrawalInput{}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&input)
	if err != nil {
		render.Error(w, err, http.StatusBadRequest, "invalid request body")
		return
	}

	output, err := h.usecase.CreateWithdrawal(ctx, input)
	if err != nil {
		apiErr := mapRequestError(err)
		render.Error(w, err, apiErr.status, apiErr.message)
		return
	}

	render.JSON(w, output, http.StatusOK)
}
