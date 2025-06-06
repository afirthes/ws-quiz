package store

import "github.com/redis/go-redis/v9"

type QuizRedisStore struct {
	rdb *redis.Client
}

func NewQuizRedisStore() *QuizRedisStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Адрес Redis
		Password: "",               // Если нет пароля, оставить пустым
		DB:       0,                // Используемая БД Redis
	})

	return &QuizRedisStore{rdb: rdb}
}
