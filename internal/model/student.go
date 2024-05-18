package model

import "time"

type Student struct {
	ID        int       `json:"ID"`
	Username  string    `json:"username"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	GroupID   int       `json:"group_id"`
	Sex       string    `json:"sex"`
	BirthDate time.Time `json:"birthdate"`
}
