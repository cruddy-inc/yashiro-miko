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
SELECT 
    CURRENT_DATE - a.opened_at::date AS account_age_days,
    c.is_pep,
    -- DEBIT =  account_id -> sending moneys to counterparty_iban
    -- CREDIT = counterparty_iban -> sending moneys to account_id
    COUNT(DISTINCT t.account_id) FILTER (WHERE t.direction = 'CREDIT') AS unique_inbound_senders,
    COUNT(DISTINCT t.account_id) FILTER (WHERE t.direction = 'DEBIT') AS unique_outbound_receivers,
    COALESCE(SUM(t.amount_minor) FILTER (WHERE t.direction = 'DEBIT'), 0) AS total_outbound_volume,
    COALESCE(SUM(t.amount_minor) FILTER (WHERE t.direction = 'CREDIT'), 0) AS total_inbound_volume,
    -- Simply, BECAUSE POSTGRES DOESN'T ALLOW using already made aliases, we need to fetch the full result by redundancy
    COALESCE(SUM(t.amount_minor) FILTER (WHERE t.direction = 'DEBIT'), 0) +
    COALESCE(SUM(t.amount_minor) FILTER (WHERE t.direction = 'CREDIT'), 0) AS total_inbound_outbound_volume
FROM bank.accounts AS a
INNER JOIN bank.customers AS c
    ON a.customer_id = c.customer_id
INNER JOIN bank.transactions AS t
    ON t.customer_id = c.customer_id
WHERE t.booked_at >= CURRENT_DATE - interval '1 day'
GROUP BY a.opened_at, c.is_pep;
