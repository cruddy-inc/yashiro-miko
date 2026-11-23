-- name: GetFailedCasesOpened :many
SELECT * FROM yaotose.cases_opened
WHERE status = "PENDING" AND published_at IS NOT NULL
LIMIT 500;

-- name: InsertCasesOpenedPublished :exec
UPDATE yaotose.cases_opened SET
    status = "SENT",
    published_at = EXTRACT(EPOCH FROM now())::bigint
WHERE case_id = sqlc.arg('case_id');