/*
message PayeeNeighborhood {
  // Co-location triggers between Payer & Payee or wider network
  bool shares_device_with_payer = 2;
  bool shares_ip_subnet_with_payer = 3;
}

message CounterpartySummary {
  int64 total_volume_transferred_minor = 2;
  int32 transaction_count = 3;
}
*/

-- name: GetPayeeNeighborhood :one
WITH target_id AS (
    SELECT acc.customer_id, acc.account_id
    FROM bank.accounts AS acc
    WHERE acc.iban = sqlc.arg('payee_iban')
),
payer_id AS (
    SELECT acc.customer_id
    FROM bank.accounts AS acc
    WHERE acc.iban = sqlc.arg('payer_iban')
),
devices AS (
    SELECT
        COALESCE(BOOL_OR(payee_dev.device_id = payer_dev.device_id), FALSE)::boolean AS shares_device_with_payer
    FROM bank.devices AS payee_dev
    CROSS JOIN bank.devices AS payer_dev
    WHERE payee_dev.customer_id = (SELECT customer_id FROM target_id)
      AND payer_dev.customer_id = (SELECT customer_id FROM payer_id)
),
sessions AS (
    SELECT
        COALESCE(BOOL_OR(payee_sess.ip = payer_sess.ip), FALSE)::boolean AS shares_ip_subnet_with_payer
    FROM bank.sessions AS payee_sess
    CROSS JOIN bank.sessions AS payer_sess
    WHERE payee_sess.customer_id = (SELECT customer_id FROM target_id)
      AND payer_sess.customer_id = (SELECT customer_id FROM payer_id)
),
transactions AS (
    SELECT
        COALESCE(SUM(amount_minor), 0)::bigint AS total_volume_transferred_minor,
        COUNT(*) AS transaction_count
    FROM bank.transactions
    WHERE account_id = (SELECT account_id FROM target_id)
)
SELECT
    (SELECT shares_device_with_payer FROM devices) AS shares_device_with_payer,
    (SELECT shares_ip_subnet_with_payer FROM sessions) AS shares_ip_subnet_with_payer,
    t.total_volume_transferred_minor,
    t.transaction_count
FROM transactions AS t;
