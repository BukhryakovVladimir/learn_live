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
	"unicode/utf8"
)

func ListCurrentUserTotalGrades(w http.ResponseWriter, r *http.Request) {
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

	isStudent, err := isStudent(claims.Issuer)
	if err != nil {
		http.Error(w, "Error while checking is user is a student", http.StatusInternalServerError)
		return
	}

	if !isStudent {
		http.Error(w, "Only students have total grades", http.StatusUnauthorized)
		return
	}

	listCurrentUserTotalGradesQuery := `
	SELECT DISTINCT stg.subject_id, s.subject_name, stg.grade
	FROM student_total_grades stg
	JOIN subject s ON stg.subject_id = s.id  
	WHERE student_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listCurrentUserTotalGradesQuery, claims.Issuer)
	defer rows.Close()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListCurrentUserTotalGrades QueryRowContext deadline exceeded: ", err)
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

	var studentTotalGrade model.StudentTotalGrade
	var studentTotalGrades []model.StudentTotalGrade

	for rows.Next() {
		if err := rows.Scan(
			&studentTotalGrade.SubjectID,
			&studentTotalGrade.SubjectName,
			&studentTotalGrade.Grade); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		studentTotalGrades = append(studentTotalGrades, studentTotalGrade)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(studentTotalGrades)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List Current User Total Grades failed: %v\n", err)
	}
}

func ListTotalGradesOfAStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	paramStudentID := r.URL.Query().Get("student_id")

	studentID, err := strconv.Atoi(paramStudentID)
	if err != nil {
		http.Error(w, "student_id must be an integer", http.StatusBadRequest)
	}

	listTotalGradesOfAStudentQuery := `
	SELECT DISTINCT stg.student_id, p.firstname, p.lastname, p.group_id, 
	       g.group_name, stg.subject_id, s.subject_name, stg.grade
	FROM student_total_grades stg
	JOIN subject s ON stg.subject_id = s.id
	JOIN person p ON stg.student_id = p.id
	JOIN group_subject gs ON p.group_id = gs.group_id
	JOIN group_uni g ON gs.group_id = g.id
	WHERE student_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listTotalGradesOfAStudentQuery, studentID)
	defer rows.Close()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListTotalGradesOfAStudent QueryRowContext deadline exceeded: ", err)
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

	var studentTotalGrade model.StudentTotalGrade
	var studentTotalGrades []model.StudentTotalGrade

	for rows.Next() {
		if err := rows.Scan(
			&studentTotalGrade.StudentID,
			&studentTotalGrade.StudentFirstname,
			&studentTotalGrade.StudentLastname,
			&studentTotalGrade.StudentGroupID,
			&studentTotalGrade.StudentGroupName,
			&studentTotalGrade.SubjectID,
			&studentTotalGrade.SubjectName,
			&studentTotalGrade.Grade); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		studentTotalGrades = append(studentTotalGrades, studentTotalGrade)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(studentTotalGrades)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List Total Grades Of A Student failed: %v\n", err)
	}
}

func ListTotalGradesOfAGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	paramGroupID := r.URL.Query().Get("group_id")

	groupID, err := strconv.Atoi(paramGroupID)
	if err != nil {
		http.Error(w, "group_id must be an integer", http.StatusBadRequest)
	}

	listTotalGradesOfAStudentQuery := `
	SELECT DISTINCT stg.student_id, p.firstname, p.lastname, p.group_id, 
	       g.group_name, stg.subject_id, s.subject_name, stg.grade
	FROM student_total_grades stg
	JOIN subject s ON stg.subject_id = s.id
	JOIN person p ON stg.student_id = p.id
	JOIN group_subject gs ON p.group_id = gs.group_id
	JOIN group_uni g ON gs.group_id = g.id
	WHERE p.group_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listTotalGradesOfAStudentQuery, groupID)
	defer rows.Close()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListTotalGradesOfAGroup QueryRowContext deadline exceeded: ", err)
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

	var studentTotalGrade model.StudentTotalGrade
	var studentTotalGrades []model.StudentTotalGrade

	for rows.Next() {
		if err := rows.Scan(
			&studentTotalGrade.StudentID,
			&studentTotalGrade.StudentFirstname,
			&studentTotalGrade.StudentLastname,
			&studentTotalGrade.StudentGroupID,
			&studentTotalGrade.StudentGroupName,
			&studentTotalGrade.SubjectID,
			&studentTotalGrade.SubjectName,
			&studentTotalGrade.Grade); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		studentTotalGrades = append(studentTotalGrades, studentTotalGrade)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(studentTotalGrades)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List Total Grades Of A Group failed: %v\n", err)
	}
}

func InsertTotalGradeOfAStudent(w http.ResponseWriter, r *http.Request) {
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

	isProfessor, err := isProfessor(claims.Issuer)
	if err != nil {
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !isProfessor {
		http.Error(w, "You do not have professor privileges", http.StatusUnauthorized)
		return
	}

	var studentTotalGrade model.StudentTotalGrade
	err = json.NewDecoder(r.Body).Decode(&studentTotalGrade)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	hasSubject, err := professorHasSubject(claims.Issuer, studentTotalGrade.SubjectID)
	if err != nil {
		log.Println("professorHasSubject error: ", err)
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !hasSubject {
		http.Error(w, "You can only set total grades for subjects that you teach", http.StatusUnauthorized)
		return
	}

	hasGroup, err := professorHasGroup(claims.Issuer, studentTotalGrade.StudentID)
	if err != nil {
		log.Println("professorHasGroup error: ", err)
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !hasGroup {
		http.Error(w, "You can only set total grades for students from groups that you teach", http.StatusUnauthorized)
		return
	}

	studentHasSubject, err := studentHasSubject(studentTotalGrade.StudentID, studentTotalGrade.SubjectID)
	if err != nil {
		http.Error(w, "Error while checking if student has this subject", http.StatusInternalServerError)
		return
	}

	if !studentHasSubject {
		http.Error(w, "Student doesn't have this subject in his program", http.StatusUnauthorized)
		return
	}

	if utf8.RuneCountInString(studentTotalGrade.Grade) > 50 {
		http.Error(w, "Grade length cannot be bigger than 50 characters", http.StatusBadRequest)
		return
	}

	insertTotalGradeOfAStudentQuery := `
		INSERT INTO student_total_grades 
		    (student_id, subject_id, grade) 
		VALUES ($1, $2, $3);`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		insertTotalGradeOfAStudentQuery,
		studentTotalGrade.StudentID,
		studentTotalGrade.SubjectID,
		studentTotalGrade.Grade)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("InsertTotalGradeOfAStudent QueryRowContext deadline exceeded: ", err)
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
			log.Println("Unique key violation, student already has a grade for this subject: ", err)
			http.Error(w, "Student already has a grade for this subject", http.StatusGatewayTimeout)
			return
		}

		log.Println("Database error: ", pgErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal("Inserting Total Grade Of A Student Successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("InsertTotalGradeOfAStudent failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func UpdateTotalGradeOfAStudent(w http.ResponseWriter, r *http.Request) {
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

	isProfessor, err := isProfessor(claims.Issuer)
	if err != nil {
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !isProfessor {
		http.Error(w, "You do not have professor privileges", http.StatusUnauthorized)
		return
	}

	var studentTotalGrade model.StudentTotalGrade
	err = json.NewDecoder(r.Body).Decode(&studentTotalGrade)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	hasSubject, err := professorHasSubject(claims.Issuer, studentTotalGrade.SubjectID)
	if err != nil {
		log.Println("professorHasSubject error: ", err)
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !hasSubject {
		http.Error(w, "You can only update total grades for subjects that you teach", http.StatusUnauthorized)
		return
	}

	hasGroup, err := professorHasGroup(claims.Issuer, studentTotalGrade.StudentID)
	if err != nil {
		log.Println("professorHasGroup error: ", err)
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !hasGroup {
		http.Error(w, "You can only update total grades for students from groups that you teach", http.StatusUnauthorized)
		return
	}

	studentHasSubject, err := studentHasSubject(studentTotalGrade.StudentID, studentTotalGrade.SubjectID)
	if err != nil {
		http.Error(w, "Error while checking if student has this subject", http.StatusInternalServerError)
		return
	}

	if !studentHasSubject {
		http.Error(w, "Student doesn't have this subject in his program", http.StatusUnauthorized)
		return
	}

	if utf8.RuneCountInString(studentTotalGrade.Grade) > 50 {
		http.Error(w, "Grade length cannot be bigger than 50 characters", http.StatusBadRequest)
		return
	}

	updateTotalGradeOfAStudentQuery := `
		UPDATE student_total_grades 
		SET student_id = $1, subject_id = $2, grade = $3 
		WHERE id = $4;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		updateTotalGradeOfAStudentQuery,
		studentTotalGrade.StudentID,
		studentTotalGrade.SubjectID,
		studentTotalGrade.Grade,
		studentTotalGrade.ID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("UpdateTotalGradeOfAStudent QueryRowContext deadline exceeded: ", err)
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
			log.Println("Unique key violation, student already has a grade for this subject: ", err)
			http.Error(w, "Student already has a grade for this subject", http.StatusGatewayTimeout)
			return
		}

		log.Println("Database error: ", pgErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal("Updating Total Grade Of A Student Successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("UpdateTotalGradeOfAStudent failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func DeleteTotalGradeOfAStudent(w http.ResponseWriter, r *http.Request) {
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

	isProfessor, err := isProfessor(claims.Issuer)
	if err != nil {
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !isProfessor {
		http.Error(w, "You do not have professor privileges", http.StatusUnauthorized)
		return
	}

	paramID := r.URL.Query().Get("id")
	id, err := strconv.Atoi(paramID)
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusBadRequest)
	}

	paramStudentID := r.URL.Query().Get("student_id")
	studentID, err := strconv.Atoi(paramStudentID)
	if err != nil {
		http.Error(w, "student_id must be an integer", http.StatusBadRequest)
	}

	paramSubjectID := r.URL.Query().Get("subject_id")
	subjectID, err := strconv.Atoi(paramSubjectID)
	if err != nil {
		http.Error(w, "subject_id must be an integer", http.StatusBadRequest)
	}

	hasSubject, err := professorHasSubject(claims.Issuer, subjectID)
	if err != nil {
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !hasSubject {
		http.Error(w, "You can only delete total grades for subjects that you teach", http.StatusUnauthorized)
		return
	}

	hasGroup, err := professorHasGroup(claims.Issuer, studentID)
	if err != nil {
		log.Println("professorHasGroup error: ", err)
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !hasGroup {
		http.Error(w, "You can only delete total grades for students from groups that you teach", http.StatusUnauthorized)
		return
	}

	studentHasSubject, err := studentHasSubject(studentID, subjectID)
	if err != nil {
		log.Println("studentHasSubject error: ", err)
		http.Error(w, "Error while checking if student has this subject", http.StatusInternalServerError)
		return
	}

	if !studentHasSubject {
		http.Error(w, "Student doesn't have this subject in his program", http.StatusUnauthorized)
		return
	}

	deleteTotalGradeOfAStudentQuery := `DELETE FROM student_total_grades WHERE id = $1 AND student_id = $2 AND subject_id = $3`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, deleteTotalGradeOfAStudentQuery, id, studentID, subjectID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("DeleteTotalGradeOfAStudent QueryRowContext deadline exceeded: ", err)
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

	resp, err := json.Marshal("Deleting Total Grade Of A Student Successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("DeleteTotalGradeOfAStudent failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
