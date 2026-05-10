-- name: ListBookingsByRoom :many
SELECT id, room_id, start_date, end_date, status
FROM bookings
WHERE room_id = $1
ORDER BY start_date DESC
LIMIT $2;

-- name: CreateBooking :one
INSERT INTO bookings (
    user_id, room_id, start_date, end_date, total_price, status
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: CreateBookingSafe :one
INSERT INTO bookings (
    user_id, room_id, start_date, end_date, total_price, status
)
SELECT $1, $2, $3, $4, $5, $6
WHERE NOT EXISTS (
    SELECT 1 FROM bookings
    WHERE room_id = $2
    AND status = 'confirmed'
    AND (
        (start_date, end_date) OVERLAPS ($3, $4)
    )
)
RETURNING *;

-- name: ListBookingsByUser :many
SELECT * FROM bookings
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetBooking :one
SELECT * FROM bookings
WHERE id = $1;

-- name: UpdateBookingStatus :one
UPDATE bookings
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ListAllBookings :many
SELECT b.*, u.email as user_email, r.room_number, r.type as room_type
FROM bookings b
JOIN users u ON b.user_id = u.id
JOIN rooms r ON b.room_id = r.id
ORDER BY b.created_at DESC;
