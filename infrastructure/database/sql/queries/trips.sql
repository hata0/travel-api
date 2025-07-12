-- name: GetTrip :one
SELECT id, name, created_at, updated_at FROM trips
WHERE id = $1 LIMIT 1;

-- name: ListTrips :many
SELECT id, name, created_at, updated_at FROM trips;

-- name: CreateTrip :exec
INSERT INTO trips (
  id, name, created_at, updated_at
) VALUES (
  $1, $2, $3, $4
);

-- name: UpdateTrip :exec
UPDATE trips
SET
  name = $2,
  updated_at = $3
WHERE id = $1;

-- name: DeleteTrip :exec
DELETE FROM trips
WHERE id = $1;
