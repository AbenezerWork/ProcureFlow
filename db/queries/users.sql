-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash,
    full_name
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE lower(email) = lower($1);

-- name: UpdateUserLastLogin :exec
UPDATE users
SET
    last_login_at = now(),
    updated_at = now()
WHERE id = $1;

-- name: ListUserOrganizations :many
SELECT
    o.*,
    m.role AS membership_role,
    m.status AS membership_status
FROM organizations AS o
JOIN organization_memberships AS m
    ON m.organization_id = o.id
WHERE m.user_id = $1
  AND m.status = 'active'
ORDER BY o.name;
