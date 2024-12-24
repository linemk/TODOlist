package handlers

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"os"
	"time"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// функция создания токена при правильном вводе пароля
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	// создаем структуру payload в токене
	var payload struct {
		Password string `json:"password"`
	}

	// декодим входящий запрос
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"Некорректный запрос"}`, http.StatusBadRequest)
		return
	}

	// достаем пароль из env файла
	envPassword := os.Getenv("TODO_PASSWORD")
	// проверка подлинности пароля
	if envPassword == "" || payload.Password != envPassword {
		http.Error(w, `{"error":"Неверный пароль"}`, http.StatusUnauthorized)
		return
	}
	// создаем токен через SH256 + добавляем payload
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"password":   envPassword,
		"expiration": time.Now().Add(time.Hour * 8).Unix(), // время истечения
	})

	// шифруем через secret фразу и подписываем
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, `{"error":"Ошибка генерации токена"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// отправляем во фронт
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token": signedToken,
	})
}

// функция - прослойка
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// если нет пароля в env, этот пункт пропускается
		envPassword := os.Getenv("TODO_PASSWORD")
		if envPassword == "" {
			next.ServeHTTP(w, r) // Если пароль не задан, пропускаем
			return
		}

		// достаем из кук токен
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, `{"error":"Необходима аутентификация"}`, http.StatusUnauthorized)
			return
		}

		// парсим
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		// проверяем валидацию
		if err != nil || !token.Valid {
			http.Error(w, `{"error":"Токен недействителен"}`, http.StatusUnauthorized)
			return
		}
		// проверяем payload
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["password"] != envPassword {
			http.Redirect(w, r, "/login.html", http.StatusFound)
			return
		}
		// кидаем на обработчик
		next.ServeHTTP(w, r)
	})
}
