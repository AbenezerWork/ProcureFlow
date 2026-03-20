-- name: CreateOrganization :one
INSERT INTO organizations (
    name,
    slug,
    created_by_user_id
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetOrganizationByID :one
SELECT *
FROM organizations
WHERE id = $1;

-- name: UpdateOrganization :one
UPDATE organizations
SET
    name = $2,
    slug = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetOrganizationBySlug :one
SELECT *
FROM organizations
WHERE lower(slug) = lower($1);

-- name: CreateOrganizationMembership :one
INSERT INTO organization_memberships (
    organization_id,
    user_id,
    role,
    status,
	created_by_user_id,
    activated_at
) VALUES (
    $1, $2, $3, $4, $5,
    CASE WHEN $4 = 'active'::membership_status THEN now() ELSE NULL END
)
RETURNING *;

-- name: GetOrganizationMembership :one
SELECT *
FROM organization_memberships
WHERE organization_id = $1
  AND user_id = $2;

-- name: ListOrganizationMemberships :many
SELECT *
FROM organization_memberships
WHERE organization_id = $1
ORDER BY created_at DESC;

-- name: UpdateMembershipRole :one
UPDATE organization_memberships
SET
    role = $3,
    updated_at = now()
WHERE organization_id = $1
  AND user_id = $2
RETURNING *;

-- name: UpdateMembershipStatus :one
UPDATE organization_memberships
SET
    status = $3,
    activated_at = CASE WHEN $3 = 'active'::membership_status AND activated_at IS NULL THEN now() ELSE activated_at END,
    suspended_at = CASE WHEN $3 = 'suspended'::membership_status THEN now() ELSE suspended_at END,
    removed_at = CASE WHEN $3 = 'removed'::membership_status THEN now() ELSE removed_at END,
    updated_at = now()
WHERE organization_id = $1
  AND user_id = $2
RETURNING *;
