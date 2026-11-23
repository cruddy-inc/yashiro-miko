-- name: InsertCasesOpened :execrows
INSERT INTO yaotose.cases_opened (
    case_id,
    kafka_topic,
    kafka_case,
    payload,
    status
) VALUES (
    sqlc.arg('case_id')::text,
    sqlc.arg('kafka_topic')::text,
    sqlc.arg('kafka_case')::bytea,
    sqlc.arg('payload')::bytea,
    sqlc.arg('status')::text
) ON CONFLICT (case_id) DO UPDATE SET 
    created_at = EXTRACT(EPOCH FROM now())::bigint WHERE cases_opened.created_at + 600 < EXTRACT(EPOCH FROM now())::bigint;

-- name: InsertCasesOpenedPublished :exec
UPDATE yaotose.cases_opened SET
    status = "SENT",
    published_at = EXTRACT(EPOCH FROM now())::bigint
WHERE case_id = sqlc.arg('case_id');