package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	//"strconv"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/suv-900/blog/models"
)

//TODO no jwt
/*
->CRUD
->getbyid
*/
func GetallpostsbyUser(w http.ResponseWriter, r *http.Request) {
	var userId uint64
	ok, p := TokenVerifier("userToken", r)
	if ok {
		userId = p.ID
	} else {
		w.WriteHeader(401)
		return
	}

	posts := models.GetPostsByUserId(userId)

	parsedRes, err := json.Marshal(posts)
	if err != nil {
		serverError(&w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(parsedRes)
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	var authorid uint64

	pipe1 := make(chan bool, 1)
	go func() {
		ok, userid := AuthenticateTokenAndSendUserID(&w, r)
		if ok {
			authorid = userid
			pipe1 <- true
			return
		}
		pipe1 <- false
		fmt.Printf("Error while parsing token.")
	}()
	if !<-pipe1 {
		//error codes are getting managed by AuthHandler
		return
	}

	pipe2 := make(chan bool, 1)
	var post models.Posts

	var postid uint64
	go func() {
		rbyte, err := io.ReadAll(r.Body)
		if err != nil {
			serverError(&w, err)
			fmt.Println("error while reading request.")
			pipe2 <- false
			return
		}
		err = json.Unmarshal(rbyte, &post)
		if err != nil {
			serverError(&w, err)
			fmt.Println("error while unmarshalling data")
			pipe2 <- false
			return
		}
		post.Author_id = authorid
		post.CreatedAt = time.Now()
		post.UpdatedAt = time.Now()
		postid, err = models.CreatePost(post)
		if err != nil {
			serverError(&w, err)
			fmt.Println("error while creating post")
			pipe2 <- false
			return
		}
		pipe2 <- true
	}()

	if !<-pipe2 {
		fmt.Println("post creation failed")
		return
	}
	w.WriteHeader(200)

	parsedres, err := json.Marshal(postid)
	if err != nil {
		serverError(&w, err)
		return
	}
	//b := make([]byte, 8)
	//binary.LittleEndian.PutUint64(b, postid)
	w.Write(parsedres)
	// var userPosts []models.Posts
	// userPosts = models.CreatePost(post.Username, post)
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	var postid uint64

	pipe1 := make(chan bool, 1)
	go func() {
		fmt.Println("verifying token.")
		ok, _ := AuthenticateTokenAndSendUserID(&w, r)
		if ok {
			pipe1 <- true
			return
		}
		pipe1 <- false
		fmt.Printf("Error while parsing token.")
	}()
	if !<-pipe1 {
		return
	}

	pipe2 := make(chan bool, 1)
	go func() {
		err := models.DeletePost(postid)
		if err != nil {
			serverError(&w, err)
			pipe2 <- false
			return
		}
		pipe2 <- true
	}()
	if !<-pipe2 {
		fmt.Println("error while deleting post")
		return
	}
	fmt.Println("post deleted succesfully")
	w.WriteHeader(200)
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	var postid uint64

	pipe1 := make(chan bool, 1)
	go func() {
		fmt.Println("verifying token.")
		ok, _ := AuthenticateTokenAndSendUserID(&w, r)
		if ok {
			pipe1 <- true
			return
		}
		pipe1 <- false
		fmt.Printf("Error while parsing token.")
	}()
	if !<-pipe1 {
		return
	}

	pipe2 := make(chan bool, 1)

	go func() {

		rbyte, err := io.ReadAll(r.Body)
		if err != nil {
			pipe2 <- false
			serverError(&w, err)
			return
		}

		var post models.Posts
		err = json.Unmarshal(rbyte, &post)
		if err != nil {
			pipe2 <- false
			serverError(&w, err)
			return
		}

		err = models.UpdatePost(postid, post)
		if err != nil {
			pipe2 <- false
			serverError(&w, err)
			return
		}
		pipe2 <- true

	}()
	if !<-pipe2 {
		fmt.Println("error while updating post")
		return
	}
	w.WriteHeader(200)
}

// sends post username and top 5 comments
func GetPostByID(w http.ResponseWriter, r *http.Request) {
	var postidstr string
	var postid uint64
	vars := mux.Vars(r)
	postidstr = vars["id"]
	postid, err := strconv.ParseUint(postidstr, 10, 64)
	if err != nil {
		serverError(&w, err)
		return
	}
	a := make(chan int, 1)
	var post models.Posts
	var username string
	go func() {
		post, username = models.PostById(postid)
		a <- 1
	}()
	<-a

	b := make(chan int, 1)
	var comments []models.UsernameAndComment
	go func() {
		comments = models.GetAllCommentsByPostID(postid)
		b <- 1
	}()
	<-b

	pipe2 := make(chan bool, 1)
	var parsedRes []byte
	finalRes := models.PostUsernameComments{Username: username, Post: post, Comments: comments}
	go func() {
		parsedRes, err = json.Marshal(finalRes)
		if err != nil {
			serverError(&w, err)
			pipe2 <- false
			return
		}
		pipe2 <- true
	}()
	if !<-pipe2 {
		fmt.Println("Error occured while parsing post to json")
		return
	}
	w.WriteHeader(200)
	w.Write(parsedRes)

}

func LikePost(w http.ResponseWriter, r *http.Request) {
	var ok bool
	var postid uint64
	var err error
	a := make(chan int, 1)
	go func() {
		ok, _ = AuthenticateTokenAndSendUserID(&w, r)
		if !ok {
			a <- 1
			return
		}

		vars := mux.Vars(r)
		postidstr := vars["postid"]
		postid, err = strconv.ParseUint(postidstr, 10, 64)
		if err != nil {
			serverError(&w, err)
			a <- 1
			return
		}

		a <- 1
	}()
	<-a
	if !ok {
		w.WriteHeader(400)
		return
	}
	if err != nil {
		return
	}

	//save user prefrence
	b := make(chan int, 1)
	go func() {
		models.LikePostByID(postid)
		b <- 1
	}()
	<-b

	w.WriteHeader(200)
}

func DislikePost(w http.ResponseWriter, r *http.Request) {
	var ok bool
	var postid uint64
	var err error
	a := make(chan int, 1)
	go func() {
		ok, _ = AuthenticateTokenAndSendUserID(&w, r)
		if !ok {
			a <- 1
			return
		}

		vars := mux.Vars(r)
		postidstr := vars["postid"]
		postid, err = strconv.ParseUint(postidstr, 10, 64)
		if err != nil {
			serverError(&w, err)
			a <- 1
			return
		}

		a <- 1
	}()
	<-a
	if !ok {
		w.WriteHeader(400)
		return
	}
	if err != nil {
		return
	}

	//save user prefrence
	b := make(chan int, 1)
	go func() {
		models.DislikePostByID(postid)
		b <- 1
	}()
	<-b

	w.WriteHeader(200)
}

func TokenVerifier(s string, r *http.Request) (bool, *CustomPayload) {
	t := GetCookieByName(r.Cookies(), s)
	if t == "" {
		fmt.Println("no cookie got.")
		return false, nil
	}
	token, err := jwt.ParseWithClaims(t, &CustomPayload{}, nil)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}
	if p, ok := token.Claims.(*CustomPayload); ok && token.Valid {
		return true, p
	} else {
		fmt.Println("Token not ok!")
		return false, nil
	}
}

func GetCookieByName(cookies []*http.Cookie, cookiename string) string {
	result := ""
	for i := 0; i < len(cookies); i++ {
		if cookies[i].Name == cookiename {
			result += cookies[i].Value
			break
		}
	}
	return result
}
