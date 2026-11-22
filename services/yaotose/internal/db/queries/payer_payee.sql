-- name: GetPayerPayee :one
SELECT n_payments, total_amount, min_amount, max_amount, first_transaction, last_transaction 
FROM yaotose.payer_payee 
WHERE tenant_id = sqlc.arg('tenant_id') AND payer_hash = sqlc.arg('payer_hash') AND payee_hash = sqlc.arg('payee_hash');
