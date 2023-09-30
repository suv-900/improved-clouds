package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/suv-900/blog/models"
	"golang.org/x/crypto/bcrypt"
)

//TODO hashing is taking too long

var bycryptCost = 15
var JWTKEY = []byte(os.Getenv("JWT_KEY"))
var Tokenexpirytime = time.Now().Add(10 * time.Minute)

type CustomPayload struct {
	id uint64
	jwt.StandardClaims
}

func CheckServerHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

}

// completed
// TODO unit tests
func CreateUser(w http.ResponseWriter, r *http.Request) {
	rbyte, err := io.ReadAll(r.Body)
	if err != nil {
		serverError(&w, err)
		return
	}
	var user models.Users
	err = json.Unmarshal(rbyte, &user)
	if err != nil {
		serverError(&w, err)
		return
	}

	pipe1 := make(chan bool, 1)
	pipe2 := make(chan bool, 1)

	go func() {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bycryptCost)
		if err != nil {
			pipe1 <- false
			serverError(&w, err)
			return
		}

		user.Password = string(hashedPassword)
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		fmt.Println(user)
		pipe1 <- true
	}()

	flag := make(chan uint, 1)
	tokenChannel := make(chan string, 1)

	go func() {

		if !<-pipe1 {
			fmt.Println("hashingPassword stage unsuccessful.")
			pipe2 <- false
			serverError(&w, nil)
			return
		}
		//TODO add user with same name
		//TODO i dont think this call gets awaited/the goroutine waits for this call

		//var userCreatewg sync.WaitGroup

		var userFound bool

		findUser := make(chan bool, 1)
		go func() {
			userFound = models.FindUser(user.Username)
			findUser <- userFound
		}()

		if <-findUser {
			flag <- 409
			pipe2 <- false
			return
		}
		flag <- 0 //so that we dont listen on dead channel

		//var id uint64
		//var err error

		//userCreatewg.Add(1)
		createUserpipe := make(chan uint64, 1)
		createUserErrorpipe := make(chan error, 1)

		go func() {
			//defer userCreatewg.Done()
			id, err := models.CreateUser(user)
			createUserpipe <- id
			createUserErrorpipe <- err
		}()
		//userCreatewg.Wait()
		id := <-createUserpipe
		err = <-createUserErrorpipe
		if err != nil {
			fmt.Println("usercreation failed.")
			serverError(&w, err)
			pipe2 <- false
			return
		}

		p := CustomPayload{
			id,
			jwt.StandardClaims{
				ExpiresAt: Tokenexpirytime.Unix(),
				Issuer:    "createUser handler",
			},
		}
		rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, p)
		token, err := rawToken.SignedString(JWTKEY)
		if err != nil {
			serverError(&w, err)
			pipe2 <- false
			return
		}
		tokenChannel <- token
		pipe2 <- true
	}()
	//TODO this flag channel is causing deadlock if there is no 409 sent from other goroutines fix CreateUser func
	if c := <-flag; c == 409 {
		//checkUsername failed conflict exists
		fmt.Println("checkUsername failed conflict exists")
		w.WriteHeader(409)
	}

	stageComplete := <-pipe2
	if !stageComplete {
		fmt.Println("stage unsuccessful.")
		return
	}

	token := <-tokenChannel
	w.WriteHeader(200)
	http.SetCookie(w, &http.Cookie{
		Name:    "userToken",
		Value:   token,
		Expires: Tokenexpirytime,
	})

}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	rbytes, err := io.ReadAll(r.Body)
	if err != nil {
		serverError(&w, err)
	}
	var user models.Users
	json.Unmarshal(rbytes, &user)

	pipe1 := make(chan bool, 1)
	errorCode := make(chan uint64)
	token := make(chan string)

	fmt.Println("main function 1")
	go func() {

		fmt.Println("login gofunc")
		pipe2 := make(chan int, 1)

		var dbpassword string
		var exists bool
		var userid uint64
		fmt.Println("loginuser gofunc")
		go func() {
			dbpassword, exists, userid = models.LoginUser(user.Username)
			pipe2 <- 1
		}()
		<-pipe2

		if !exists {
			errorCode <- 404
			return
		}

		var isOK error

		fmt.Println("verifypass gofunc")
		pipe3 := make(chan int, 1)
		go func() {
			isOK = bcrypt.CompareHashAndPassword([]byte(dbpassword), []byte(user.Password))
			pipe3 <- 1
		}()
		<-pipe3

		if isOK != nil {
			//failure
			fmt.Println(isOK)
			errorCode <- 401
			return
		}
		payload := CustomPayload{
			userid,
			jwt.StandardClaims{
				ExpiresAt: Tokenexpirytime.Unix(),
				Issuer:    "loginuser handler",
			},
		}
		Token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
		t, err := Token.SignedString(JWTKEY)
		if err != nil {
			serverError(&w, err)
			pipe1 <- false
			return
		}
		token <- t
		pipe1 <- true
	}()
	fmt.Println("main function 2")
	if !<-pipe1 {
		fmt.Println("stage failed.")
		return
	}
	code := <-errorCode
	if code == 404 {
		w.WriteHeader(404)
		return
	}
	if code == 401 {
		w.WriteHeader(401)
		return
	}

	t := <-token
	http.SetCookie(w, &http.Cookie{
		Name:    "userToken",
		Value:   t,
		Expires: Tokenexpirytime,
	})
	w.WriteHeader(200)

}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	reqbyte, err := io.ReadAll(r.Body)
	if err != nil {
		serverError(&w, err)
		return
	}
	var username string
	json.Unmarshal(reqbyte, &username)

	var workers sync.WaitGroup
	channel1 := make(chan uint64, 1)
	workers.Add(1)
	go func() {
		defer workers.Done()
		token, err := jwt.ParseWithClaims(GetCookieByName(r.Cookies(), "userToken"), &CustomPayload{}, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		if claims, ok := token.Claims.(*CustomPayload); ok && token.Valid {
			channel1 <- claims.id
		}
	}()
	userid := <-channel1

	channel2 := make(chan error)
	workers.Add(1)
	go func() {
		defer workers.Done()
		channel2 <- models.DeleteUser(userid)

	}()
	err = <-channel2
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	workers.Wait()

	w.WriteHeader(200)

}

