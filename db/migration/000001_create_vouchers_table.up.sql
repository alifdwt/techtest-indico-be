CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS vouchers (
    -- id, voucher_code, discount_percent, expiry_date, created_at, updated_at
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    voucher_code VARCHAR(255) UNIQUE NOT NULL,
    discount_percent INTEGER NOT NULL CHECK (discount_percent >= 0 AND discount_percent <= 100),
    expiry_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_voucher_code ON vouchers(voucher_code);
CREATE INDEX IF NOT EXISTS idx_expiry_date ON vouchers(expiry_date);