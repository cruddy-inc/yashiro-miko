-- name: InsertBlacklist :exec
INSERT INTO yaotose.blacklist (
    tenant_id,
    iban_hash,
    key_id,
    reason,
    ai_message,
    expires_at
) VALUES (
    sqlc.arg('tenant_id')::text,
    sqlc.arg('iban_hash')::text,
    sqlc.arg('key_id')::text,
    sqlc.arg('reason')::text[],
    sqlc.arg('ai_message')::text,
    now() + interval '24 hours'
) ON CONFLICT (tenant_id, iban_hash) DO UPDATE SET 
    key_id     = EXCLUDED.key_id,
    reason     = EXCLUDED.reason,
    ai_message = EXCLUDED.ai_message,
    expires_at = GREATEST(blacklist.expires_at, EXCLUDED.expires_at);

-- name: GetBlacklist :one
SELECT EXISTS (
    SELECT 1 FROM yaotose.blacklist WHERE iban_hash = sqlc.arg('iban_hash') AND tenant_id = sqlc.arg('tenant_id') AND expires_at > NOW()
);