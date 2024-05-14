package model

type StudentGrade struct {
	ID               int    `json:"id,omitempty"`
	StudentID        int    `json:"student_id,omitempty"`
	StudentFirstname string `json:"student_firstname,omitempty"`
	StudentLastname  string `json:"student_lastname,omitempty"`
	StudentGroupID   int    `json:"student_group_id,omitempty"`
	StudentGroupName string `json:"student_group_name,omitempty"`
	SubjectID        int    `json:"subject_id"`
	SubjectName      string `json:"subject_name,omitempty"`
	Grade            int    `json:"grade"`
	HasAttended      bool   `json:"has_attended"`
}
