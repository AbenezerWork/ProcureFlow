-- name: CreateActivityLog :one
INSERT INTO activity_logs (
    organization_id,
    actor_user_id,
    entity_type,
    entity_id,
    action,
    summary,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: ListActivityLogsByEntity :many
SELECT *
FROM activity_logs
WHERE organization_id = $1
  AND entity_type = $2
  AND entity_id = $3
ORDER BY occurred_at DESC;
