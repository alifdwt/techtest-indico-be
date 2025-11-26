-- name: CreateVoucher :one
INSERT INTO vouchers (
    voucher_code,
    discount_percent,
    expiry_date
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetVoucherByID :one
SELECT * FROM vouchers WHERE id = $1 LIMIT 1;

-- name: GetVoucherByCode :one
SELECT * FROM vouchers WHERE voucher_code = $1 LIMIT 1;

-- name: ListVouchers :many
SELECT * FROM vouchers
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: UpdateVoucher :one
UPDATE vouchers SET
    voucher_code = $2,
    discount_percent = $3,
    expiry_date = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteVoucher :exec
DELETE FROM vouchers WHERE id = $1;

-- name: CountVouchers :one
SELECT COUNT(*) FROM vouchers;