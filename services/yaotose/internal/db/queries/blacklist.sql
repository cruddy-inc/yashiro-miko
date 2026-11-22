-- name: GetBlacklist :one
SELECT EXISTS (
    SELECT 1 FROM yaotose.blacklist 
    WHERE (iban_hash = sqlc.arg('curr_iban_hash') 
    OR iban_hash = sqlc.arg('prev_iban_hash')) 
    AND tenant_id = sqlc.arg('tenant_id') 
    AND expires_at > NOW()
);