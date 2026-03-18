package v1

import (
	"errors"
	"net/http"

	"github.com/scaliann/withdrawal_service_test/internal/domain"
)

type requestError struct {
	status  int
	message string
}

func mapRequestError(err error) requestError {
	switch {
	case errors.Is(err, domain.ErrUUIDInvalid):
		return requestError{status: http.StatusBadRequest, message: "invalid id"}
	case errors.Is(err, domain.ErrInvalidAmount):
		return requestError{status: http.StatusBadRequest, message: domain.ErrInvalidAmount.Error()}
	case errors.Is(err, domain.ErrInvalidBalanceID):
		return requestError{status: http.StatusBadRequest, message: domain.ErrInvalidBalanceID.Error()}
	case errors.Is(err, domain.ErrDestinationRequired):
		return requestError{status: http.StatusBadRequest, message: domain.ErrDestinationRequired.Error()}
	case errors.Is(err, domain.ErrIdempotencyKeyRequired):
		return requestError{status: http.StatusBadRequest, message: domain.ErrIdempotencyKeyRequired.Error()}
	case errors.Is(err, domain.ErrNotFound):
		return requestError{status: http.StatusNotFound, message: "object not found"}
	case errors.Is(err, domain.ErrInsufficientBalance):
		return requestError{status: http.StatusConflict, message: domain.ErrInsufficientBalance.Error()}
	case errors.Is(err, domain.ErrIdempotencyKeyAlreadyExist):
		return requestError{status: http.StatusConflict, message: domain.ErrIdempotencyKeyAlreadyExist.Error()}
	case errors.Is(err, domain.ErrIdempotencyKeyPayloadDiff):
		return requestError{status: http.StatusUnprocessableEntity, message: domain.ErrIdempotencyKeyPayloadDiff.Error()}
	case errors.Is(err, domain.ErrInvalidCredentials):
		return requestError{status: http.StatusUnauthorized, message: domain.ErrInvalidCredentials.Error()}
	case errors.Is(err, domain.ErrUnauthorized):
		return requestError{status: http.StatusUnauthorized, message: domain.ErrUnauthorized.Error()}
	case errors.Is(err, domain.ErrInvalidTokenFormat):
		return requestError{status: http.StatusUnauthorized, message: domain.ErrInvalidTokenFormat.Error()}
	case errors.Is(err, domain.ErrAccessTokenRequired):
		return requestError{status: http.StatusUnauthorized, message: domain.ErrAccessTokenRequired.Error()}
	case errors.Is(err, domain.ErrRefreshTokenRequired):
		return requestError{status: http.StatusBadRequest, message: domain.ErrRefreshTokenRequired.Error()}
	case errors.Is(err, domain.ErrInvalidFromDate):
		return requestError{status: http.StatusBadRequest, message: domain.ErrInvalidFromDate.Error()}
	case errors.Is(err, domain.ErrInvalidToDate):
		return requestError{status: http.StatusBadRequest, message: domain.ErrInvalidToDate.Error()}
	case errors.Is(err, domain.ErrInvalidDateRange):
		return requestError{status: http.StatusBadRequest, message: domain.ErrInvalidDateRange.Error()}
	case errors.Is(err, domain.ErrInvalidLimit):
		return requestError{status: http.StatusBadRequest, message: domain.ErrInvalidLimit.Error()}
	case errors.Is(err, domain.ErrInvalidOffset):
		return requestError{status: http.StatusBadRequest, message: domain.ErrInvalidOffset.Error()}
	default:
		return requestError{status: http.StatusInternalServerError, message: "internal server error"}
	}
}
