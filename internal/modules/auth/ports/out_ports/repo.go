// internal/modules/auth/ports/out_ports/user_repository.go
package out_ports

import (
	"context"
	"errors"

	"steam-observer/internal/modules/auth/domain"
)

// ErrNotFound - стандартная ошибка "запись не найдена"
// Используется всеми репозиториями для консистентности
//
// Проверка через errors.Is():
//
//	user, err := repo.FindByID(ctx, id)
//	if errors.Is(err, out_ports.ErrNotFound) {
//	    // handle not found case
//	}
var ErrNotFound = errors.New("not found")

// UserRepository - интерфейс для работы с пользователями
// Определён в out_ports (domain layer), реализован в adapters/out (infrastructure)
//
// Это Dependency Inversion Principle:
//   - High-level module (domain) НЕ зависит от low-level (database)
//   - Оба зависят от абстракции (интерфейс)
//
// Методы следуют Repository Pattern:
//   - Find* - чтение данных
//   - Create - создание новой записи
//   - Update - обновление существующей
//   - Delete - удаление (пока не нужен)
type UserRepository interface {
	// FindByGoogleID - поиск пользователя по Google OAuth ID
	//
	// Параметры:
	//   - ctx: контекст для отмены операции и передачи deadline
	//   - googleID: уникальный ID из Google (поле "sub" в OAuth response)
	//
	// Возвращает:
	//   - *domain.User: найденный пользователь
	//   - error: ErrNotFound если не найден, другие ошибки при проблемах с БД
	//
	// Пример:
	//   user, err := repo.FindByGoogleID(ctx, "108123456789")
	//   if errors.Is(err, out_ports.ErrNotFound) {
	//       // create new user
	//   }
	FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error)

	// FindByID - поиск пользователя по внутреннему ID
	//
	// Параметры:
	//   - ctx: контекст
	//   - userID: UUID пользователя (domain.UserID)
	//
	// Возвращает:
	//   - *domain.User: найденный пользователь
	//   - error: ErrNotFound если не найден
	//
	// Используется для:
	//   - Загрузки профиля пользователя
	//   - Проверки существования при авторизации
	FindByID(ctx context.Context, userID domain.UserID) (*domain.User, error)

	// Create - создание нового пользователя в БД
	//
	// Параметры:
	//   - ctx: контекст
	//   - user: доменная сущность для сохранения
	//
	// Побочные эффекты:
	//   - Обновляет user.CreatedAt и user.UpdatedAt значениями из БД
	//   - Если ID генерируется БД (не наш случай), обновляет user.ID
	//
	// Возвращает:
	//   - error: ошибка если не удалось сохранить (constraint violation, connection error)
	//
	// Пример:
	//   user := domain.NewUser("google123", &email)
	//   err := repo.Create(ctx, user)
	//   // После успешного Create, user.CreatedAt содержит время из БД
	Create(ctx context.Context, user *domain.User) error

	// Update - обновление существующего пользователя
	//
	// Параметры:
	//   - ctx: контекст
	//   - user: доменная сущность с изменёнными полями
	//
	// Обновляемые поля:
	//   - Email
	//   - GoogleID (хотя обычно не меняется)
	//   - UpdatedAt (автоматически устанавливается в NOW())
	//
	// Возвращает:
	//   - error: ErrNotFound если пользователь не существует
	//           или другая ошибка при проблемах с БД
	//
	// Пример:
	//   user.UpdateEmail("new@email.com")
	//   err := repo.Update(ctx, user)
	Update(ctx context.Context, user *domain.User) error
}
