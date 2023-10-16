package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/suv-900/blog/models"
)

func AddComment(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup

	//1 verify token
	channel1 := make(chan uint64, 2)
	errorChannel := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ok, p := TokenVerifier("userToken", r)
		if !ok {
			errorChannel <- true
			return
		}
		channel1 <- p.ID

		ok, p = TokenVerifier("postToken", r)
		if !ok {
			errorChannel <- true
			return
		}
		channel1 <- p.ID
	}()

	if <-errorChannel {
		w.WriteHeader(401)
		return
	}

	userid := <-channel1
	postid := <-channel1

	rbody, err := io.ReadAll(r.Body)
	if err != nil {
		serverError(&w, err)
		return
	}
	var comment string
	json.Unmarshal(rbody, &comment)

	//2 add comment
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := models.AddComment(postid, userid, comment)
		if err != nil {
			fmt.Println(err)
			errorChannel <- true
			return
		}
	}()
	if <-errorChannel {
		serverError(&w, nil)
		return
	}
	//not sending back comment string use in frontend
	wg.Wait()

	w.WriteHeader(200)
}

func FetchComments(w http.ResponseWriter, r *http.Request) {
	//session token
	rbody, err := io.ReadAll(r.Body)
	if err != nil {
		serverError(&w, err)
		return
	}
	var offset uint16
	json.Unmarshal(rbody, offset)

	var wg sync.WaitGroup

	//1
	channel1 := make(chan uint64)
	errChannel := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ok, p := TokenVerifier("postToken", r)
		if !ok {
			fmt.Println("Error while parsing token!")
			errChannel <- true
			return
		}
		channel1 <- p.ID
		errChannel <- false
	}()

	if <-errChannel {
		w.WriteHeader(401)
		return
	}
	postid := <-channel1

	//2
	channel2 := make(chan []byte)
	err2 := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		comments := models.GetCommentsByPostID(postid)
		parsedRes, err := json.Marshal(comments)
		if err != nil {
			fmt.Println(err)
			err2 <- true
			return
		}
		channel2 <- parsedRes
		err2 <- false
	}()
	if <-err2 {
		w.WriteHeader(500)
		return
	}

	wg.Wait()

	w.WriteHeader(200)
	w.Write(<-channel2)
}

func EditComment(w http.ResponseWriter, r *http.Request) {
	var commentId uint64
	ok, _ := TokenVerifier("userToken", r)
	if ok {
		commentId, ok = ParseToken(GetCookieByName(r.Cookies(), "commentToken"))
		if !ok {
			w.WriteHeader(500)
			fmt.Println("Token not ok OR not Valid")
			return
		}

	} else {
		w.WriteHeader(401)
		return
	}

	rbody, err := io.ReadAll(r.Body)
	if err != nil {
		serverError(&w, err)
		return
	}

	var comment string
	json.Unmarshal(rbody, &comment)

	models.EditComment(commentId, comment)

	w.WriteHeader(200)

}

func RemoveComment(w http.ResponseWriter, r *http.Request) {
	var commentId uint64
	ok, _ := TokenVerifier("userToken", r)
	if ok {
		commentId, ok = ParseToken(GetCookieByName(r.Cookies(), "commentToken"))
		if !ok {
			w.WriteHeader(500)
			fmt.Println("Token not ok OR not Valid")
			return
		}

	} else {
		w.WriteHeader(401)
		return
	}

	models.RemoveComment(commentId)

	w.WriteHeader(200)

}

func ParseToken(token string) (uint64, bool) {
	t, err := jwt.ParseWithClaims(token, &CustomPayload{}, nil)
	if err != nil {
		fmt.Println("Error while parsing token.", err)
	}
	if p, ok := t.Claims.(*CustomPayload); ok && t.Valid {
		return p.ID, true
	}
	return 0, false
}
