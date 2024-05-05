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

func ListProfessorsSubjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	listProfessorsSubjectsQuery := `
	SELECT ps.professor_id, p.firstname, p.lastname, p.email,
	       p.phone_number, p.sex, p.birthdate, ps.subject_id, s.subject_name
	FROM professor_subject ps
	JOIN person p ON ps.professor_id = p.id
    JOIN subject s ON ps.subject_id = s.id
	ORDER BY ps.professor_id;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listProfessorsSubjectsQuery)
	defer rows.Close()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListProfessorsSubjects QueryRowContext deadline exceeded: ", err)
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

	var professors []model.ProfessorSubject
	var currentProfessor *model.ProfessorSubject

	for rows.Next() {
		var professor model.ProfessorSubject
		var subject model.Subject

		if err := rows.Scan(&professor.ProfessorID, &professor.FirstName, &professor.LastName, &professor.Email,
			&professor.PhoneNumber, &professor.Sex, &professor.BirthDate, &subject.ID, &subject.SubjectName); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if currentProfessor != nil && professor.ProfessorID == currentProfessor.ProfessorID {
			currentProfessor.Subjects = append(currentProfessor.Subjects, subject)
		} else {
			if currentProfessor != nil {
				professors = append(professors, *currentProfessor)
			}

			professor.Subjects = []model.Subject{subject}
			currentProfessor = &professor
		}
	}

	// Аппенд последнего професосра после выхода из цикла
	if currentProfessor != nil {
		professors = append(professors, *currentProfessor)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(professors)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List Professors and Subjects failed: %v\n", err)
	}
}

func ListSubjectsOfAProfessors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	paramProfessorID := r.URL.Query().Get("professor_id")

	professorID, err := strconv.Atoi(paramProfessorID)
	if err != nil {
		http.Error(w, "professor_id must be an integer", http.StatusBadRequest)
		return
	}

	listSubjectsOfAProfessorsQuery := `
	SELECT ps.professor_id, p.firstname, p.lastname, p.email,
	       p.phone_number, p.sex, p.birthdate, ps.subject_id, s.subject_name
	FROM professor_subject ps
	JOIN person p ON ps.professor_id = p.id
    JOIN subject s ON ps.subject_id = s.id
	WHERE ps.professor_id = $1
	ORDER BY ps.professor_id;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listSubjectsOfAProfessorsQuery, professorID)
	defer rows.Close()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListSubjectsOfAProfessors QueryRowContext deadline exceeded: ", err)
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

	var professors []model.ProfessorSubject
	var currentProfessor *model.ProfessorSubject

	for rows.Next() {
		var professor model.ProfessorSubject
		var subject model.Subject

		if err := rows.Scan(&professor.ProfessorID, &professor.FirstName, &professor.LastName, &professor.Email,
			&professor.PhoneNumber, &professor.Sex, &professor.BirthDate, &subject.ID, &subject.SubjectName); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if currentProfessor != nil && professor.ProfessorID == currentProfessor.ProfessorID {
			currentProfessor.Subjects = append(currentProfessor.Subjects, subject)
		} else {
			if currentProfessor != nil {
				professors = append(professors, *currentProfessor)
			}

			professor.Subjects = []model.Subject{subject}
			currentProfessor = &professor
		}
	}

	// Аппенд последнего професосра после выхода из цикла
	if currentProfessor != nil {
		professors = append(professors, *currentProfessor)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(professors)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List Subjects Of A Professors failed: %v\n", err)
	}
}

