package domain

import "errors"

var (
	ErrNotFound                   = errors.New("not found")
	ErrUUIDInvalid                = errors.New("uuid is invalid")
	ErrInvalidAmount              = errors.New("amount must be greater than zero")
	ErrInvalidBalanceID           = errors.New("balance_id must be greater than zero")
	ErrDestinationRequired        = errors.New("destination is required")
	ErrIdempotencyKeyRequired     = errors.New("idempotency_key is required")
	ErrInvalidCredentials         = errors.New("invalid credentials")
	ErrUnauthorized               = errors.New("unauthorized")
	ErrInvalidTokenFormat         = errors.New("invalid authorization header format")
	ErrAccessTokenRequired        = errors.New("access token is required")
	ErrRefreshTokenRequired       = errors.New("refresh token is required")
	ErrInsufficientBalance        = errors.New("insufficient balance")
	ErrIdempotencyKeyAlreadyExist = errors.New("idempotency key already exist")
	ErrIdempotencyKeyPayloadDiff  = errors.New("idempotency key has different payload")
	ErrKeyNotFound                = errors.New("redis key not found")
	ErrInvalidFromDate            = errors.New("from must be valid RFC3339 datetime")
	ErrInvalidToDate              = errors.New("to must be valid RFC3339 datetime")
	ErrInvalidDateRange           = errors.New("from must be earlier than to")
	ErrInvalidLimit               = errors.New("limit must be in range [1..500]")
	ErrInvalidOffset              = errors.New("offset must be >= 0")
)
