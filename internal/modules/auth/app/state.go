// internal/modules/auth/app/state_store.go
package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

// StateStore - интерфейс для хранения OAuth state tokens
// Используется для защиты от CSRF атак через state parameter
//
// OAuth Flow с state:
// 1. User → /auth/google/login
// 2. Backend генерирует random state, сохраняет с redirect URL
// 3. Backend редиректит на Google с state в URL
// 4. Google редиректит обратно с тем же state
// 5. Backend проверяет что state совпадает → защита от CSRF
type StateStore interface {
	// Save - сохраняет state с metadata (redirect URL)
	// TTL нужен чтобы state не жил вечно (защита от replay attacks)
	Save(ctx context.Context, state string, redirectURL string, ttl time.Duration) error

	// Get - получает redirect URL по state и УДАЛЯЕТ запись (one-time use)
	// One-time use важен для security: state нельзя переиспользовать
	Get(ctx context.Context, state string) (string, error)
}

// inMemoryStateStore - простая in-memory реализация для MVP
// ⚠️  Ограничения:
// - Теряется при рестарте сервера (пользователь получит "state not found")
// - Не работает с несколькими инстансами (state на instance A, callback на instance B)
// - Нет persistence (не переживёт падение сервера)
//
// Для production используй Redis:
// - Distributed (работает с N инстансами)
// - Persistent (переживает рестарт)
// - TTL из коробки (автоматическое удаление)
type inMemoryStateStore struct {
	mu     sync.RWMutex          // Защита от race conditions
	states map[string]stateEntry // state → metadata
}

// stateEntry - данные сохранённые вместе с state
type stateEntry struct {
	redirectURL string    // URL куда редиректить после login
	expiresAt   time.Time // Время истечения
}

// NewInMemoryStateStore - создаёт in-memory state store
func NewInMemoryStateStore() StateStore {
	store := &inMemoryStateStore{
		states: make(map[string]stateEntry),
	}

	// Запускаем фоновую горутину для очистки expired states
	// Без cleanup память будет расти бесконечно!
	go store.cleanup()

	return store
}

// Save - сохраняет state в памяти
func (s *inMemoryStateStore) Save(ctx context.Context, state string, redirectURL string, ttl time.Duration) error {
	// RWMutex.Lock() - exclusive lock (блокирует и чтение и запись)
	// Используем Lock (не RLock) потому что ИЗМЕНЯЕМ map
	s.mu.Lock()
	defer s.mu.Unlock() // defer гарантирует unlock даже при panic

	s.states[state] = stateEntry{
		redirectURL: redirectURL,
		expiresAt:   time.Now().Add(ttl), // Текущее время + TTL
	}

	return nil
}

// Get - получает redirect URL и удаляет state (one-time use)
func (s *inMemoryStateStore) Get(ctx context.Context, state string) (string, error) {
	// Lock (не RLock) потому что УДАЛЯЕМ из map
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем существование state
	entry, exists := s.states[state]
	if !exists {
		return "", errors.New("state not found")
	}

	// Проверяем истёк ли state
	// time.Now().After(t) эквивалентно time.Now() > t
	if time.Now().After(entry.expiresAt) {
		// State expired - удаляем и возвращаем ошибку
		delete(s.states, state)
		return "", errors.New("state expired")
	}

	// Удаляем state после использования (one-time use)
	// Защита от replay attacks: нельзя переиспользовать тот же state
	delete(s.states, state)

	return entry.redirectURL, nil
}

// cleanup - фоновая горутина для удаления expired states
// Запускается при создании store (NewInMemoryStateStore)
func (s *inMemoryStateStore) cleanup() {
	// time.Ticker - отправляет события в channel с фиксированным интервалом
	// В отличие от time.Sleep не блокирует другие операции
	ticker := time.NewTicker(5 * time.Minute) // Проверка каждые 5 минут
	defer ticker.Stop()                       // Останавливаем ticker при выходе из горутины

	// Бесконечный цикл: range по channel блокирует пока не придёт событие
	for range ticker.C {
		s.mu.Lock() // Блокируем на время очистки

		now := time.Now()

		// Итерируемся по всем states и удаляем expired
		// Можно безопасно удалять во время итерации (в отличие от slice)
		for state, entry := range s.states {
			if now.After(entry.expiresAt) {
				delete(s.states, state)
			}
		}

		s.mu.Unlock()
	}
}

// generateSecureState - генерирует криптографически безопасный random state
// Используется для защиты от CSRF атак в OAuth flow
func generateSecureState() (string, error) {
	// crypto/rand.Read - криптографически безопасный генератор случайных чисел
	// Источник энтропии: /dev/urandom (Linux) или CryptGenRandom (Windows)
	//
	// ⚠️  НЕ ИСПОЛЬЗУЙ math/rand для security!
	// math/rand предсказуемый (можно восстановить seed):
	//   rand.Seed(time.Now().UnixNano()) // ПЛОХО!
	//   value := rand.Intn(100)
	//
	// crypto/rand непредсказуемый (истинная случайность):
	//   rand.Read(b) // ХОРОШО!

	b := make([]byte, 32) // 32 байта = 256 бит (достаточно для security)

	// Read заполняет slice случайными байтами
	// Возвращает (n, err):
	// - n: количество прочитанных байт (всегда len(b) при успехе)
	// - err: ошибка если не удалось прочитать (крайне редко)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// base64.URLEncoding - кодирует байты в URL-safe строку
	// Алфавит: A-Z, a-z, 0-9, -, _ (без + и / которые нужно экранировать в URL)
	//
	// Альтернативы:
	// - base64.StdEncoding: использует + и / (нужно экранировать)
	// - hex.EncodeToString: длиннее в 2 раза
	//
	// 32 байта → ~43 символа в base64 (4/3 ratio + padding)
	return base64.URLEncoding.EncodeToString(b), nil
}
