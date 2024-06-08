-- Добавляем столбцы include_keywords и exclude_keywords, если они не существуют
ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS include_keywords JSONB,
    ADD COLUMN IF NOT EXISTS exclude_keywords JSONB;

