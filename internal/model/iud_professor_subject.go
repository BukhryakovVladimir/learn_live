package model

type IUDProfessorSubject struct {
	ProfessorID    int `json:"professor_id"`
	SubjectID      int `json:"subject_id"`
	OldProfessorID int `json:"old_professor_id,omitempty"`
	OldSubjectID   int `json:"old_subject_id,omitempty"`
}
