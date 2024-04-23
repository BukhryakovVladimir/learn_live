package routes

import (
	"context"
	"database/sql"
	"errors"
	"github.com/BukhryakovVladimir/learn_live/internal/postgres"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB // Пул соединений с БД

var (
	queryTimeLimit int
	secretKey      string
	jwtName        string
)

// jwtCheck парсит JWT токен из переданного HTTP cookie используя секретный ключ secretKey
func jwtCheck(cookie *http.Cookie) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(cookie.Value, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	return token, err
}

func isAdmin(issuer string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	var isAdmin bool

	isAdminQuery := `SELECT is_admin FROM person WHERE id = $1`

	err := db.QueryRowContext(ctx, isAdminQuery, issuer).Scan(&isAdmin)

	if err != nil {
		return false, err
	}

	return isAdmin, nil
}

func checkUserExists(issuer string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	userExistsQuery := `SELECT username FROM person WHERE id = $1`

	var username string
	err := db.QueryRowContext(ctx, userExistsQuery, issuer).Scan(&username)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("user not found. Unauthorized access not allowed")
		}
		return false, err
	}

	if username != "" {
		return true, nil
	}

	return false, errors.New("user not found. Unauthorized access not allowed")
}

// InitConnPool создаёт пул соединений с БД
func InitConnPool() error {
	var err error
	strQueryTimeLimit := os.Getenv("QUERY_TIME_LIMIT")
	if strQueryTimeLimit == "" {
		return errors.New("environment variable QUERY_TIME_LIMIT is empty")
	}
	queryTimeLimit, err = strconv.Atoi(strQueryTimeLimit)
	if err != nil {
		return err
	}
	secretKey = os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return errors.New("environment variable SECRET_KEY is empty")
	}
	jwtName = os.Getenv("JWT_NAME")
	if jwtName == "" {
		return errors.New("environment variable JWT_NAME is empty")
	}

	db, err = postgres.Dial()
	if err != nil {
		return err
	}
	return nil
}
