-- Таблица заявок
CREATE TABLE claims (
                        id TEXT PRIMARY KEY,                         -- KSUID, строка
                        details JSONB NOT NULL,                      -- Детали заявки
                        created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Таблица изображений, связанных с заявкой
CREATE TABLE claim_images (
                              id SERIAL PRIMARY KEY,
                              claim_id TEXT NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
                              filename TEXT NOT NULL,                      -- Имя файла (UUID или оригинальное)
                              description TEXT,                            -- Описание изображения
                              type TEXT NOT NULL,                          -- Тип (например: "damage", "overview", "interior")
                              uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Индекс для ускорения поиска по claim_id
CREATE INDEX idx_claim_images_claim_id ON claim_images(claim_id);