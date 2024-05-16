package model

type AdminOrProfessor struct {
	UserID      string `json:"user_id"`
	IsAdmin     bool   `json:"is_admin"`
	IsProfessor bool   `json:"is_professor"`
}
