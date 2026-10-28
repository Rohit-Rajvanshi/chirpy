-- name: GetUserFromRefreshToken :one 

SELECT users.* FROM users 
INNER JOIN refresh_tokens ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.user_id = $1 ;

