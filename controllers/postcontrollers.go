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
	var tokenExpired bool
	var tokenInvalid bool
	a := make(chan int, 1)
	go func() {
		tokenExpired, authorid, tokenInvalid = AuthenticateTokenAndSendUserID(r)
		a <- 1
	}()
	<-a

	if tokenExpired {
		w.WriteHeader(401)
		return
	}
	if tokenInvalid {
		w.WriteHeader(400)
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
	var tokenExpired bool
	var tokenInvalid bool
	a := make(chan int, 1)
	go func() {
		tokenExpired, _, tokenInvalid = AuthenticateTokenAndSendUserID(r)
		a <- 1
	}()
	<-a

	if tokenInvalid {
		w.WriteHeader(400)
		return
	}
	if tokenExpired {
		w.WriteHeader(401)
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
	var tokenExpired bool
	var tokenInvalid bool
	a := make(chan int, 1)
	go func() {
		tokenExpired, _, tokenInvalid = AuthenticateTokenAndSendUserID(r)
		a <- 1
	}()
	<-a

	if tokenExpired {
		w.WriteHeader(401)
		return
	}
	if tokenInvalid {
		w.WriteHeader(400)
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
		post, username, err = models.PostById(postid)
		a <- 1
	}()
	<-a
	if err != nil {
		serverError(&w, err)
		return
	}

	b := make(chan int, 1)
	var comments []models.UsernameAndComment
	go func() {
		comments = models.GetAllCommentsByPostID(postid)
		b <- 1
	}()
	<-b

	c := make(chan int, 1)
	var parsedRes []byte
	finalRes := models.PostUsernameComments{Username: username, Post: post, Comments: comments}
	go func() {
		parsedRes, err = json.Marshal(finalRes)
		c <- 1
	}()
	<-c
	if err != nil {
		serverError(&w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(parsedRes)

}
func GetPost_ByID_WithToken(w http.ResponseWriter, r *http.Request) {
	var postidstr string
	var postid uint64
	vars := mux.Vars(r)
	postidstr = vars["id"]
	postid, err := strconv.ParseUint(postidstr, 10, 64)
	if err != nil {
		serverError(&w, err)
		return
	}

	var userid uint64
	var tokenExpired bool
	var tokenInvalid bool
	a := make(chan int, 1)
	go func() {
		tokenExpired, userid, tokenInvalid = AuthenticateTokenAndSendUserID(r)
		a <- 1
	}()
	if tokenExpired {
		w.WriteHeader(401)
		return
	}
	if tokenInvalid {
		w.WriteHeader(400)
		return
	}

	b := make(chan int, 1)
	var finalResult models.PostUsernameComments_WithUserPreference
	go func() {
		finalResult.PostAndUserPreferences, err = models.GetPostAndUserPreferences(postid, userid)
		b <- 1
	}()
	<-b
	if err != nil {
		serverError(&w, err)
		return
	}

	c := make(chan int, 1)
	go func() {
		finalResult.Comments, err = models.Get5CommentsByPostID(postid)
		c <- 1
	}()
	<-c
	if err != nil {
		serverError(&w, err)
		return
	}

	var jsonReply []byte
	jsonReply, err = json.Marshal(finalResult)
	if err != nil {
		serverError(&w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(jsonReply)

}
func LikePost(w http.ResponseWriter, r *http.Request) {
	var postid uint64
	var err error
	var tokenInvalid bool
	var tokenExpired bool
	var userid uint64
	a := make(chan int, 1)
	go func() {
		tokenExpired, userid, tokenInvalid = AuthenticateTokenAndSendUserID(r)
		a <- 1
	}()
	<-a

	if tokenExpired {
		w.WriteHeader(401)
		return
	}
	if tokenInvalid {
		w.WriteHeader(400)
		return
	}
	vars := mux.Vars(r)
	postidstr := vars["postid"]
	postid, err = strconv.ParseUint(postidstr, 10, 64)
	if err != nil {
		serverError(&w, err)
		return
	}

	//save user prefrence
	b := make(chan int, 1)
	go func() {
		err = models.LikePostByID(userid, postid)
		b <- 1
	}()
	<-b
	if err != nil {
		serverError(&w, err)
		return
	}
	w.WriteHeader(200)
}

func DislikePost(w http.ResponseWriter, r *http.Request) {
	var postid uint64
	var err error
	var tokenExpired bool
	var tokenInvalid bool
	var userid uint64
	a := make(chan int, 1)
	go func() {
		tokenExpired, userid, tokenInvalid = AuthenticateTokenAndSendUserID(r)
		a <- 1

		a <- 1
	}()
	<-a
	if tokenExpired {
		w.WriteHeader(401)
	}
	if tokenInvalid {
		w.WriteHeader(400)
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

	//save user prefrence
	b := make(chan int, 1)
	go func() {
		err = models.DislikePostByID(userid, postid)
		b <- 1
	}()
	<-b
	if err != nil {
		serverError(&w, err)
		return
	}
	w.WriteHeader(200)
}

func RemoveLikeFromPost(w http.ResponseWriter, r *http.Request) {
	var postid uint64
	var userid uint64
	var err error
	var tokenExpired bool
	var tokenInvalid bool
	a := make(chan int, 1)
	go func() {
		tokenExpired, userid, tokenInvalid = AuthenticateTokenAndSendUserID(r)
		a <- 1

		a <- 1
	}()
	<-a
	if tokenExpired {
		w.WriteHeader(401)
	}

	if tokenInvalid {
		w.WriteHeader(400)
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

	b := make(chan int, 1)
	go func() {
		err = models.RemoveLikeFromPost(userid, postid)
		b <- 1
	}()
	<-b
	if err != nil {
		serverError(&w, err)
		return
	}

	w.WriteHeader(200)
}
func RemoveDislikeFromPost(w http.ResponseWriter, r *http.Request) {
	var postid uint64
	var err error
	var tokenExpired bool
	var tokenInvalid bool
	var userid uint64
	a := make(chan int, 1)
	go func() {
		tokenExpired, userid, tokenInvalid = AuthenticateTokenAndSendUserID(r)
		a <- 1

		a <- 1
	}()
	<-a
	if tokenExpired {
		w.WriteHeader(401)
	}

	if tokenInvalid {
		w.WriteHeader(400)
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

	b := make(chan int, 1)
	go func() {
		err = models.RemoveDislikeFromPost(userid, postid)
		b <- 1
	}()
	<-b
	if err != nil {
		serverError(&w, err)
		return
	}

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
