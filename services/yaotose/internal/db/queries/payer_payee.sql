-- name: InsertPayerPayee :exec
INSERT INTO yaotose.payer_payee (
    tenant_id,
    payer_hash,
    payee_hash,
    key_id,
    n_payments,
    total_amount,
    min_amount,
    max_amount,
    last_transaction
) VALUES (
    sqlc.arg('tenant_id')::text,
    sqlc.arg('payer_hash')::text,
    sqlc.arg('payee_hash')::text,
    sqlc.arg('key_id')::text,
    sqlc.arg('n_payments')::int,
    sqlc.arg('total_amount')::bigint,
    sqlc.arg('min_amount')::bigint,
    sqlc.arg('max_amount')::bigint,
    now()
);