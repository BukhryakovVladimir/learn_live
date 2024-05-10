package learn_live_handler

import (
	"github.com/BukhryakovVladimir/learn_live/internal/routes"
	"net/http"
)

func SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/signup", routes.SignupPerson)
	//list students for professors and admins
	//list professors for all users
	//list admins is admin and professor privilege
	//delete person
	//update person
	mux.HandleFunc("/api/login", routes.LoginPerson)

	mux.HandleFunc("/api/list-subjects", routes.ListSubjects)
	mux.HandleFunc("/api/add-subject", routes.AddSubject)
	mux.HandleFunc("/api/update-subject", routes.UpdateSubject)
	mux.HandleFunc("/api/delete-subject", routes.DeleteSubject)

	mux.HandleFunc("/api/list-rooms", routes.ListRooms)
	mux.HandleFunc("/api/list-rooms-of-a-subject", routes.ListRoomsOfASubject)
	mux.HandleFunc("/api/add-room", routes.AddRoom)
	mux.HandleFunc("/api/update-room", routes.UpdateRoom)
	mux.HandleFunc("/api/delete-room", routes.DeleteRoom)

	mux.HandleFunc("/api/list-groups", routes.ListGroups)
	mux.HandleFunc("/api/add-group", routes.AddGroup)
	mux.HandleFunc("/api/update-group", routes.UpdateGroup)
	mux.HandleFunc("/api/delete-group", routes.DeleteGroup)

	mux.HandleFunc("/api/list-groups-and-subjects-relations", routes.ListGroupsSubjects)      // list both ids and names
	mux.HandleFunc("/api/list-subjects-of-a-group", routes.ListSubjectsOfAGroup)              // list both ids and names
	mux.HandleFunc("/api/list-groups-that-have-a-subject", routes.ListGroupsThatHaveASubject) // list both ids and names
	mux.HandleFunc("/api/add-groups-and-subjects-relation", routes.AddGroupSubject)           // just insert
	mux.HandleFunc("/api/update-groups-and-subjects-relation", routes.UpdateGroupSubject)     // just update set names by id
	mux.HandleFunc("/api/delete-groups-and-subjects-relation", routes.DeleteGroupSubject)     // just delete by ids (remember, no body)

	mux.HandleFunc("/api/list-professors-and-subjects-relations", routes.ListProfessorsSubjects)      // list both ids and names
	mux.HandleFunc("/api/list-subjects-of-a-professor", routes.ListSubjectsOfAProfessors)             // list both ids and names
	mux.HandleFunc("/api/list-professors-that-have-a-subject", routes.ListProfessorsThatHaveASubject) // list both ids and names
	mux.HandleFunc("/api/add-professors-and-subjects-relation", routes.AddProfessorSubject)           // just insert
	mux.HandleFunc("/api/update-professors-and-subjects-relation", routes.UpdateProfessorSubject)     // just update set names by id
	mux.HandleFunc("/api/delete-professors-and-subjects-relation", routes.DeleteProfessorSubject)     // just delete by ids (remember, no body)

	mux.HandleFunc("/api/list-professors-and-groups-relations", routes.ListProfessorsGroups)      // list both ids and names
	mux.HandleFunc("/api/list-groups-of-a-professor", routes.ListGroupsOfAProfessors)             // list both ids and names
	mux.HandleFunc("/api/list-professors-that-have-a-group", routes.ListProfessorsThatHaveAGroup) // list both ids and names
	mux.HandleFunc("/api/add-professors-and-groups-relation", routes.AddProfessorGroup)           // just insert
	mux.HandleFunc("/api/update-professors-and-groups-relation", routes.UpdateProfessorGroup)     // just update set names by id
	mux.HandleFunc("/api/delete-professors-and-groups-relation", routes.DeleteProfessorGroup)     // just delete by ids (remember, no body)
}
