-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeed :one
SELECT * FROM feeds
WHERE url = $1;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetUserFeeds :many
SELECT * FROM feeds
WHERE user_id = $1;

-- name: GetFeedUser :one
SELECT name FROM users
WHERE id = (SELECT user_id FROM feeds
WHERE url = $1);