func SearchUsername(w http.ResponseWriter, r *http.Request) {
	rbyte, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Print(err)
	}
	var username string

	json.Unmarshal(rbyte, &username)
	found := make(chan bool, 1)
	go func() {
		res := models.FindUser(username)
		found <- res
	}()

	if <-found {
		w.WriteHeader(409)
		return
	}
	w.WriteHeader(200)

}

func UpdateUserPass(w http.ResponseWriter, r *http.Request) {
	rbyte, err := io.ReadAll(r.Body)
	if err != nil {
		serverError(&w, err)
		return
	}

	var newPass string
	err = json.Unmarshal(rbyte, &newPass)
	if err != nil {
		serverError(&w, err)
		return
	}

	var workers sync.WaitGroup

	//1
	channel1 := make(chan uint64, 1)
	errorChannel := make(chan error)
	workers.Add(1)
	go func() {
		defer workers.Done()
		tokenstr := GetCookieByName(r.Cookies(), "userToken")
		token, err := jwt.ParseWithClaims(tokenstr, &CustomPayload{}, nil)
		if err != nil {
			errorChannel <- err
			return
		}
		if payload, ok := token.Claims.(*CustomPayload); ok && token.Valid {
			channel1 <- payload.id
		}
	}()
	if <-errorChannel != nil {
		//fmt.Println(<-errorChannel)
		//w.WriteHeader(500)
		//return
		serverError(&w, <-errorChannel)
		return
	}
	userid := <-channel1

	//2

	channel2 := make(chan []byte)
	workers.Add(1)
	go func() {
		defer workers.Done()
		pass, err := bcrypt.GenerateFromPassword([]byte(newPass), bycryptCost)
		if err != nil {
			errorChannel <- err
			return
		}
		channel2 <- pass
	}()
	if <-errorChannel != nil {
		serverError(&w, <-errorChannel)
		return
	}
	hashpass := <-channel2

	//3

	workers.Add(1)
	go func() {
		defer workers.Done()
		errorChannel <- models.UpdatePass(string(hashpass), userid)
	}()
	if <-errorChannel != nil {
		serverError(&w, <-errorChannel)
		return
	}

	//wait
	workers.Wait()

	w.WriteHeader(200)

}

func GetUserById(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(400)
		return
	}

	var wg sync.WaitGroup

	//1

	channel1 := make(chan models.Users)
	wg.Add(1)
	go func() {
		defer wg.Done()
		channel1 <- models.GetUserDetails(username)
		//userDetails.Username = username
	}()
	userDetails := <-channel1
	fmt.Println(userDetails)

	//2
	channel2 := make(chan []models.Posts)
	wg.Add(1)
	go func() {
		defer wg.Done()
		channel2 <- models.GetPostsByUserId(userDetails.ID)
	}()
	posts := <-channel2

	wg.Wait()

	userPost := models.UserAndPost{User: userDetails, Posts: posts}
	parsedRes, err := json.Marshal(userPost)
	if err != nil {
		serverError(&w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(parsedRes)

}

// func parseReply(data any)
func serverError(w *http.ResponseWriter, err error) {
	if err != nil {
		fmt.Println(err)
	}
	(*w).WriteHeader(500)
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

/*
func GenerateSessionId() int64 {

}
*/
