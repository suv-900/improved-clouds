package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	//"strconv"

	"github.com/golang-jwt/jwt"
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
		userId = p.id
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
		fmt.Println("verifying token.")
		ok, userid := AuthenticateToken(&w, r)
		if ok {
			authorid = userid
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
	var post models.Posts

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
		post.Authorid = authorid

		_, err = models.CreatePost(post)
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
	// var userPosts []models.Posts
	// userPosts = models.CreatePost(post.Username, post)
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	var postid uint64

	pipe1 := make(chan bool, 1)
	go func() {
		fmt.Println("verifying token.")
		ok, _ := AuthenticateToken(&w, r)
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
		ok, _ := AuthenticateToken(&w, r)
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

func PostViewer(w http.ResponseWriter, r *http.Request) {

	var postid uint64
	p := r.URL.Query().Get("postid")
	if p == "" {
		w.WriteHeader(400)
		return
	}
	postid, err := strconv.ParseUint(p, 10, 64)
	if err != nil {
		serverError(&w, err)
		return
	}

	pipe1 := make(chan int, 1)
	var post models.Posts
	var username string
	go func() {
		post, username = models.PostById(postid)
		pipe1 <- 1
	}()
	<-pipe1

	pipe2 := make(chan bool, 1)
	var parsedRes []byte
	go func() {
		parsedRes, err = json.Marshal(models.UsernameAndPost{Username: username, Post: post})
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
