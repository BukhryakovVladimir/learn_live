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

func ListProfessorsGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	listProfessorsGroupsQuery := `
	SELECT pg.professor_id, p.firstname, p.lastname, p.email,
	       p.phone_number, p.sex, p.birthdate, pg.group_id, g.group_name
	FROM professor_group pg
	JOIN person p ON pg.professor_id = p.id
    JOIN group_uni g ON pg.group_id = g.id
	ORDER BY pg.professor_id;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listProfessorsGroupsQuery)
	defer rows.Close()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListProfessorsGroups QueryRowContext deadline exceeded: ", err)
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

	var professors []model.ProfessorGroup
	var currentProfessor *model.ProfessorGroup

	for rows.Next() {
		var professor model.ProfessorGroup
		var group model.Group

		if err := rows.Scan(&professor.ProfessorID, &professor.FirstName, &professor.LastName, &professor.Email,
			&professor.PhoneNumber, &professor.Sex, &professor.BirthDate, &group.ID, &group.GroupName); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if currentProfessor != nil && professor.ProfessorID == currentProfessor.ProfessorID {
			currentProfessor.Groups = append(currentProfessor.Groups, group)
		} else {
			if currentProfessor != nil {
				professors = append(professors, *currentProfessor)
			}

			professor.Groups = []model.Group{group}
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
		log.Printf("List Professors and Groups failed: %v\n", err)
	}
}

func ListGroupsOfAProfessors(w http.ResponseWriter, r *http.Request) {
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

	listGroupsOfAProfessorsQuery := `
	SELECT pg.professor_id, p.firstname, p.lastname, p.email,
	       p.phone_number, p.sex, p.birthdate, pg.group_id, g.group_name
	FROM professor_group pg
	JOIN person p ON pg.professor_id = p.id
    JOIN group_uni g ON pg.group_id = g.id
	WHERE pg.professor_id = $1
	ORDER BY pg.professor_id;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listGroupsOfAProfessorsQuery, professorID)
	defer rows.Close()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListGroupsOfAProfessors QueryRowContext deadline exceeded: ", err)
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

	var professors []model.ProfessorGroup
	var currentProfessor *model.ProfessorGroup

	for rows.Next() {
		var professor model.ProfessorGroup
		var group model.Group

		if err := rows.Scan(&professor.ProfessorID, &professor.FirstName, &professor.LastName, &professor.Email,
			&professor.PhoneNumber, &professor.Sex, &professor.BirthDate, &group.ID, &group.GroupName); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if currentProfessor != nil && professor.ProfessorID == currentProfessor.ProfessorID {
			currentProfessor.Groups = append(currentProfessor.Groups, group)
		} else {
			if currentProfessor != nil {
				professors = append(professors, *currentProfessor)
			}

			professor.Groups = []model.Group{group}
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
		log.Printf("List Groups Of A Professors failed: %v\n", err)
	}
}

func ListProfessorsThatHaveAGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	paramGroupID := r.URL.Query().Get("group_id")

	groupID, err := strconv.Atoi(paramGroupID)
	if err != nil {
		http.Error(w, "group_id must be an integer", http.StatusBadRequest)
		return
	}

	listProfessorsThatHaveAGroupQuery := `
	SELECT pg.professor_id, p.firstname, p.lastname, p.email,
	       p.phone_number, p.sex, p.birthdate, pg.group_id, g.group_name
	FROM professor_group pg
	JOIN person p ON pg.professor_id = p.id
    JOIN group_uni g ON pg.group_id = g.id
	WHERE pg.group_id = $1
	ORDER BY pg.professor_id;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, listProfessorsThatHaveAGroupQuery, groupID)
	defer rows.Close()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("ListProfessorsThatHaveAGroup QueryRowContext deadline exceeded: ", err)
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

	var professors []model.ProfessorGroup
	var currentProfessor *model.ProfessorGroup

	for rows.Next() {
		var professor model.ProfessorGroup
		var group model.Group

		if err := rows.Scan(&professor.ProfessorID, &professor.FirstName, &professor.LastName, &professor.Email,
			&professor.PhoneNumber, &professor.Sex, &professor.BirthDate, &group.ID, &group.GroupName); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if currentProfessor != nil && professor.ProfessorID == currentProfessor.ProfessorID {
			currentProfessor.Groups = append(currentProfessor.Groups, group)
		} else {
			if currentProfessor != nil {
				professors = append(professors, *currentProfessor)
			}

			professor.Groups = []model.Group{group}
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
		log.Printf("List Professors That Have A Group failed: %v\n", err)
	}
}

func AddProfessorGroup(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "You do not have administrator privileges to add professor_group relation", http.StatusUnauthorized)
		return
	}

	var professorGroup model.IUDProfessorGroup
	err = json.NewDecoder(r.Body).Decode(&professorGroup)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	insertProfessorGroupQuery := `INSERT INTO professor_group (professor_id, group_id) VALUES ($1, $2);`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, insertProfessorGroupQuery, professorGroup.ProfessorID, professorGroup.GroupID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("AddProfessorGroup QueryRowContext deadline exceeded: ", err)
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
			log.Println("Unique key violation, professor_group relation already exists: ", err)
			http.Error(w, "Professor Group relation already exists", http.StatusGatewayTimeout)
			return
		}

		log.Println("Database error: ", pgErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal("Inserting professor group relation successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("AddProfessorGroup failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func UpdateProfessorGroup(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "You do not have administrator privileges to update professor_group relation", http.StatusUnauthorized)
		return
	}

	var updateProfessorGroup model.IUDProfessorGroup
	err = json.NewDecoder(r.Body).Decode(&updateProfessorGroup)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	updateProfessorGroupQuery := `UPDATE professor_group SET professor_id = $1, group_id = $2
	WHERE professor_id = $3 AND group_id = $4;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, updateProfessorGroupQuery, updateProfessorGroup.ProfessorID,
		updateProfessorGroup.GroupID, updateProfessorGroup.OldProfessorID, updateProfessorGroup.OldGroupID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("UpdateProfessorGroup QueryRowContext deadline exceeded: ", err)
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
			log.Println("Unique key violation, professor_group relation already exists: ", err)
			http.Error(w, "Professor Group relation already exists", http.StatusGatewayTimeout)
			return
		}

		log.Println("Database error: ", pgErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal("Updating professor group relation successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("updateProfessorGroup failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func DeleteProfessorGroup(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "You do not have administrator privileges to delete professor group relation", http.StatusUnauthorized)
		return
	}

	ParamProfessorID := r.URL.Query().Get("professor_id")
	ParamGroupID := r.URL.Query().Get("group_id")

	professorID, err := strconv.Atoi(ParamProfessorID)
	if err != nil {
		http.Error(w, "professor_id must be an integer number", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(ParamGroupID)
	if err != nil {
		http.Error(w, "group_id must be an integer number", http.StatusBadRequest)
		return
	}

	deleteProfessorGroupQuery := `DELETE FROM professor_group WHERE professor_id = $1 and group_id = $2;`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, deleteProfessorGroupQuery, professorID, groupID)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("DeleteProfessorGroup QueryRowContext deadline exceeded: ", err)
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

	resp, err := json.Marshal("Deleting professor group relation successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Delete Professor Group failed: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
