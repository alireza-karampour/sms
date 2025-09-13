-- name: AddPhoneNumber :exec
INSERT INTO
    phone_numbers (user_id, phone_number)
VALUES ($1, $2);

-- name: AddPhoneNumberByUsername :exec
INSERT INTO
    phone_numbers (user_id, phone_number)
VALUES (
        (
            SELECT id
            FROM users
            WHERE
                username = $1
        ),
        $2
    );

-- name: GetPhoneNumber :one
SELECT id, user_id, phone_number FROM phone_numbers WHERE id = $1;

-- name: DeletePhoneNumber :one
DELETE FROM phone_numbers WHERE id = $1 RETURNING id;

-- name: GetPhoneNumbersByUsername :many
SELECT pn.id, pn.user_id, pn.phone_number
FROM phone_numbers pn
    JOIN users u ON pn.user_id = u.id
WHERE
    u.username = $1;

-- name: AddUser :exec
INSERT INTO users (username, balance) VALUES ($1, $2);

-- name: AddBalance :one
UPDATE users
SET
    balance = balance + $1
WHERE
    username = $2
RETURNING
    balance;

-- name: GetUserId :one
SELECT id FROM users u WHERE u.username = $1;

-- name: AddSms :exec
INSERT INTO sms (user_id,phone_number_id,to_phone_number,status,message) VALUES ($1, $2, $3, $4, $5);


