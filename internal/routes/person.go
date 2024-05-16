package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BukhryakovVladimir/learn_live/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"regexp"
	"time"
)

const UniqueViolationErr = pq.ErrorCode("23505")

func SignupPerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid HTTP method. Only POST is allowed.", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie(jwtName)

	if err != nil {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}

	token, err := jwtCheck(cookie)

	if err != nil {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}

	claims := token.Claims.(*jwt.RegisteredClaims)

	userExists, err := checkUserExists(claims.Issuer)
	if err != nil {
		http.Error(w, "Error while checking user authorization", http.StatusInternalServerError)
		return
	}

	if !userExists {
		log.Println("User with id ", claims.Issuer, "does not exist: ", err)
		http.Error(w, "You are not logged in", http.StatusUnauthorized)
		return
	}

	isAdmin, err := isAdmin(claims.Issuer)
	if err != nil {
		http.Error(w, "Error while checking administrator privileges", http.StatusInternalServerError)
		return
	}

	if !isAdmin {
		http.Error(w, "You do not have administrator privileges to add movies", http.StatusUnauthorized)
		return
	}

	var person model.Person

	err = json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if !isValidUsername(person.Username) {
		http.Error(w,
			"Username should have at least 3 characters and consist only of English letters and digits.",
			http.StatusBadRequest)
		return
	}

	if !isValidPassword(person.Password) {
		http.Error(w,
			"Password should have at least 8 characters and include both English letters and digits. Special characters optionally.",
			http.StatusBadRequest)
		return
	}

	if person.BirthDate.After(time.Now()) {
		http.Error(w, "Birth date cannot be in the future", http.StatusBadRequest)
		return
	}

	insertPersonQuery := `INSERT INTO person (username, password, firstname, lastname, email, phone_number, group_id,
                    is_professor, is_admin, sex, birthDate) 
					VALUES ($1::text, $2::text, $3::text, $4::text, $5::text, $6::text, $7::numeric, 
					        $8::bool, $9::bool, $10::text, $11::date);`

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(person.Password), 14)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, insertPersonQuery,
		person.Username,
		passwordHash,
		person.FirstName,
		person.LastName,
		person.Email,
		person.PhoneNumber,
		person.GroupID,
		person.IsProfessor,
		person.IsAdmin,
		person.Sex,
		person.BirthDate,
	)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("SignupPerson QueryRowContext deadline exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
			return
		}

		var pgErr *pq.Error
		if ok := errors.As(err, &pgErr); !ok {
			log.Println("Internal server error: ", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if pgErr.Code == "23505" {
			log.Println("Unique key violation, username already exists: ", err)
			http.Error(w, "Username already exists", http.StatusGatewayTimeout)
			return
		}

		log.Println("Database error: ", pgErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal("Signup successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Write failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CheckIsAdminOrProfessor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie(jwtName)

	if err != nil {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}

	token, err := jwtCheck(cookie)

	if err != nil {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}

	claims := token.Claims.(*jwt.RegisteredClaims)

	userExists, err := checkUserExists(claims.Issuer)
	if err != nil {
		http.Error(w, "Error while checking user authorization", http.StatusInternalServerError)
		return
	}

	if !userExists {
		log.Println("User with id ", claims.Issuer, "does not exist: ", err)
		http.Error(w, "You are not logged in", http.StatusUnauthorized)
		return
	}

	isProfessor, err := isProfessor(claims.Issuer)
	if err != nil {
		http.Error(w, "Error while checking if user is a professor", http.StatusInternalServerError)
		return
	}

	isAdmin, err := isAdmin(claims.Issuer)
	if err != nil {
		http.Error(w, "Error while checking if user is an admin", http.StatusInternalServerError)
		return
	}

	adminOrProfessor := model.AdminOrProfessor{
		UserID:      claims.Issuer,
		IsProfessor: isProfessor,
		IsAdmin:     isAdmin}

	resp, err := json.Marshal(adminOrProfessor)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Check Is Admin Or Professor failed: %v\n", err)
	}
}

func LoginPerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid HTTP method. Only POST is allowed.", http.StatusMethodNotAllowed)
		return
	}

	var person model.Person

	err := json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	getUserDataQuery := `SELECT id, firstname, lastname, password FROM person WHERE username = $1::text`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	row := db.QueryRowContext(ctx, getUserDataQuery, person.Username)
	if err = row.Err(); err != nil {
		if errors.Is(row.Err(), context.DeadlineExceeded) {
			log.Println("LoginPerson QueryRowContext deadline exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
			return
		}

		log.Println("Database error: ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	var userID, firstname, lastname, passwordHash string
	if err := row.Scan(&userID, &firstname, &lastname, &passwordHash); err != nil {
		http.Error(w, "Username not found", http.StatusNotFound)
		return
	}

	if err := row.Err(); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(person.Password)); err != nil {
		http.Error(w, "Incorrect password", http.StatusUnauthorized)
		return
	}

	subject := fmt.Sprintf("%v %v (%v ID: %v)",
		lastname,
		firstname,
		person.Username,
		userID)

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    userID,
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
	})

	token, err := claims.SignedString([]byte(secretKey))

	if err != nil {
		http.Error(w, "Could not login", http.StatusUnauthorized)
		return
	}

	tokenCookie := http.Cookie{
		Name:     jwtName,
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HttpOnly: false,
	}

	http.SetCookie(w, &tokenCookie)
	resp, err := json.Marshal("Successfully logged in")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Write failed: %v\n", err)
	}
}

// RegEx. Обязательно латинские буквы, цифры и длина >= 3.
func isValidUsername(username string) bool {
	pattern := "^[a-zA-Z0-9]{3,}$"

	regexpPattern := regexp.MustCompile(pattern)

	return regexpPattern.MatchString(username)
}

// RegEx. Обязательно латинские буквы, цифры и длина >= 8. Опционально специальные символы.
func isValidPassword(password string) bool {
	pattern := `^[a-zA-Z0-9!@#$%^&*()-_=+,.?;:{}|<>]*[a-zA-Z]+[0-9!@#$%^&*()-_=+,.?;:{}|<>]*[0-9]+[a-zA-Z0-9!@#$%^&*()-_=+,.?;:{}|<>]*$`

	regexpPattern := regexp.MustCompile(pattern)

	return regexpPattern.MatchString(password) && len(password) >= 8
}
