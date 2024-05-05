package model

import "time"

type ProfessorSubject struct {
	ProfessorID int       `json:"professor_id"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	Sex         string    `json:"sex"`
	BirthDate   time.Time `json:"birthdate"`
	Subjects    []Subject `json:"subjects,omitempty"`
}
