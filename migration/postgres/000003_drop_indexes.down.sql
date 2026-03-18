CREATE INDEX IF NOT EXISTS idx_withdrawals_balance_created_at
    ON withdrawals (balance_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_created_at
    ON idempotency_keys (created_at DESC);
