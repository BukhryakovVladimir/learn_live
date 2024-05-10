package model

type IUDProfessorGroup struct {
	ProfessorID    int `json:"professor_id"`
	GroupID        int `json:"group_id"`
	OldProfessorID int `json:"old_professor_id,omitempty"`
	OldGroupID     int `json:"old_group_id,omitempty"`
}
