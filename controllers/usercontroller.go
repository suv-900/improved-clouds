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

// TODO token blacklist
var sessionKey = "dasd0871hudsliuqnbvc872832madsa1207badp1831ajlq32103avkwqe871181"
var stokenlength = 10
var bycryptCost = 3
var JWTKEY = []byte(os.Getenv("JWT_KEY"))
var Tokenexpirytime = time.Now().Add(60 * time.Minute)

type CustomPayload struct {
	ID uint64 `json:"id"`
	jwt.StandardClaims
}

func CheckServerHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

}

// completed
// TODO unit tests
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var err error

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

	var userFound bool
	c := make(chan int, 1)
	go func() {
		userFound = models.FindUser(user.Username)
		c <- 1
	}()
	<-c

	if userFound {
		w.WriteHeader(409)
		return
	}

	a := make(chan int, 1)
	go func() {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bycryptCost)
		if err != nil {
			a <- 1
			return
		}

		user.Password = string(hashedPassword)
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		a <- 1
	}()
	<-a
	if err != nil {
		serverError(&w, err)
		return
	}

	b := make(chan int, 1)
	var userid uint64
	go func() {
		//defer userCreatewg.Done()
		userid, err = models.CreateUser(user)
		b <- 1
	}()
	<-b
	if err != nil {
		serverError(&w, err)
		return
	}

	d := make(chan int, 1)
	var token string
	go func() {

		//TODO add user with same name
		//TODO i dont think this call gets awaited/the goroutine waits for this call

		//var err error

		payload := CustomPayload{
			ID: userid,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: Tokenexpirytime.Unix(),
				Issuer:    "createUser handler",
			},
		}
		rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
		token, err = rawToken.SignedString(JWTKEY)
		d <- 1
	}()
	<-d
	if err != nil {
		serverError(&w, err)
		return
	}

	t, err := json.Marshal(token)
	if err != nil {
		serverError(&w, err)
		return
	}
	w.WriteHeader(200)
	w.Write(t)
}

// completed
func LoginUser(w http.ResponseWriter, r *http.Request) {
	rbytes, err := io.ReadAll(r.Body)
	if err != nil {
		serverError(&w, err)
	}
	var user models.Users
	json.Unmarshal(rbytes, &user)

	var dbpassword string
	var exists bool
	var id uint64
	a := make(chan int, 1)
	go func() {
		dbpassword, exists, id = models.LoginUser(user.Username)
		a <- 1
	}()
	<-a
	if err != nil {
		serverError(&w, err)
		return
	}
	if !exists {
		w.WriteHeader(404)
		return
	}

	var passValid error
	b := make(chan int, 1)
	go func() {
		passValid = bcrypt.CompareHashAndPassword([]byte(dbpassword), []byte(user.Password))
		b <- 1
	}()
	<-b
	if passValid != nil {
		w.WriteHeader(401)
		return
	}
	payload := CustomPayload{
		ID: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: Tokenexpirytime.Unix(),
			Issuer:    "loginHandler",
		},
	}
	Token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := Token.SignedString(JWTKEY)
	if err != nil {
		serverError(&w, err)
		return
	}
	ts, err := json.Marshal(t)
	if err != nil {
		serverError(&w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(ts)
	/*
		var stoken string
		pipe4 := make(chan bool, 1)
		go func() {
			s := make([]byte, stokenlength)
			for i := range s {
				s[i] = sessionKey[rand.Intn(len(sessionKey))]
			}
			ok := models.AddSessionToken(string(s), userid)
			if !ok {
				pipe4 <- false
				fmt.Println("adding session token failed.")
				return
			}
			stoken = string(s)
			pipe4 <- true
		}()
		if !<-pipe4 {
			fmt.Println("adding sessionToken stage failed.")
			serverError(&w, nil)
			pipe1 <- false
			errorCode <- 0
			return
		}*/
	//stokenchan <- stoken

	/*
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   t,
			Expires: Tokenexpirytime,
		})
	*/

}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
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
	/*
		var userExists bool
		c := make(chan int, 1)
		go func() {
			userExists = models.CheckUserExists(userid)
			c <- 1
		}()
		<-c
		if !userExists {
			w.WriteHeader(400)
			return
		}
	*/
	var err error
	err = nil
	b := make(chan int, 1)
	go func() {

		err = models.DeleteUser(userid)
		if err != nil {
			b <- 1
			return
		}
		//err = models.DeleteUser(userid)
		/*
			if err != nil {
				b <- 1
				return
			}*/
		b <- 1
	}()
	<-b
	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err)
		return
	}
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
			channel1 <- payload.ID
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

/*
func GenerateSessionId() int64 {

}
*/

func CreateToken(w http.ResponseWriter, r *http.Request) {
	p := CustomPayload{
		1,
		jwt.StandardClaims{
			ExpiresAt: Tokenexpirytime.Unix(),
			Issuer:    "createUser handler",
		},
	}
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, p)
	token, err := rawToken.SignedString(JWTKEY)
	if err != nil {
		serverError(&w, err)
		return
	}
	res, _ := json.Marshal(token)
	w.Write(res)
}
