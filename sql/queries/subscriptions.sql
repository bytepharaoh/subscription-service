-- name: CreateSubscription :one
INSERT INTO subscriptions (
    service_name,
    price,
    user_id,
    start_date,
    end_date
) VALUES (
    @service_name,
    @price,
    @user_id,
    @start_date,
    @end_date
)
RETURNING *;

-- name: GetSubscriptionByID :one
SELECT * FROM subscriptions
WHERE id = @id
LIMIT 1;

-- name: UpdateSubscription :one
UPDATE subscriptions
SET
    service_name = @service_name,
    price        = @price,
    start_date   = @start_date,
    end_date     = @end_date,
    updated_at   = NOW()
WHERE id = @id
RETURNING *;

-- name: DeleteSubscription :exec
DELETE FROM subscriptions
WHERE id = @id;

-- name: ListSubscriptions :many
SELECT * FROM subscriptions
WHERE
    (sqlc.narg('user_id')::uuid      IS NULL OR user_id      = sqlc.narg('user_id'))
    AND
    (sqlc.narg('service_name')::text IS NULL OR service_name = sqlc.narg('service_name'))
ORDER BY created_at DESC
LIMIT  sqlc.arg('limit_count')
OFFSET sqlc.arg('offset_count');

-- name: GetTotalCost :one
SELECT COALESCE(SUM(price), 0)::integer AS total
FROM subscriptions
WHERE
    (sqlc.narg('user_id')::uuid      IS NULL OR user_id      = sqlc.narg('user_id'))
    AND
    (sqlc.narg('service_name')::text IS NULL OR service_name = sqlc.narg('service_name'))
    AND start_date <= sqlc.arg('period_end')::date
    AND (end_date IS NULL OR end_date >= sqlc.arg('period_start')::date);