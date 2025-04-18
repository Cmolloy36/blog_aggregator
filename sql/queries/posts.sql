-- name: CreatePost :one
INSERT INTO posts (ID, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $2,
    $3,
    $4,
    $5,
    $2,
    $6
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT * FROM posts
WHERE feed_id IN (
    SELECT feed_id FROM feeds
    WHERE user_id = $1
)
ORDER BY updated_at
LIMIT $2;