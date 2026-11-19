-- name: InsertPayerPayee :exec
INSERT INTO yaotose.payer_payee (
    tenant_id,
    payer_hash,
    payee_hash,
    key_id,
    total_amount,
    min_amount,
    max_amount,
    first_transaction,
    last_transaction
) VALUES (
    sqlc.arg('tenant_id')::text,
    sqlc.arg('payer_hash')::text,
    sqlc.arg('payee_hash')::text,
    sqlc.arg('key_id')::text,
    sqlc.arg('amount')::bigint,
    sqlc.arg('amount')::bigint,
    sqlc.arg('amount')::bigint,
    EXTRACT(EPOCH FROM now())::bigint,
    EXTRACT(EPOCH FROM now())::bigint
) ON CONFLICT (tenant_id, payer_hash, payee_hash) DO UPDATE SET
    n_payments = payer_payee.n_payments+1,
    total_amount = payer_payee.total_amount+EXCLUDED.total_amount,
    min_amount = LEAST(payer_payee.min_amount, EXCLUDED.min_amount),
    max_amount = GREATEST(payer_payee.max_amount, EXCLUDED.max_amount),
    last_transaction = EXTRACT(EPOCH FROM now())::bigint;

-- name: GetPayerPayee :one
SELECT n_payments, total_amount, min_amount, max_amount, first_transaction, last_transaction 
FROM yaotose.payer_payee 
WHERE tenant_id = sqlc.arg('tenant_id') AND payer_hash = sqlc.arg('payer_hash') AND payee_hash = sqlc.arg('payee_hash');
