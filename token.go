package main

import (
	"fmt"
	"net/http"
	"time"

	"firebase.google.com/go/auth"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func verifyToken(c echo.Context) (*jwt.Token, error) {
	// Pobierz token z ciasteczka
	cookie, err := c.Cookie("token")
	if err != nil {
		return nil, err
	}

	// Weryfikacja tokena
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		// Sprawdź metodę podpisu
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Nieprawidłowa metoda podpisu")
		}

		// Zwróć klucz do weryfikacji
		return []byte("sekret"), nil
	})
	if err != nil {
		return nil, err
	}

	// Sprawdź, czy token jest ważny
	if !token.Valid {
		return nil, fmt.Errorf("Token nieprawidłowy")
	}

	return token, nil
}

func CreateTokenWithCookie(c echo.Context, user *auth.UserRecord) error {
	// Tworzenie tokena
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.UID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() // Token wygasa po 72 godzinach

	// Podpisanie tokena i dodanie do ciasteczka
	t, err := token.SignedString([]byte("sekret"))
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    t,
		Expires:  time.Now().Add(72 * time.Hour),
		HttpOnly: true,
	}
	c.SetCookie(cookie)

	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
