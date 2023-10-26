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
	router.HandleFunc("/authtoken", controllers.CreatePost).Methods("POST")
	router.HandleFunc("/deleteuser", controllers.DeleteUser).Methods("DELETE")
	router.HandleFunc("/addcomment/{id:[0-9]+}", controllers.AddComment).Methods("POST")
	router.HandleFunc("/viewpost/{id:[0-9]+}", controllers.GetPostByID).Methods("GET")
	router.HandleFunc("/likecomment", controllers.LikeComment).Methods("POST")
	router.HandleFunc("/dislikecomment", controllers.DislikeComment).Methods("POST")
	router.HandleFunc("/likepost/{postid:[0-9]+}", controllers.LikePost).Methods("POST")
	router.HandleFunc("/dislikepost/{postid:[0-9]+}", controllers.DislikePost).Methods("POST")
}
