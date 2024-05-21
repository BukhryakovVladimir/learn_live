package model

import "time"

type Person struct {
	ProfessorID int       `json:"professor_id"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	GroupID     int       `json:"group_id"`
	IsProfessor bool      `json:"is_professor"`
	IsAdmin     bool      `json:"is_admin"`
	Sex         string    `json:"sex"`
	BirthDate   time.Time `json:"birthdate"`
}