func ListProfessorsThatHaveASubject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	paramSubjectID := r.URL.Query().Get("subject_id")

	subjectID, err := strconv.Atoi(paramSubjectID)
	if err != nil {
		http.Error(w, "subject_id must be an integer", http.StatusBadRequest)
		return
	}

	listProfessorsThatHaveASubjectQuery := `
	SELECT ps.professor_id, p.firstname, p.lastname, p.email,
	       p.phone_number, p.sex, p.birthdate, ps.subject_id, s.subject_name
	FROM professor_subject ps
	JOIN person p ON ps.professor_id = p.id
    JOIN subject s ON ps.subject_id = s.id
	WHERE ps.subject_id = $1
	ORDER BY ps.professor_id;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listProfessorsThatHaveASubjectQuery, subjectID)
	defer rows.Close()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListProfessorsThatHaveASubject QueryRowContext deadline exceeded: ", err)
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

	var professors []model.ProfessorSubject
	var currentProfessor *model.ProfessorSubject

	for rows.Next() {
		var professor model.ProfessorSubject
		var subject model.Subject

		if err := rows.Scan(&professor.ProfessorID, &professor.FirstName, &professor.LastName, &professor.Email,
			&professor.PhoneNumber, &professor.Sex, &professor.BirthDate, &subject.ID, &subject.SubjectName); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if currentProfessor != nil && professor.ProfessorID == currentProfessor.ProfessorID {
			currentProfessor.Subjects = append(currentProfessor.Subjects, subject)
		} else {
			if currentProfessor != nil {
				professors = append(professors, *currentProfessor)
			}

			professor.Subjects = []model.Subject{subject}
			currentProfessor = &professor
		}
	}

	// Аппенд последнего професосра после выхода из цикла
	if currentProfessor != nil {
		professors = append(professors, *currentProfessor)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(professors)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("List Professors That Have A Subject failed: %v\n", err)
	}
}

func AddProfessorSubject(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "You do not have administrator privileges to add professor_subject relation", http.StatusUnauthorized)
		return
	}

	var professorSubject model.IUDProfessorSubject
	err = json.NewDecoder(r.Body).Decode(&professorSubject)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	insertProfessorSubjectQuery := `INSERT INTO professor_subject (professor_id, subject_id) VALUES ($1, $2);`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, insertProfessorSubjectQuery, professorSubject.ProfessorID, professorSubject.SubjectID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("AddProfessorSubject QueryRowContext deadline exceeded: ", err)
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

	resp, err := json.Marshal("Inserting professor subject relation successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("AddProfessorSubject failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func UpdateProfessorSubject(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "You do not have administrator privileges to update professor_subject relation", http.StatusUnauthorized)
		return
	}

	var updateProfessorSubject model.IUDProfessorSubject
	err = json.NewDecoder(r.Body).Decode(&updateProfessorSubject)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	updateProfessorSubjectQuery := `UPDATE professor_subject SET professor_id = $1, subject_id = $2
	WHERE professor_id = $3 AND subject_id = $4;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, updateProfessorSubjectQuery, updateProfessorSubject.ProfessorID,
		updateProfessorSubject.SubjectID, updateProfessorSubject.OldProfessorID, updateProfessorSubject.OldSubjectID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("updateProfessorSubject QueryRowContext deadline exceeded: ", err)
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
			log.Println("Unique key violation, professor_subject relation already exists: ", err)
			http.Error(w, "Professor Subject relation already exists", http.StatusGatewayTimeout)
			return
		}

		log.Println("Database error: ", pgErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal("Updating professor subject relation successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("updateProfessorSubject failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func DeleteProfessorSubject(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "You do not have administrator privileges to delete professor subject relation", http.StatusUnauthorized)
		return
	}

	ParamProfessorID := r.URL.Query().Get("professor_id")
	ParamSubjectID := r.URL.Query().Get("subject_id")

	professorID, err := strconv.Atoi(ParamProfessorID)
	if err != nil {
		http.Error(w, "professor_id must be an integer number", http.StatusBadRequest)
		return
	}

	subjectID, err := strconv.Atoi(ParamSubjectID)
	if err != nil {
		http.Error(w, "subject_id must be an integer number", http.StatusBadRequest)
		return
	}

	deleteProfessorSubjectQuery := `DELETE FROM professor_subject WHERE professor_id = $1 and subject_id = $2;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, deleteProfessorSubjectQuery, professorID, subjectID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("DeleteProfessorSubject QueryRowContext deadline exceeded: ", err)
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

	resp, err := json.Marshal("Deleting professor subject relation successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Delete Professor Subject failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
