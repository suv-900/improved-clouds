package routers

import (
	"github.com/gorilla/mux"
	"github.com/suv-900/blog/controllers"
)

func HandleRoutes(router *mux.Router) {
	router.HandleFunc("/register", controllers.CreateUser).Methods("POST")
	router.HandleFunc("/login", controllers.LoginUser).Methods("POST")
	router.HandleFunc("/checkusername", controllers.SearchUsername).Methods("POST")
	//	router.HandleFunc("/profile", controllers.CreatePost).Methods("POST")
	router.HandleFunc("/serverstatus", controllers.CheckServerHealth).Methods("GET")
	router.HandleFunc("/createpost", controllers.CreatePost).Methods("POST")
	// router.HandleFunc("/viewpost", controllers.PostViewer).Methods("GET")
	// router.HandleFunc("/pollme", controllers.PollMe).Methods("GET")
	// router.HandleFunc("/checkcon", controllers.CheckConcurr).Methods("GET")
}
