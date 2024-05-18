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

func ListCurrentUserGradesAndAttendanceBySubject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	paramSubjectID := r.URL.Query().Get("subject_id")

	subjectID, err := strconv.Atoi(paramSubjectID)
	if err != nil {
		http.Error(w, "subject_id must be an integer", http.StatusBadRequest)
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
		http.Error(w, "Only students have grades", http.StatusUnauthorized)
		return
	}

	listCurrentUserGradesAndAttendanceQuery := `
	SELECT sg.subject_id, s.subject_name, sg.grade, sg.has_attended
	FROM student_grades sg
	JOIN subject s ON sg.subject_id = s.id  
	WHERE student_id = $1 AND sg.subject_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listCurrentUserGradesAndAttendanceQuery, claims.Issuer, subjectID)
	defer rows.Close()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListCurrentUserGradesAndAttendance QueryRowContext deadline exceeded: ", err)
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

	var studentGrade model.StudentGrade
	var studentGrades []model.StudentGrade

	for rows.Next() {
		if err := rows.Scan(
			&studentGrade.SubjectID,
			&studentGrade.SubjectName,
			&studentGrade.Grade,
			&studentGrade.HasAttended); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		studentGrades = append(studentGrades, studentGrade)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(studentGrades)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List Current User Grades And Attendance failed: %v\n", err)
	}
}

func ListGradesAndAttendanceOfAStudentBySubject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
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

	listGradesAndAttendanceOfAStudentQuery := `
	SELECT sg.student_id, p.firstname, p.lastname, p.group_id, 
	       g.group_name, sg.subject_id, s.subject_name, sg.grade, sg.has_attended
	FROM student_grades sg
	JOIN subject s ON sg.subject_id = s.id
	JOIN person p ON sg.student_id = p.id
	JOIN group_subject gs ON p.group_id = gs.group_id
	JOIN group_uni g ON gs.group_id = g.id
	WHERE student_id = $1 AND sg.subject_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listGradesAndAttendanceOfAStudentQuery, studentID, subjectID)
	defer rows.Close()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListGradesAndAttendanceOfAStudent QueryRowContext deadline exceeded: ", err)
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

	var studentGrade model.StudentGrade
	var studentGrades []model.StudentGrade

	for rows.Next() {
		if err := rows.Scan(
			&studentGrade.StudentID,
			&studentGrade.StudentFirstname,
			&studentGrade.StudentLastname,
			&studentGrade.StudentGroupID,
			&studentGrade.StudentGroupName,
			&studentGrade.SubjectID,
			&studentGrade.SubjectName,
			&studentGrade.Grade,
			&studentGrade.HasAttended); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		studentGrades = append(studentGrades, studentGrade)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(studentGrades)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List Grades And Attendance Of A Student failed: %v\n", err)
	}
}

func ListGradesAndAttendanceOfAGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	paramGroupID := r.URL.Query().Get("group_id")

	groupID, err := strconv.Atoi(paramGroupID)
	if err != nil {
		http.Error(w, "group_id must be an integer", http.StatusBadRequest)
	}

	listGradesAndAttendanceOfAStudentQuery := `
	SELECT sg.student_id, p.firstname, p.lastname, p.group_id, 
	       g.group_name, sg.subject_id, s.subject_name, sg.grade, sg.has_attended
	FROM student_grades sg
	JOIN subject s ON sg.subject_id = s.id
	JOIN person p ON sg.student_id = p.id
	JOIN group_subject gs ON p.group_id = gs.group_id
	JOIN group_uni g ON gs.group_id = g.id
	WHERE p.group_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listGradesAndAttendanceOfAStudentQuery, groupID)
	defer rows.Close()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListGradesAndAttendanceOfAGroup QueryRowContext deadline exceeded: ", err)
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

	var studentGrade model.StudentGrade
	var studentGrades []model.StudentGrade

	for rows.Next() {
		if err := rows.Scan(
			&studentGrade.StudentID,
			&studentGrade.StudentFirstname,
			&studentGrade.StudentLastname,
			&studentGrade.StudentGroupID,
			&studentGrade.StudentGroupName,
			&studentGrade.SubjectID,
			&studentGrade.SubjectName,
			&studentGrade.Grade,
			&studentGrade.HasAttended); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		studentGrades = append(studentGrades, studentGrade)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(studentGrades)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List Grades And Attendance Of A Group failed: %v\n", err)
	}
}

func InsertGradeAndAttendanceOfAStudent(w http.ResponseWriter, r *http.Request) {
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

	var studentGrade model.StudentGrade
	err = json.NewDecoder(r.Body).Decode(&studentGrade)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	hasSubject, err := professorHasSubject(claims.Issuer, studentGrade.SubjectID)
	if err != nil {
		log.Println("professorHasSubject error: ", err)
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !hasSubject {
		http.Error(w, "You can only set grades for subjects that you teach", http.StatusUnauthorized)
		return
	}

	hasGroup, err := professorHasGroup(claims.Issuer, studentGrade.StudentID)
	if err != nil {
		log.Println("professorHasGroup error: ", err)
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !hasGroup {
		http.Error(w, "You can only set grades for students from groups that you teach", http.StatusUnauthorized)
		return
	}

	studentHasSubject, err := studentHasSubject(studentGrade.StudentID, studentGrade.SubjectID)
	if err != nil {
		http.Error(w, "Error while checking if student has this subject", http.StatusInternalServerError)
		return
	}

	if !studentHasSubject {
		http.Error(w, "Student doesn't have this subject in his program", http.StatusUnauthorized)
		return
	}

	if studentGrade.Grade != 0 && studentGrade.HasAttended == false {
		studentGrade.HasAttended = true
	}

	insertGradeAndAttendanceOfAStudentQuery := `
		INSERT INTO student_grades 
		    (student_id, subject_id, grade, has_attended) 
		VALUES ($1, $2, $3, $4);`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		insertGradeAndAttendanceOfAStudentQuery,
		studentGrade.StudentID,
		studentGrade.SubjectID,
		studentGrade.Grade,
		studentGrade.HasAttended)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("InsertGradeAndAttendanceOfAStudent QueryRowContext deadline exceeded: ", err)
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

	resp, err := json.Marshal("Inserting Grade And Attendance Of A Student Successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("InsertGradeAndAttendanceOfAStudent failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func UpdateGradeAndAttendanceOfAStudent(w http.ResponseWriter, r *http.Request) {
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

	var studentGrade model.StudentGrade
	err = json.NewDecoder(r.Body).Decode(&studentGrade)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	hasSubject, err := professorHasSubject(claims.Issuer, studentGrade.SubjectID)
	if err != nil {
		log.Println("professorHasSubject error: ", err)
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !hasSubject {
		http.Error(w, "You can only update grades for subjects that you teach", http.StatusUnauthorized)
		return
	}

	hasGroup, err := professorHasGroup(claims.Issuer, studentGrade.StudentID)
	if err != nil {
		log.Println("professorHasGroup error: ", err)
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !hasGroup {
		http.Error(w, "You can only update grades for students from groups that you teach", http.StatusUnauthorized)
		return
	}

	studentHasSubject, err := studentHasSubject(studentGrade.StudentID, studentGrade.SubjectID)
	if err != nil {
		http.Error(w, "Error while checking if student has this subject", http.StatusInternalServerError)
		return
	}

	if !studentHasSubject {
		http.Error(w, "Student doesn't have this subject in his program", http.StatusUnauthorized)
		return
	}

	if studentGrade.Grade != 0 && studentGrade.HasAttended == false {
		studentGrade.HasAttended = true
	}

	updateGradeAndAttendanceOfAStudentQuery := `
		UPDATE student_grades 
		SET student_id = $1, subject_id = $2, grade = $3, has_attended = $4 
		WHERE id = $5;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		updateGradeAndAttendanceOfAStudentQuery,
		studentGrade.StudentID,
		studentGrade.SubjectID,
		studentGrade.Grade,
		studentGrade.HasAttended,
		studentGrade.ID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("UpdateGradeAndAttendanceOfAStudent QueryRowContext deadline exceeded: ", err)
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

	resp, err := json.Marshal("Updating Grade And Attendance Of A Student Successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("UpdateGradeAndAttendanceOfAStudent failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func DeleteGradeAndAttendanceOfAStudent(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "You can only delete grades for subjects that you teach", http.StatusUnauthorized)
		return
	}

	hasGroup, err := professorHasGroup(claims.Issuer, studentID)
	if err != nil {
		log.Println("professorHasGroup error: ", err)
		http.Error(w, "Error while checking professor privileges", http.StatusInternalServerError)
		return
	}

	if !hasGroup {
		http.Error(w, "You can only delete grades for students from groups that you teach", http.StatusUnauthorized)
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

	deleteGradeAndAttendanceOfAStudentQuery := `DELETE FROM student_grades WHERE id = $1 AND student_id = $2 AND subject_id = $3`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, deleteGradeAndAttendanceOfAStudentQuery, id, studentID, subjectID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("DeleteGradeAndAttendanceOfAStudent QueryRowContext deadline exceeded: ", err)
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

	resp, err := json.Marshal("Deleting Grade And Attendance Of A Student Successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("DeleteGradeAndAttendanceOfAStudent failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
