-- name: CreateRoom :one
INSERT INTO rooms (
    room_number, type, price, status
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetRoom :one
SELECT * FROM rooms
WHERE id = $1 LIMIT 1;

-- name: GetRoomByNumber :one
SELECT * FROM rooms
WHERE room_number = $1 LIMIT 1;

-- name: ListRooms :many
SELECT * FROM rooms
WHERE 
    (sqlc.narg('type')::text IS NULL OR type = sqlc.narg('type')) AND
    (sqlc.narg('min_price')::bigint IS NULL OR price >= sqlc.narg('min_price')) AND
    (sqlc.narg('max_price')::bigint IS NULL OR price <= sqlc.narg('max_price'))
ORDER BY room_number
LIMIT $1 OFFSET $2;

-- name: UpdateRoom :one
UPDATE rooms
SET 
    room_number = $2,
    type = $3,
    price = $4,
    status = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteRoom :exec
DELETE FROM rooms
WHERE id = $1;

-- name: UpdateRoomStatus :one
UPDATE rooms
SET 
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ListAvailableRooms :many
SELECT * FROM rooms
WHERE 
    status = 'available' AND
    (sqlc.narg('type')::text IS NULL OR type = sqlc.narg('type')) AND
    (sqlc.narg('min_price')::bigint IS NULL OR price >= sqlc.narg('min_price')) AND
    (sqlc.narg('max_price')::bigint IS NULL OR price <= sqlc.narg('max_price')) AND
    id NOT IN (
        SELECT room_id FROM bookings
        WHERE status = 'confirmed'
        AND NOT (end_date <= @start_date OR start_date >= @end_date)
    )
ORDER BY room_number
LIMIT @limit_count OFFSET @offset_count;
