package routes

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/BukhryakovVladimir/learn_live/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"time"
)

func ListGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	listGroupsQuery := `SELECT id, group_name FROM group_uni WHERE id <> 1 AND id <> 2;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listGroupsQuery)
	defer rows.Close()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListGroups QueryRowContext deadline exceeded: ", err)
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
			log.Println("Unique key violation, group already exists: ", err)
			http.Error(w, "Group already exists", http.StatusGatewayTimeout)
			return
		}

		log.Println("Database error: ", pgErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var group model.Group
	var groups []model.Group

	for rows.Next() {
		if err := rows.Scan(&group.ID, &group.GroupName); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(groups)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List groups failed: %v\n", err)
	}
}

func ListStudentsOfAGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	paramGroupID := r.URL.Query().Get("group_id")

	groupID, err := strconv.Atoi(paramGroupID)
	if err != nil {
		http.Error(w, "group_id must be an integer", http.StatusBadRequest)
	}

	listGroupsQuery := `
		SELECT id, username, firstname, lastname, group_id, sex, birthdate
		FROM person
		WHERE group_id = $1 AND is_professor = false AND is_admin = false;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listGroupsQuery, groupID)
	defer rows.Close()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListStudentsOfAGroup QueryRowContext deadline exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
			return
		}

		var pgErr *pq.Error
		if ok := errors.As(err, &pgErr); !ok {
			log.Println("Internal server error: ", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Println("Database error: ", pgErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var student model.Student
	var students []model.Student

	for rows.Next() {
		if err := rows.Scan(
			&student.ID,
			&student.Username,
			&student.FirstName,
			&student.LastName,
			&student.GroupID,
			&student.Sex,
			&student.BirthDate); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(students)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List Students Of A Group failed: %v\n", err)
	}

}

func AddGroup(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "You do not have administrator privileges to add groups", http.StatusUnauthorized)
		return
	}

	var group model.Group

	err = json.NewDecoder(r.Body).Decode(&group)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if len([]rune(group.GroupName)) > 255 {
		http.Error(w, "Maximum group name length is 255 characters", http.StatusBadRequest)
		return
	}

	insertGroupQuery := `INSERT INTO group_uni (group_name) VALUES ($1::text);`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, insertGroupQuery, group.GroupName)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("AddGroup QueryRowContext deadline exceeded: ", err)
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
			log.Println("Unique key violation, group already exists: ", err)
			http.Error(w, "Group already exists", http.StatusGatewayTimeout)
			return
		}

		log.Println("Database error: ", pgErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal("Inserting group successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("add-group failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid HTTP method. Only PUT is allowed.", http.StatusMethodNotAllowed)
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
		http.Error(w, "You do not have administrator privileges to update groups", http.StatusUnauthorized)
		return
	}

	var group model.Group

	err = json.NewDecoder(r.Body).Decode(&group)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if len([]rune(group.GroupName)) > 255 {
		http.Error(w, "Maximum group name length is 255 characters", http.StatusBadRequest)
		return
	}

	updateGroupQuery := `UPDATE group_uni SET group_name = $1::text WHERE id = $2;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, updateGroupQuery, group.GroupName, group.ID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("UpdateGroup QueryRowContext deadline exceeded: ", err)
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
			log.Println("Unique key violation, group already exists: ", err)
			http.Error(w, "Group already exists", http.StatusGatewayTimeout)
			return
		}

		log.Println("Database error: ", pgErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal("Update group successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Update group failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid HTTP method. Only DELETE is allowed.", http.StatusMethodNotAllowed)
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
		http.Error(w, "You do not have administrator privileges to delete groups", http.StatusUnauthorized)
		return
	}

	idParam := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusBadRequest)
		return
	}

	deleteGroupQuery := `DELETE FROM group_uni WHERE id = $1;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, deleteGroupQuery, id)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("DeleteGroup QueryRowContext deadline exceeded: ", err)
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
			log.Println("Unique key violation, group already exists: ", err)
			http.Error(w, "Group already exists", http.StatusGatewayTimeout)
			return
		}

		log.Println("Database error: ", pgErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal("Delete group successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Delete group failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
