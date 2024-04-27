package learn_live_handler

import (
	"github.com/BukhryakovVladimir/learn_live/internal/routes"
	"net/http"
)

func SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/signup", routes.SignupPerson)
	//delete person
	//update person
	mux.HandleFunc("/api/login", routes.LoginPerson)

	mux.HandleFunc("/api/list-subjects", routes.ListSubjects)
	mux.HandleFunc("/api/add-subject", routes.AddSubject)
	mux.HandleFunc("/api/update-subject", routes.UpdateSubject)
	mux.HandleFunc("/api/delete-subject", routes.DeleteSubject)

	mux.HandleFunc("/api/list-groups", routes.ListGroups)
	mux.HandleFunc("/api/add-group", routes.AddGroup)
	mux.HandleFunc("/api/update-group", routes.UpdateGroup)
	mux.HandleFunc("/api/delete-group", routes.DeleteGroup)
}
