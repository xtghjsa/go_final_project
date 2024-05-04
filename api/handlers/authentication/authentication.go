package authentication

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"main/api/handlers"
	"net/http"
)

// HashPassword - Возвращает JWT-токен
func HashPassword(password string) string {
	secret := []byte(password)
	jwtToken := jwt.New(jwt.SigningMethodHS256)
	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		log.Println("Ошибка при подписании токена", err)
	}
	log.Println(signedToken)
	return signedToken
}

type Password struct {
	SavedPassword string
}

// CheckPasswordHandler - Проверка пароля, аутентификация
func (p *Password) CheckPasswordHandler(w http.ResponseWriter, r *http.Request) {
	type requestedPassword struct {
		Password string `json:"password"`
	}
	var requested requestedPassword

	err := json.NewDecoder(r.Body).Decode(&requested)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка при декодировании запроса", err)
		return
	}
	if requested.Password == "" {
		_, err = w.Write(handlers.ErrorResponse("Введите пароль"))
		if err != nil {
			log.Println("Ошибка возвращения ответа", err)
		}
		return
	}
	if p.SavedPassword == "" {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(handlers.ErrorResponse("Пароль не задан"))
		if err != nil {
			log.Println("Ошибка возвращения ответа", err)
		}
		return
	}
	if p.SavedPassword == requested.Password {
		tokenResponse, err := json.Marshal(map[string]string{"token": HashPassword(requested.Password)})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при записи в json", err)
		}
		_, err = w.Write(tokenResponse)
		if err != nil {
			log.Println("Ошибка возвращения ответа", err)
		}
	} else {
		_, err = w.Write(handlers.ErrorResponse("Неверный пароль"))
		if err != nil {
			log.Println("Ошибка возвращения ответа", err)
		}

	}

}

// Auth - Проверка авторизации для всех основных запросов
func (p *Password) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(p.SavedPassword) > 0 {
			var jwtToken string
			cookie, err := r.Cookie("token")
			if err == nil {
				jwtToken = cookie.Value
			}

			valid := p.ValidateToken(jwtToken)
			if !valid {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}

// ValidateToken - Проверка токена
func (p *Password) ValidateToken(token string) bool {
	secret := []byte(p.SavedPassword)
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		log.Println("Ошибка при декодировании токена", err)
		return false
	}
	if jwtToken.Valid {
		log.Println("Токен валиден, доступ разрешен")
		return true
	} else {
		log.Println("Токен невалиден, доступ запрещен")
		return false
	}
}
