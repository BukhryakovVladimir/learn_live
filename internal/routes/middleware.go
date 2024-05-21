package routes

import (
	"context"
	"database/sql"
	"errors"
	"github.com/BukhryakovVladimir/learn_live/internal/postgres"
	"github.com/golang-jwt/jwt/v5"
	"log"
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

func isProfessor(issuer string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	var isProfessor bool

	isAdminQuery := `SELECT is_professor FROM person WHERE id = $1`

	err := db.QueryRowContext(ctx, isAdminQuery, issuer).Scan(&isProfessor)

	if err != nil {
		return false, err
	}

	return isProfessor, nil
}

func professorHasSubject(issuer string, subjectID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	var hasSubject bool

	isAdminQuery := `SELECT true FROM professor_subject WHERE professor_id = $1 AND subject_id = $2`

	err := db.QueryRowContext(ctx, isAdminQuery, issuer, subjectID).Scan(&hasSubject)

	if err != nil {
		return false, err
	}

	return hasSubject, nil
}

func professorHasGroup(issuer string, studentID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	var hasGroup bool
	log.Println("professorHasGroup: issuer = ", issuer, " studentID = ", studentID)
	isAdminQuery := `
		SELECT true 
		FROM professor_group pg 
		LEFT JOIN person p ON pg.group_id = p.group_id
		WHERE professor_id = $1 AND p.id = $2`

	err := db.QueryRowContext(ctx, isAdminQuery, issuer, studentID).Scan(&hasGroup)

	if err != nil {
		return false, err
	}

	return hasGroup, nil
}

func studentHasSubject(studentID, subjectID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	var hasSubject bool

	isAdminQuery := `
		SELECT true 
		FROM person p
		JOIN group_subject gs ON p.group_id = gs.group_id
		WHERE p.id = $1 AND gs.subject_id = $2`

	err := db.QueryRowContext(ctx, isAdminQuery, studentID, subjectID).Scan(&hasSubject)

	if err != nil {
		return false, err
	}

	return hasSubject, nil
}

func isStudent(issuer string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	var isAdmin bool
	var isProfessor bool

	isAdminQuery := `SELECT is_admin, is_professor FROM person WHERE id = $1`

	err := db.QueryRowContext(ctx, isAdminQuery, issuer).Scan(&isAdmin, &isProfessor)

	if err != nil {
		return false, err
	}

	return !isAdmin && !isProfessor, nil
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
