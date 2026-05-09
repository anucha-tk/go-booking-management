-- name: CreateUser :one
INSERT INTO users (
    email, password_hash, role
) VALUES (
    $1, $2, $3
)
RETURNING id, email, password_hash, role, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, role, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, role, created_at, updated_at
FROM users
WHERE id = $1;

-- name: RevokeToken :exec
INSERT INTO revoked_tokens (
    jti, expires_at
) VALUES (
    $1, $2
);

-- name: IsTokenRevoked :one
SELECT EXISTS (
    SELECT 1 FROM revoked_tokens WHERE jti = $1
);

-- name: CleanupExpiredTokens :exec
DELETE FROM revoked_tokens WHERE expires_at < NOW();

