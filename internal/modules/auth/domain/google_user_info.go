// internal/modules/auth/domain/google_user_info.go
package domain

// GoogleUserInfo - данные пользователя полученные от Google OAuth2 API
// Структура соответствует ответу от https://www.googleapis.com/oauth2/v2/userinfo
//
// Документация Google:
// https://developers.google.com/identity/protocols/oauth2/openid-connect#obtainuserinfo
type GoogleUserInfo struct {
	// Sub - Subject (уникальный идентификатор пользователя в Google)
	// Это НЕ email! Это постоянный ID который не меняется
	// Формат: строка из цифр, например: "108123456789012345678"
	Sub string `json:"sub"`

	// Email - адрес электронной почты
	// Может быть пустым если:
	//   1. Пользователь не дал разрешение на scope "email"
	//   2. Email скрыт в настройках конфиденциальности Google
	Email string `json:"email"`

	// EmailVerified - подтверждён ли email в Google
	// true = Google подтвердил что пользователь владеет этим email
	// false = email не подтверждён или отсутствует
	EmailVerified bool `json:"email_verified"`

	// Name - полное имя пользователя
	// Пример: "John Doe"
	Name string `json:"name"`

	// GivenName - имя (first name)
	// Пример: "John"
	GivenName string `json:"given_name"`

	// FamilyName - фамилия (last name)
	// Пример: "Doe"
	FamilyName string `json:"family_name"`

	// Picture - URL аватарки пользователя
	// Пример: "https://lh3.googleusercontent.com/a/..."
	// Размер по умолчанию: 96x96 пикселей
	// Можно изменить размер добавив ?sz=200 в конец URL
	Picture string `json:"picture"`

	// Locale - предпочитаемый язык пользователя
	// Формат: ISO 639-1 код
	// Примеры: "en" (English), "ru" (Russian), "uk" (Ukrainian)
	Locale string `json:"locale"`
}

// ToUser - конвертирует данные Google в доменную сущность User
// Это Anti-Corruption Layer паттерн: защищает domain от изменений внешнего API
//
// Логика конвертации:
//  1. Sub (Google ID) → обязательное поле
//  2. Email → опциональное (может быть пустым)
//  3. Остальные поля (Name, Picture) игнорируются в текущей версии
//     (можно добавить позже если понадобится хранить профиль)
//
// Возвращает новый экземпляр User через фабричный метод NewUser()
func (g *GoogleUserInfo) ToUser() *User {
	// Обрабатываем опциональный email
	// Если email пустой (пользователь не дал разрешение), сохраняем nil
	var email *string
	if g.Email != "" {
		email = &g.Email
	}

	// NewUser() создаст User с:
	//   - новым UUID в поле ID
	//   - g.Sub в поле GoogleID
	//   - email (или nil) в поле Email
	//   - текущим временем в CreatedAt/UpdatedAt
	return NewUser(g.Sub, email)
}

// ShouldStoreEmail - проверяет стоит ли сохранять email в БД
// Сохраняем только если:
//  1. Email не пустой
//  2. Email подтверждён Google
//
// Пример использования:
//
//	if userInfo.ShouldStoreEmail() {
//	    user.Email = &userInfo.Email
//	}
func (g *GoogleUserInfo) ShouldStoreEmail() bool {
	return g.Email != "" && g.EmailVerified
}
