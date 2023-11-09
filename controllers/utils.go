package controllers

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"
)

func AuthenticateTokenAndSendUserID(r *http.Request) (bool, uint64, bool) {
	//tokenExpired ID tokenInvalid
	var token string
	var userid uint64
	token = r.Header.Get("Authorization")
	//token = GetCookieByName(r.Cookies(), "Authorization")
	t, err := jwt.ParseWithClaims(token, &CustomPayload{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTKEY), nil
	})
	if err != nil {
		fmt.Println(err)
		return true, 0, false
	}

	if p, ok := t.Claims.(*CustomPayload); ok && t.Valid {
		userid = p.ID
		/*
			pipe1 := make(chan bool, 1)
			go func() {

				active := models.CheckUserLoggedIn(userid)
				if !active {
					pipe1 <- false
					return
				}
				pipe1 <- true
			}()
			if !<-pipe1 {
				fmt.Println("User is !active")
				(*w).WriteHeader(401)
				return false, 0
			}*/
		return false, userid, false
	}
	return false, 0, true
}
