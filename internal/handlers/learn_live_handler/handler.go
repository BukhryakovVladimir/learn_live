package learn_live_handler

import (
	"github.com/BukhryakovVladimir/learn_live/internal/routes"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(r chi.Router) {
	r.Use(corsMiddleware)

	r.Route("/api", func(r chi.Router) {
		r.Post("/signup", routes.SignupPerson)
		r.Get("/check-is-admin-or-professor", routes.CheckIsAdminOrProfessor)
		//list students for professors and admins
		//list professors for all users
		//list admins is admin and professor privilege
		//delete person
		//update person
		r.Post("/login", routes.LoginPerson)

		r.Get("/list-current-user-subjects", routes.ListCurrentUserSubjects) // list both ids and names
		r.Get("/list-subjects", routes.ListSubjects)
		r.Post("/add-subject", routes.AddSubject)
		r.Put("/update-subject", routes.UpdateSubject)
		r.Delete("/delete-subject", routes.DeleteSubject)

		r.Get("/list-rooms", routes.ListRooms)
		r.Get("/list-rooms-of-a-subject", routes.ListRoomsOfASubject)
		r.Post("/add-room", routes.AddRoom)
		r.Put("/update-room", routes.UpdateRoom)
		r.Delete("/delete-room", routes.DeleteRoom)

		r.Get("/list-groups", routes.ListGroups)
		r.Post("/add-group", routes.AddGroup)
		r.Put("/update-group", routes.UpdateGroup)
		r.Delete("/delete-group", routes.DeleteGroup)

		r.Get("/list-groups-and-subjects-relations", routes.ListGroupsSubjects)      // list both ids and names
		r.Get("/list-subjects-of-a-group", routes.ListSubjectsOfAGroup)              // list both ids and names
		r.Get("/list-groups-that-have-a-subject", routes.ListGroupsThatHaveASubject) // list both ids and names
		r.Post("/add-groups-and-subjects-relation", routes.AddGroupSubject)          // just insert
		r.Put("/update-groups-and-subjects-relation", routes.UpdateGroupSubject)     // just update set names by id
		r.Delete("/delete-groups-and-subjects-relation", routes.DeleteGroupSubject)  // just delete by ids (remember, no body)

		r.Get("/list-professors-and-subjects-relations", routes.ListProfessorsSubjects)      // list both ids and names
		r.Get("/list-subjects-of-a-professor", routes.ListSubjectsOfAProfessors)             // list both ids and names
		r.Get("/list-professors-that-have-a-subject", routes.ListProfessorsThatHaveASubject) // list both ids and names
		r.Post("/add-professors-and-subjects-relation", routes.AddProfessorSubject)          // just insert
		r.Put("/update-professors-and-subjects-relation", routes.UpdateProfessorSubject)     // just update set names by id
		r.Delete("/delete-professors-and-subjects-relation", routes.DeleteProfessorSubject)  // just delete by ids (remember, no body)

		r.Get("/list-professors-and-groups-relations", routes.ListProfessorsGroups)      // list both ids and names
		r.Get("/list-groups-of-a-professor", routes.ListGroupsOfAProfessors)             // list both ids and names
		r.Get("/list-professors-that-have-a-group", routes.ListProfessorsThatHaveAGroup) // list both ids and names
		r.Post("/add-professors-and-groups-relation", routes.AddProfessorGroup)          // just insert
		r.Put("/update-professors-and-groups-relation", routes.UpdateProfessorGroup)     // just update set names by id
		r.Delete("/delete-professors-and-groups-relation", routes.DeleteProfessorGroup)  // just delete by ids (remember, no body)

		r.Get("/list-current-user-grades-and-attendance", routes.ListCurrentUserGradesAndAttendance)
		r.Get("/list-grades-and-attendance-of-a-student", routes.ListGradesAndAttendanceOfAStudent)
		r.Get("/list-grades-and-attendance-of-a-group", routes.ListGradesAndAttendanceOfAGroup)
		r.Post("/insert-grade-and-attendance-of-a-student", routes.InsertGradeAndAttendanceOfAStudent)
		r.Put("/update-grade-and-attendance-of-a-student", routes.UpdateGradeAndAttendanceOfAStudent)
		r.Delete("/delete-grade-and-attendance-of-a-student", routes.DeleteGradeAndAttendanceOfAStudent)

		r.Get("/list-current-user-total-grades", routes.ListCurrentUserTotalGrades)
		r.Get("/list-total-grades-of-a-student", routes.ListTotalGradesOfAStudent)
		r.Get("/list-total-grades-of-a-group", routes.ListTotalGradesOfAGroup)
		r.Post("/insert-total-grade-of-a-student", routes.InsertTotalGradeOfAStudent)
		r.Put("/update-total-grade-of-a-student", routes.UpdateTotalGradeOfAStudent)
		r.Delete("/delete-total-grade-of-a-student", routes.DeleteTotalGradeOfAStudent)

		r.Get("/get-token", routes.GetToken)

	})

}
