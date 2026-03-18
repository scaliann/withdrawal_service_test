package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/scaliann/withdrawal_service_test/internal/domain"
	"github.com/scaliann/withdrawal_service_test/internal/dto"
)

const ttlCached = time.Minute * 1

func (u *UseCase) GetWithdrawal(ctx context.Context, input dto.GetWithdrawalInput) (dto.GetWithdrawalOutput, error) {
	var output dto.GetWithdrawalOutput // Пустой объект

	id, err := uuid.Parse(input.ID) // Парсим ID из запроса
	if err != nil {
		return output, domain.ErrUUIDInvalid
	}

	key := fmt.Sprintf("v1/withdrawals/%s", id)

	cached, err := u.redis.Get(ctx, key)
	if err == nil {
		var w domain.Withdrawal
		if err := json.Unmarshal(cached, &w); err == nil {
			log.Info().Msg("Withdrawal: Take from cache")
			return dto.GetWithdrawalOutput{Withdrawal: w}, nil
		}
		_ = u.redis.Del(ctx, key)
	}

	w, err := u.postgres.GetWithdrawal(ctx, id)
	if err != nil {
		return output, err
	}

	payload, err := json.Marshal(w)
	if err == nil {
		_ = u.redis.Set(ctx, key, payload, ttlCached)
	}
	log.Info().Msg("Withdrawal: Take from pg")
	return dto.GetWithdrawalOutput{Withdrawal: w}, nil
}
