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
) ON CONFLICT DO NOTHING;