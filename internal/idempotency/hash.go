package idempotency

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/scaliann/withdrawal_service_test/internal/dto"
)

func BuildPayloadHash(p dto.CreateWithdrawalInput) string {
	raw := fmt.Sprintf("balance_id=%d|amount=%d|destination=%s", p.BalanceID, p.Amount, p.Destination)
	sum := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(sum[:])
}
