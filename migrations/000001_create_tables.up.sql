CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ENUMs
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_status') THEN
        CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'failed', 'refunded');
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_currency') THEN
        CREATE TYPE payment_currency AS ENUM ('rub', 'usd', 'eur');
    END IF;
END $$;

-- Пользователи Telegram
CREATE TABLE IF NOT EXISTS telegram_users (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    telegram_id BIGINT NOT NULL UNIQUE,
    username    TEXT NOT NULL DEFAULT '',
    first_name  TEXT NOT NULL DEFAULT '',
    last_name   TEXT NOT NULL DEFAULT '',
    phone       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_telegram_users_telegram_id ON telegram_users(telegram_id);

-- Клиенты аренды (1:1 с telegram_users)
CREATE TABLE IF NOT EXISTS clients (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    telegram_user_id UUID NOT NULL UNIQUE REFERENCES telegram_users(id),
    notes            TEXT NOT NULL DEFAULT '',
    is_blocked       BOOLEAN NOT NULL DEFAULT FALSE,
    last_visit       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clients_telegram_user_id ON clients(telegram_user_id);

-- Сессии аренды
CREATE TABLE IF NOT EXISTS rents (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id      UUID NOT NULL REFERENCES clients(id),
    starts_at      TIMESTAMPTZ NOT NULL,
    ends_at        TIMESTAMPTZ,
    is_cancelled   BOOLEAN NOT NULL DEFAULT FALSE,
    is_paid        BOOLEAN NOT NULL DEFAULT FALSE,
    price_per_hour INT NOT NULL DEFAULT 0,
    paid_duration  INT NOT NULL DEFAULT 0,
    bonus_duration INT NOT NULL DEFAULT 0,
    door_code      TEXT NOT NULL DEFAULT '',
    type           TEXT NOT NULL CHECK (type IN ('dj', 'production')),
    paid_at        TIMESTAMPTZ,
    notes          TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rents_client_id ON rents(client_id);
CREATE INDEX IF NOT EXISTS idx_rents_starts_at ON rents(starts_at);

-- Каталог доп. товаров
CREATE TABLE IF NOT EXISTS products (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price       DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Попытки оплаты клиента
CREATE TABLE IF NOT EXISTS payments (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id          UUID NOT NULL REFERENCES clients(id),
    amount             INT NOT NULL,
    currency           payment_currency NOT NULL DEFAULT 'rub',
    status             payment_status NOT NULL DEFAULT 'pending',
    tribute_order_uuid TEXT NOT NULL DEFAULT '',
    payment_url        TEXT NOT NULL DEFAULT '',
    paid_at            TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_client_id ON payments(client_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_payments_tribute_order_uuid
    ON payments(tribute_order_uuid)
    WHERE tribute_order_uuid <> '';
