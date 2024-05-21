package model

type UpdateGroupSubject struct {
	OldGroupID   int `json:"old_group_id"`
	OldSubjectID int `json:"old_subject_id"`
	NewGroupID   int `json:"new_group_id"`
	NewSubjectID int `json:"new_subject_id"`
}
