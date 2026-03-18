package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/scaliann/withdrawal_service_test/internal/domain"
	"github.com/scaliann/withdrawal_service_test/internal/dto"
)

const (
	defaultSummaryLimit  = 50
	defaultSummaryOffset = 0
	maxSummaryLimit      = 500
	defaultSummaryWindow = time.Hour * 24 * 30
)

func (u *UseCase) GetUserWithdrawalsSummary(
	ctx context.Context,
	input dto.GetUserWithdrawalsSummaryInput,
) (dto.GetUserWithdrawalsSummaryOutput, error) {
	from, to, err := parseSummaryDates(input.From, input.To)
	if err != nil {
		return dto.GetUserWithdrawalsSummaryOutput{}, err
	}

	limit, err := parseSummaryLimit(input.Limit)
	if err != nil {
		return dto.GetUserWithdrawalsSummaryOutput{}, err
	}

	offset, err := parseSummaryOffset(input.Offset)
	if err != nil {
		return dto.GetUserWithdrawalsSummaryOutput{}, err
	}

	items, err := u.postgres.GetUserWithdrawalsSummary(ctx, from, to, limit, offset)
	if err != nil {
		return dto.GetUserWithdrawalsSummaryOutput{}, fmt.Errorf("postgres.GetUserWithdrawalsSummary: %w", err)
	}

	return dto.NewUserWithdrawalsSummaryOutput(items), nil
}

func parseSummaryDates(fromRaw string, toRaw string) (time.Time, time.Time, error) {
	now := time.Now().UTC()

	to := now
	if strings.TrimSpace(toRaw) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(toRaw))
		if err != nil {
			return time.Time{}, time.Time{}, domain.ErrInvalidToDate
		}
		to = parsed.UTC()
	}

	from := to.Add(-defaultSummaryWindow)
	if strings.TrimSpace(fromRaw) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(fromRaw))
		if err != nil {
			return time.Time{}, time.Time{}, domain.ErrInvalidFromDate
		}
		from = parsed.UTC()
	}

	if !from.Before(to) {
		return time.Time{}, time.Time{}, domain.ErrInvalidDateRange
	}

	return from, to, nil
}

func parseSummaryLimit(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultSummaryLimit, nil
	}

	limit, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, domain.ErrInvalidLimit
	}
	if limit <= 0 || limit > maxSummaryLimit {
		return 0, domain.ErrInvalidLimit
	}

	return limit, nil
}

func parseSummaryOffset(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultSummaryOffset, nil
	}

	offset, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, domain.ErrInvalidOffset
	}
	if offset < 0 {
		return 0, domain.ErrInvalidOffset
	}

	return offset, nil
}
