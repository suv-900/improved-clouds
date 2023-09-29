package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	//"strconv"
	"sync"

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
	//	cookie, _ := r.Cookie("usertoken")
	var worker sync.WaitGroup

	var authorid uint64

	//1
	worker.Add(1)
	signal := make(chan bool, 1)
	go func() {
		defer worker.Done()
		ok, p := TokenVerifier("userToken", r)
		if ok {
			authorid = p.id
		}
		signal <- true
		fmt.Printf("Error while parsing token.")

	}()
	if <-signal {
		w.WriteHeader(500)
		return
	}
	//2
	worker.Add(1)
	channel1 := make(chan []byte)
	signal1 := make(chan bool, 1)

	go func() {
		defer worker.Done()
		rbyte, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Error while reading body. %w", err)
			signal1 <- true
			return
		}
		channel1 <- rbyte
	}()
	if <-signal {
		w.WriteHeader(500)
		return
	}

	rbyte := <-channel1

	var post models.Posts
	err := json.Unmarshal(rbyte, &post)
	if err != nil {
		fmt.Printf("Error while unmarshalling. %w", err)
		w.WriteHeader(500)
		return
	}
	post.Authorid = authorid

	worker.Add(1)
	channel3 := make(chan uint64)
	go func() {
		defer worker.Done()
		p, err := models.CreatePost(post)
		if err != nil {
			fmt.Printf("Error while unmarshalling. %w", err)
			signal <- true
			return
		}
		channel3 <- p
	}()
	if <-signal {
		w.WriteHeader(500)
		return
	}
	postid := <-channel3

	worker.Add(1)
	channel4 := make(chan string, 1)
	go func() {
		defer worker.Done()
		payload := CustomPayload{
			postid,
			jwt.StandardClaims{
				ExpiresAt: Tokenexpirytime.Unix(),
				Issuer:    "createpost H",
			},
		}
		rawtoken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
		ptoken, err := rawtoken.SignedString(JWTKEY)
		if err != nil {
			fmt.Printf("Error while creating token. %w", err)
			signal <- true
			return
		}
		channel4 <- ptoken
	}()
	if <-signal {
		w.WriteHeader(500)
		return
	}

	posttoken := <-channel4

	w.WriteHeader(200)
	http.SetCookie(w, &http.Cookie{
		Name:    "postToken",
		Value:   posttoken,
		Expires: Tokenexpirytime,
	})

	// var userPosts []models.Posts
	// userPosts = models.CreatePost(post.Username, post)
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	var postid uint64

	ok, payload := TokenVerifier("postToken", r)
	if ok {
		postid = payload.id
	} else {
		w.WriteHeader(500)
		return
	}

	err := models.DeletePost(postid)
	if err != nil {
		serverError(&w, err)
		return
	}
	w.WriteHeader(200)
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	var postid uint64

	ok, payload := TokenVerifier("postToken", r)
	if ok {
		postid = payload.id
	} else {
		w.WriteHeader(500)
		return
	}

	rbyte, err := io.ReadAll(r.Body)
	if err != nil {
		serverError(&w, err)
		return
	}

	var post models.Posts
	err = json.Unmarshal(rbyte, &post)
	if err != nil {
		serverError(&w, err)
		return
	}

	err = models.UpdatePost(postid, post)
	if err != nil {
		serverError(&w, err)
		return
	}

	w.WriteHeader(200)
}

func PostViewer(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup

	rbody, err := io.ReadAll(r.Body)
	if err != nil {
		serverError(&w, err)
		return
	}

	var postid uint64
	err = json.Unmarshal(rbody, &postid)

	// p := r.URL.Query().Get("postid")
	// if p == "" {
	// 	w.WriteHeader(400)
	// 	return
	// }
	// postid, err := strconv.ParseUint(p, 10, 64)
	// if err != nil {
	// 	serverError(&w, err)
	// 	return
	// }

	wg.Add(1)
	channel1 := make(chan string, 1)
	channel2 := make(chan models.Posts)
	go func() {
		defer wg.Done()
		p, u := models.PostById(postid)
		channel1 <- u
		channel2 <- p
	}()

	post := <-channel2
	username := <-channel1

	channel3 := make(chan []byte)
	errchannel := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		parsedRes, err := json.Marshal(models.UsernameAndPost{Username: username, Post: post})
		if err != nil {
			fmt.Println(err)
			errchannel <- true
		}
		channel3 <- parsedRes
	}()
	if <-errchannel {
		w.WriteHeader(500)
		return
	}
	parsedRes := <-channel3

	wg.Wait()

	w.WriteHeader(200)
	w.Write(parsedRes)
	//1
	// wg.Add(1)
	// err1:=make(chan bool,1)
	// go func(){
	// 	defer wg.Done()
	// 	ok,_:=TokenVerifier("userToken",r)
	// 	if !ok{
	// 		err1<-true
	// 		return
	// 	}
	// }()
	// if <-err1{
	// 	w.WriteHeader(401)
	// 	return
	// }

}

func TokenVerifier(s string, r *http.Request) (bool, *CustomPayload) {
	token, err := jwt.ParseWithClaims(GetCookieByName(r.Cookies(), s), &CustomPayload{}, nil)
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

func SendTokenLocal(w http.ResponseWriter, r *http.Request) {

	payload := CustomPayload{
		23,
		jwt.StandardClaims{
			ExpiresAt: Tokenexpirytime.Unix(),
			Issuer:    "sex handler",
		},
	}
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	token, err := rawToken.SignedString(JWTKEY)
	if err != nil {
		serverError(&w, err)
		return
	}

	fmt.Println(token)
	parsedToken, err := json.Marshal(token)
	if err != nil {
		serverError(&w, err)
		return
	}
	w.WriteHeader(200)
	w.Write(parsedToken)
}

func EchoSocket(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	parseVal, err := json.Marshal(token)
	if err != nil {
		serverError(&w, err)
		return
	}
	w.WriteHeader(200)
	w.Write(parseVal)
}

func createError() bool {
	return true
}
func cleanup() {
	fmt.Println("cleanup code.")
}
func CheckConcurr(w http.ResponseWriter, r *http.Request) {
	defer cleanup()
	var wg sync.WaitGroup
	wg.Add(1)
	errchannel := make(chan bool, 1)
	go func() {
		defer wg.Done()
		defer cleanup()
		fmt.Println("inside thread")
		if createError() {
			errchannel <- true
			serverError(&w, nil)
			return
		}
	}()
	wg.Wait()
	if <-errchannel {
		fmt.Println("returning thread err.")
		return
	}
	fmt.Println("inside the function")
}
