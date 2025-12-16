-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS public.users (
    id TEXT PRIMARY KEY,
    email TEXT,
    google_id TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индекс для быстрого поиска по google_id
CREATE INDEX IF NOT EXISTS idx_users_google_id ON public.users(google_id);

-- Индекс для поиска по email (если понадобится)
CREATE INDEX IF NOT EXISTS idx_users_email ON public.users(email) WHERE email IS NOT NULL;