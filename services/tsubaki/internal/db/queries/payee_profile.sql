/* REFERENCE (proto)
message PayeeProfile {
  string payee_identifier_hash = 1;
  int32 account_age_days = 2;
  bool is_pep = 3;

  // Inbound and outbound information
  int32 unique_inbound_senders_24h = 4;
  int32 unique_outbound_recipients_24h = 5;
  int64 total_inbound_volume_24h_minor = 6;
  int64 total_outbound_volume_24h_minor = 7;
  float outbound_to_inbound_ratio_24h = 8;
}
*/

-- name: GetPayeeProfile :one
WITH transactions AS (
    SELECT
        account_id,
        COUNT(DISTINCT counterparty_iban) FILTER (WHERE direction = 'CREDIT')
            AS unique_inbound_senders,
        COUNT(DISTINCT counterparty_iban) FILTER (WHERE direction = 'DEBIT')
            AS unique_outbound_receivers,
        SUM(amount_minor) FILTER (WHERE direction = 'DEBIT')
            AS total_outbound_volume,
        SUM(amount_minor) FILTER (WHERE direction = 'CREDIT')
            AS total_inbound_volume
    FROM bank.transactions
    WHERE booked_at >= CURRENT_DATE - INTERVAL '1 day'
        AND account_id = (
            SELECT acc.account_id
            FROM bank.accounts AS acc
            WHERE acc.iban = sqlc.arg('payee_iban')
        )
    GROUP BY account_id
)
SELECT
    (CURRENT_DATE - a.opened_at::date) AS account_age_days,
    c.is_pep,
    COALESCE(t.unique_inbound_senders, 0) AS unique_inbound_senders,
    COALESCE(t.unique_outbound_receivers, 0) AS unique_outbound_receivers,
    COALESCE(t.total_outbound_volume, 0) AS total_outbound_volume,
    COALESCE(t.total_inbound_volume, 0) AS total_inbound_volume,
    COALESCE(t.total_outbound_volume, 0) / NULLIF(COALESCE(t.total_inbound_volume, 0), 0) AS outbound_to_inbound_ratio
FROM bank.customers AS c
INNER JOIN bank.accounts AS a ON a.customer_id = c.customer_id
LEFT JOIN transactions AS t ON t.account_id = a.account_id
WHERE a.iban = sqlc.arg('payee_iban');
