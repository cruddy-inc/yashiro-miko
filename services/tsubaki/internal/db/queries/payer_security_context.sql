/* REFERENCE (proto)
message PayerSecurityContext {
  bool is_trusted_device = 1;
  bool is_proxy_session = 2;
  bool ip_country_mismatch = 3; // Session IP country != Customer registration country

  // Account Takeover (ATO) Signals
  bool phone_changed_last_72h = 4;
  bool email_changed_last_72h = 5;
  int32 failed_auth_attempts_24h = 6;
  bool password_reset_last_24h = 7;
}
*/

-- name: GetPayerSecurityContext :one
WITH target_id AS (
    SELECT customer_id
    FROM bank.accounts
    WHERE iban = sqlc.arg('payee_iban')
),
profile_changes AS (
    SELECT
        customer_id,
        BOOL_OR(field = 'PHONE') AS phone_changed,
        BOOL_OR(field = 'EMAIL') AS email_changed
    FROM bank.profile_changes
    WHERE changed_at >= CURRENT_DATE - interval '3 days'
      AND customer_id = (SELECT customer_id FROM target_id)
    GROUP BY customer_id
),
auth_events AS (
    SELECT
        customer_id,
        COUNT(*) FILTER (WHERE event_type IN ('LOGIN_FAILED', 'MFA_FAILED')) AS failed_auth_attempts,
        BOOL_OR(event_type = 'PASSWORD_RESET') AS password_reset
    FROM bank.auth_events
    WHERE occurred_at >= CURRENT_DATE - interval '1 day'
      AND customer_id = (SELECT customer_id FROM target_id)
    GROUP BY customer_id
),
devices AS (
    SELECT
        customer_id,
        BOOL_OR(is_trusted = TRUE) AS is_trusted
    FROM bank.devices
    WHERE customer_id = (SELECT customer_id FROM target_id)
    GROUP BY customer_id
),
sessions AS (
    SELECT
        customer_id,
        BOOL_OR(is_proxy = TRUE) AS is_proxy,
        MAX(ip_country) AS last_ip_country
    FROM bank.sessions
    WHERE customer_id = (SELECT customer_id FROM target_id)
    GROUP BY customer_id
)
SELECT
    COALESCE(d.is_trusted, FALSE) AS is_trusted,
    COALESCE(s.is_proxy, FALSE) AS is_proxy,
    COALESCE(c.country != s.last_ip_country, FALSE) AS ip_country_mismatch,
    COALESCE(p.phone_changed, FALSE) AS phone_changed,
    COALESCE(p.email_changed, FALSE) AS email_changed,
    COALESCE(a.failed_auth_attempts, 0) AS failed_auth_attempts,
    COALESCE(a.password_reset, FALSE) AS password_reset
FROM bank.customers AS c
LEFT JOIN devices AS d ON d.customer_id = c.customer_id
LEFT JOIN sessions AS s ON s.customer_id = c.customer_id
LEFT JOIN profile_changes AS p ON p.customer_id = c.customer_id
LEFT JOIN auth_events AS a ON a.customer_id = c.customer_id
WHERE c.customer_id = (SELECT customer_id FROM target_id);