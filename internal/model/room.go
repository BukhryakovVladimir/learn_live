package model

type Room struct {
	ID          int    `json:"id,omitempty"`
	SubjectID   int    `json:"subject_id"`
	SubjectName string `json:"subject_name,omitempty"`
	RoomName    string `json:"room_name"`
}
