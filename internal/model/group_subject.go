package model

type GroupSubject struct {
	GroupID     int    `json:"group_id"`
	SubjectID   int    `json:"subject_id"`
	GroupName   string `json:"group_name"`
	SubjectName string `json:"subject_name"`
}
