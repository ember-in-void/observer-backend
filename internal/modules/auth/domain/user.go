// internal/modules/auth/domain/user.go
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// UserID - типизированный ID пользователя для строгой типизации
// Предотвращает случайное использование других строк как ID
type UserID string

// User - доменная сущность пользователя
// Содержит бизнес-логику и инварианты (правила которые всегда должны быть истинны)
type User struct {
	ID        UserID    // Уникальный идентификатор
	Email     *string   // Nullable: может быть не указан в Google профиле
	GoogleID  *string   // Google OAuth subject ID (уникальный)
	CreatedAt time.Time // Время создания записи
	UpdatedAt time.Time // Время последнего обновления
}

// NewUser - фабричный метод для создания нового пользователя
// Гарантирует что User создаётся в валидном состоянии
//
// Параметры:
//   - googleID: уникальный ID из Google OAuth (sub claim)
//   - email: опциональный email (может быть пустым если пользователь скрыл)
//
// Возвращает указатель на User с заполненными полями:
//   - Генерирует новый UUID для ID
//   - Устанавливает CreatedAt и UpdatedAt в текущее время
func NewUser(googleID string, email *string) *User {
	now := time.Now()

	// uuid.New() генерирует UUID v4 (случайный)
	// String() конвертирует UUID в строку формата: "550e8400-e29b-41d4-a716-446655440000"
	id := UserID(uuid.New().String())

	return &User{
		ID:        id,
		GoogleID:  &googleID,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate - валидация бизнес-правил
// Проверяет инварианты доменной модели (не технические проверки формата!)
//
// Бизнес-правила:
//  1. Google ID обязателен (мы авторизуем только через Google)
//  2. Email опционален (пользователь может скрыть в настройках Google)
//
// Возвращает ошибку если нарушены бизнес-правила
func (u *User) Validate() error {
	// Проверяем что GoogleID не nil и не пустая строка
	if u.GoogleID == nil || *u.GoogleID == "" {
		return errors.New("google ID is required")
	}

	// Email НЕ проверяем - он опционален по бизнес-логике
	// Если бы требовался, добавили бы проверку здесь

	return nil
}

// UpdateEmail - обновляет email пользователя
// Также обновляет UpdatedAt timestamp
//
// Это domain method (бизнес-операция), а не просто setter
// Инкапсулирует логику обновления: email + timestamp
func (u *User) UpdateEmail(email string) {
	u.Email = &email
	u.UpdatedAt = time.Now() // Автоматически обновляем timestamp
}

// String - имплементация Stringer interface для удобного логирования
// Возвращает строковое представление User без чувствительных данных
func (u *User) String() string {
	emailStr := "<nil>"
	if u.Email != nil {
		emailStr = *u.Email
	}

	return "User{ID=" + string(u.ID) + ", Email=" + emailStr + "}"
}
