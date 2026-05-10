-- name: ListBookingsByRoom :many
SELECT id, room_id, start_date, end_date, status
FROM bookings
WHERE room_id = $1
ORDER BY start_date DESC
LIMIT $2;
