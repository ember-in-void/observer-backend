package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"steam-observer/internal/modules/auth/domain"
	"steam-observer/internal/modules/auth/ports/out_ports"
)

// userRepository - PostgreSQL реализация UserRepository
type userRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository - создаёт новый PostgreSQL репозиторий
// Принимает pgxpool.Pool который управляется в shared/db/db.go
func NewUserRepository(pool *pgxpool.Pool) out_ports.UserRepository {
	return &userRepository{pool: pool}
}

// FindByGoogleID - поиск пользователя по Google OAuth ID
func (r *userRepository) FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	// SQL запрос с именованными параметрами ($1, $2, ...)
	// pgx автоматически защищает от SQL injection при использовании параметров
	query := `
        SELECT id, email, google_id, created_at, updated_at
        FROM public.users
        WHERE google_id = $1
    `

	// QueryRow - выполняет запрос и ожидает РОВНО одну строку
	// Если строк 0 → вернёт pgx.ErrNoRows
	// Если строк >1 → вернёт только первую (но это ошибка дизайна БД!)
	//
	// Контекст (ctx) позволяет:
	// - Отменить запрос если ctx.Done() закрылся
	// - Установить timeout через context.WithTimeout
	// - Передать trace ID для distributed tracing
	row := r.pool.QueryRow(ctx, query, googleID)

	// Создаём пустую структуру для результата
	var user domain.User
	var email, dbGoogleID *string // nullable поля в БД → указатели в Go

	// Scan - копирует данные из row в переменные
	// ВАЖНО: порядок переменных ДОЛЖЕН совпадать с SELECT!
	//
	// Автоматическая конвертация типов PostgreSQL → Go:
	// - TEXT/VARCHAR → string
	// - TEXT (nullable) → *string
	// - INTEGER → int, int32, int64
	// - BOOLEAN → bool
	// - TIMESTAMP → time.Time
	// - UUID → string или github.com/google/uuid.UUID
	// - JSON/JSONB → []byte или struct с json tags
	err := row.Scan(
		&user.ID,        // TEXT → domain.UserID (type alias для string)
		&email,          // TEXT (nullable) → *string
		&dbGoogleID,     // TEXT (nullable) → *string
		&user.CreatedAt, // TIMESTAMP → time.Time
		&user.UpdatedAt, // TIMESTAMP → time.Time
	)
	if err != nil {
		// pgx.ErrNoRows - специальная ошибка означающая "запись не найдена"
		// Важно: НЕ проверяй err == pgx.ErrNoRows (не сработает с wrapped errors!)
		// Используй errors.Is() который поддерживает error wrapping (%w)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, out_ports.ErrNotFound
		}

		// Оборачиваем ошибку с контекстом для лучшего debugging
		// %w сохраняет оригинальную ошибку для errors.Is/errors.As
		return nil, fmt.Errorf("query user by google_id: %w", err)
	}

	// Присваиваем nullable поля
	user.Email = email
	user.GoogleID = dbGoogleID

	return &user, nil
}

// FindByID - поиск пользователя по внутреннему ID
func (r *userRepository) FindByID(ctx context.Context, userID domain.UserID) (*domain.User, error) {
	query := `
        SELECT id, email, google_id, created_at, updated_at
        FROM public.users
        WHERE id = $1
    `

	// Конвертируем domain.UserID (type alias) обратно в string для SQL
	row := r.pool.QueryRow(ctx, query, string(userID))

	var user domain.User
	var email, googleID *string

	err := row.Scan(
		&user.ID,
		&email,
		&googleID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, out_ports.ErrNotFound
		}
		return nil, fmt.Errorf("query user by id: %w", err)
	}

	user.Email = email
	user.GoogleID = googleID

	return &user, nil
}

// Create - создаёт нового пользователя в БД
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	// Валидация бизнес-правил перед сохранением
	// Защита от сохранения невалидных данных
	if err := user.Validate(); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	// INSERT с RETURNING - вставляет строку и сразу возвращает значения
	// Полезно для:
	// - Получения auto-generated ID (если не генерируем на клиенте)
	// - Получения server-side timestamps (NOW())
	// - Получения значений с DEFAULT constraints
	//
	// NOW() - PostgreSQL функция возвращающая текущий timestamp
	// Почему NOW() а не time.Now() в Go?
	// - Одно время на сервере БД (важно для distributed systems)
	// - Не зависит от часового пояса клиента
	// - Гарантирует консистентность если несколько INSERT в транзакции
	query := `
        INSERT INTO public.users (id, email, google_id, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        RETURNING id, created_at, updated_at
    `

	// QueryRow потому что RETURNING возвращает одну строку
	row := r.pool.QueryRow(ctx, query,
		user.ID,       // $1 - UUID generated in domain.NewUser()
		user.Email,    // $2 - может быть NULL (тип *string)
		user.GoogleID, // $3 - обязательное поле
	)

	// Обновляем user новыми значениями из БД
	// В нашем случае ID не меняется (генерим на клиенте)
	// Но updated_at/created_at берём из БД для консистентности
	err := row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		// Возможные ошибки:
		// - pgx.ErrNoRows - не должно произойти с RETURNING
		// - constraint violation (UNIQUE, FOREIGN KEY)
		// - connection error
		// - serialization error (если используем transactions)
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

// Update - обновляет существующего пользователя
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	// Валидация перед сохранением
	if err := user.Validate(); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	// UPDATE обновляет только указанные поля
	// updated_at автоматически обновляется через NOW()
	//
	// Почему не UPDATE всех полей?
	// - Меньше нагрузка на БД (только изменённые поля)
	// - Не перезаписываем created_at по ошибке
	// - Явно показываем что именно меняется
	query := `
        UPDATE public.users
        SET email = $2, google_id = $3, updated_at = NOW()
        WHERE id = $1
    `

	// Exec - выполняет команду БЕЗ возврата строк
	// В отличие от Query/QueryRow не создаёт result set
	// Используй для: INSERT, UPDATE, DELETE, CREATE TABLE, etc.
	//
	// Возвращает CommandTag - метаданные о выполнении:
	// - RowsAffected() - количество затронутых строк
	// - Insert() - был ли это INSERT
	// - Update() - был ли это UPDATE
	// - Delete() - был ли это DELETE
	commandTag, err := r.pool.Exec(ctx, query,
		user.ID,       // $1 WHERE id = ?
		user.Email,    // $2 SET email = ?
		user.GoogleID, // $3 SET google_id = ?
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	// RowsAffected - количество изменённых строк
	// Если 0 - значит WHERE clause не нашёл записей
	// В нашем случае это означает что пользователь не существует
	if commandTag.RowsAffected() == 0 {
		return out_ports.ErrNotFound
	}

	return nil
}
