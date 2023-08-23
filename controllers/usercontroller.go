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

func PollMe(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "help")
}

func serverError(w *http.ResponseWriter, err error) {
	fmt.Println(err)
	(*w).WriteHeader(500)
}

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

	var workergroup sync.WaitGroup

	workergroup.Add(1)
	go func() {
		defer workergroup.Done()
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bycryptCost)
		if err != nil {
			serverError(&w, err)
			return
		}
		user.Password = string(hashedPassword)
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		fmt.Println(user)
	}()

	channel1 := make(chan uint64, 1)
	flag := make(chan uint, 1)

	workergroup.Add(1)
	go func() {
		defer workergroup.Done()
		if !models.FindUser(user.Username) {
			id, err := models.CreateUser(user)
			if err != nil {
				serverError(&w, err)
				return
			}
			channel1 <- id
		}
		flag <- 409
	}()
	if c := <-flag; c == 409 {
		//checkUsername failed conflict exists
		fmt.Println("checkUsername failed conflict exists")
		w.WriteHeader(409)
		return
	}
	userid := <-channel1

	channel2 := make(chan string, 1)
	err1 := make(chan bool, 1)
	workergroup.Add(1)
	go func() {
		defer workergroup.Done()
		p := CustomPayload{
			userid,
			jwt.StandardClaims{
				ExpiresAt: Tokenexpirytime.Unix(),
				Issuer:    "createUser handler",
			},
		}
		rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, p)
		token, err := rawToken.SignedString(JWTKEY)
		if err != nil {
			fmt.Println(err)
			err1 <- true
			return
		}
		channel2 <- token
	}()
	if <-err1 {
		serverError(&w, nil)
		return
	}

	token := <-channel2

	workergroup.Wait()

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

	var workers sync.WaitGroup

	channel1 := make(chan int32, 1)
	channel2 := make(chan uint64)
	workers.Add(1)
	go func() {
		defer workers.Done()
		dbpassword, exists, userid := models.LoginUser(user.Username)
		if !exists {
			channel1 <- 404
			return
		}
		isOK := bcrypt.CompareHashAndPassword([]byte(dbpassword), []byte(user.Password))
		if isOK != nil {
			//failure
			channel1 <- 401
			return
		}
		channel2 <- userid
	}()

	errorCode := <-channel1
	if errorCode == 404 {
		w.WriteHeader(404)
		return
	}
	if errorCode == 401 {
		w.WriteHeader(401)
		return
	}
	userid := <-channel2

	channel3 := make(chan string, 1)
	go func() {
		defer workers.Done()
		payload := CustomPayload{
			userid,
			jwt.StandardClaims{
				ExpiresAt: Tokenexpirytime.Unix(),
				Issuer:    "loginuser handler",
			},
		}
		Token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
		token, err := Token.SignedString(JWTKEY)
		if err != nil {
			fmt.Println(err)
			return
		}
		channel3 <- token
	}()

	workers.Wait()

	http.SetCookie(w, &http.Cookie{
		Name:    "userToken",
		Value:   <-channel3,
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
		channel2 <- models.DeleteUser(username, userid)
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

	channel1 := make(chan bool)
	go func() {
		channel1 <- models.FindUser(username)
	}()

	isFound := <-channel1
	if isFound {
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
		channel2 <- models.GetAllPostsByUserId(userDetails.ID)
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

//func parseReply(data any)

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
