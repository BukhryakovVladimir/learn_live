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

func ListGroupsSubjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	listGroupsSubjectsQuery := `SELECT group_id, group_name, subject_id, subject_name 
	FROM group_subject
	LEFT JOIN group_uni gu on group_id = gu.id 
    LEFT JOIN subject s on subject_id = s.id;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listGroupsSubjectsQuery)
	defer rows.Close()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListGroupsSubjects QueryRowContext deadline exceeded: ", err)
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

	var groupSubject model.GroupSubject
	var groupsSubjects []model.GroupSubject

	for rows.Next() {
		if err := rows.Scan(&groupSubject.GroupID, &groupSubject.GroupName,
			&groupSubject.SubjectID, &groupSubject.SubjectName); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		groupsSubjects = append(groupsSubjects, groupSubject)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(groupsSubjects)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List Groups and Subjects failed: %v\n", err)
	}
}

func ListSubjectsOfAGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	paramGroupID := r.URL.Query().Get("group_id")

	GroupID, err := strconv.Atoi(paramGroupID)
	if err != nil {
		http.Error(w, "group_id must be an integer", http.StatusBadRequest)
		return
	}

	listSubjectsOfAGroupQuery := `SELECT group_id, group_name, subject_id, subject_name from group_subject
    LEFT JOIN group_uni gu ON group_id = gu.id
    LEFT JOIN subject s ON subject_id = s.id
    WHERE group_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listSubjectsOfAGroupQuery, GroupID)
	defer rows.Close()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListSubjectsOfAGroup QueryRowContext deadline exceeded: ", err)
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

	var groupSubject model.GroupSubject
	var groupsSubjects []model.GroupSubject

	for rows.Next() {
		if err := rows.Scan(&groupSubject.GroupID, &groupSubject.GroupName,
			&groupSubject.SubjectID, &groupSubject.SubjectName); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		groupsSubjects = append(groupsSubjects, groupSubject)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(groupsSubjects)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List Subjects Of A Group failed: %v\n", err)
	}
}

func ListGroupsThatHaveASubject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	paramSubjectID := r.URL.Query().Get("subject_id")

	SubjectID, err := strconv.Atoi(paramSubjectID)
	if err != nil {
		http.Error(w, "group_id must be an integer", http.StatusBadRequest)
		return
	}

	listGroupsThatHaveASubjectQuery := `SELECT group_id, group_name, subject_id, subject_name from group_subject
    LEFT JOIN subject s ON subject_id = s.id
    LEFT JOIN group_uni gu ON group_id = gu.id
    WHERE subject_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listGroupsThatHaveASubjectQuery, SubjectID)
	defer rows.Close()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListGroupsThatHaveASubject QueryRowContext deadline exceeded: ", err)
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

	var groupSubject model.GroupSubject
	var groupsSubjects []model.GroupSubject

	for rows.Next() {
		if err := rows.Scan(&groupSubject.GroupID, &groupSubject.GroupName,
			&groupSubject.SubjectID, &groupSubject.SubjectName); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		groupsSubjects = append(groupsSubjects, groupSubject)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(groupsSubjects)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List Groups That Have A Subject failed: %v\n", err)
	}
}

func AddGroupSubject(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "You do not have administrator privileges to add group_subject relation", http.StatusUnauthorized)
		return
	}

	var groupSubject model.GroupSubject
	err = json.NewDecoder(r.Body).Decode(&groupSubject)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	insertGroupSubjectQuery := `INSERT INTO group_subject (group_id, subject_id) VALUES ($1, $2);`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, insertGroupSubjectQuery, groupSubject.GroupID, groupSubject.SubjectID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("AddGroupSubject QueryRowContext deadline exceeded: ", err)
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
			log.Println("Unique key violation, group_subject already exists: ", err)
			http.Error(w, "Group Subject relation already exists", http.StatusGatewayTimeout)
			return
		}

		log.Println("Database error: ", pgErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal("Inserting group subject relation successful")
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

func UpdateGroupSubject(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "You do not have administrator privileges to update group_subject relation", http.StatusUnauthorized)
		return
	}

	var updateGroupSubject model.UpdateGroupSubject
	err = json.NewDecoder(r.Body).Decode(&updateGroupSubject)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	insertGroupSubjectQuery := `UPDATE group_subject SET group_id = $1, subject_id = $2
	WHERE group_id = $3 AND subject_id = $4;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, insertGroupSubjectQuery, updateGroupSubject.NewGroupID,
		updateGroupSubject.NewSubjectID, updateGroupSubject.OldGroupID, updateGroupSubject.OldSubjectID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("UpdateGroupSubject QueryRowContext deadline exceeded: ", err)
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
			log.Println("Unique key violation, group_subject already exists: ", err)
			http.Error(w, "Group Subject relation already exists", http.StatusGatewayTimeout)
			return
		}

		log.Println("Database error: ", pgErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal("Updating group subject relation successful")
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

func DeleteGroupSubject(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "You do not have administrator privileges to delete group_subject relation", http.StatusUnauthorized)
		return
	}

	ParamGroupID := r.URL.Query().Get("group_id")
	ParamSubjectID := r.URL.Query().Get("subject_id")

	groupID, err := strconv.Atoi(ParamGroupID)
	if err != nil {
		http.Error(w, "group_id must be an integer number", http.StatusBadRequest)
		return
	}

	subjectID, err := strconv.Atoi(ParamSubjectID)
	if err != nil {
		http.Error(w, "subject_id must be an integer number", http.StatusBadRequest)
		return
	}

	insertGroupSubjectQuery := `DELETE FROM group_subject WHERE group_id = $1 and subject_id = $2;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, insertGroupSubjectQuery, groupID, subjectID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("DeleteGroupSubject QueryRowContext deadline exceeded: ", err)
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

	resp, err := json.Marshal("Deleting group subject relation successful")
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
