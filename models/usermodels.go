package models

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	//"gorm.io/plugin/prometheus"
	//	_ "github.com/lib/pq"
)

//root:Core@123@/blogweb?
// postgres://core:12345678@localhost:5432/cloud

var db *gorm.DB

func ConnectDB() error {
	dsn := "host=localhost user=core password=12345678 dbname=cloud"
	dbget, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	db = dbget
	return nil

}

// DONE
func FindUser(username string) bool {
	var res int32
	db.Raw("SELECT COUNT(username) FROM users WHERE username=?", username).Scan(&res)
	return res != 0
}

func GetUsername(userid uint64) string {
	var u string
	db.Raw("SELECT username FROM users WHERE user_id=?", userid).Scan(&u)
	return u
}

// DONE
func CreateUser(user Users) (uint64, error) {
	var userid uint64
	var err error
	pipe1 := make(chan bool, 1)
	go func() {
		pipe2 := make(chan bool, 1)
		pipe3 := make(chan int, 1)

		go func() {
			tx := db.Begin()
			sql := "INSERT INTO users (username,email,password,createdat,updatedat) VALUES(?,?,?,?,?) RETURNING user_id"
			res := tx.Raw(sql, user.Username, user.Email, user.Password, user.CreatedAt, user.UpdatedAt).Scan(&userid)
			if res.Error != nil {
				tx.Rollback()
				pipe2 <- false
				err = res.Error
				return
			}
			tx.Commit()
			err = nil
			pipe2 <- true
		}()

		if !<-pipe2 {
			return
		}

		go func() {
			db.Raw("UPDATE active FROM users WHERE user_id=?", true)
			pipe3 <- 1
		}()
		<-pipe3

		pipe1 <- true
	}()
	if !<-pipe1 {
		return 0, err
	}
	return userid, nil
}

/*
	func AddSessionToken(s string, userid uint64) bool {
		tx := db.Begin()

		pipe1 := make(chan bool, 1)
		go func() {
			r := tx.Raw("INSERT INTO users(sessionToken,sessionValid) VALUES(?,?) WHERE user_id=?", s, true, userid)
			if r.Error != nil {
				fmt.Println(r.Error)
				pipe1 <- false
				return
			}
			pipe1 <- true
		}()

		p := <-pipe1
		return p

}
*/
//DONE
func LoginUser(username string) (string, bool, uint64) {
	var check bool
	var res Passanduserid

	a := make(chan int, 1)

	go func() {
		db.Raw("SELECT EXISTS (SELECT 1 FROM users WHERE username=?)", username).Scan(&check)
		a <- 1
	}()
	<-a
	if check {
		p := make(chan int, 1)
		go func() {
			db.Raw("SELECT user_id,password FROM users WHERE username= ?", username).Scan(&res)
			p <- 1
		}()
		<-p
		q := make(chan int, 1)
		go func() {
			r := db.Exec("UPDATE users SET active= ? WHERE user_id= ?", true, res.User_id)
			fmt.Println(r.Error)
			q <- 1
		}()
		<-q
		return res.Password, true, res.User_id
	}
	return "", false, 0
}

// checks if user is logged in
func CheckUserLoggedIn(userid uint64) bool {
	p := make(chan bool, 1)
	go func() {
		var active bool
		db.Raw("SELECT active FROM users WHERE user_id=?", userid).Scan(&active)
		p <- active
	}()
	res := <-p
	return res
}

// DONE
func LogOut(userid uint64) bool {
	p := make(chan int, 1)
	var ok bool
	go func() {
		tx := db.Begin()
		r := tx.Exec("UPDATE users SET active=? WHERE user_id=?", false, userid)
		if r.Error != nil {
			fmt.Println("error while logging out user")
			p <- 1
			ok = false
		}
		ok = true
		p <- 1
	}()
	<-p
	return ok
}

// TODO
func GetUserDetails(username string) Users {
	var user Users
	db.Raw("SELECT (user_id,username,userabout) FROM users WHERE username=?", username).Scan(&user)
	return user
}

// DONE
func DeleteUser(userid uint64) error {
	var err error
	err = nil
	//comments can be deleted or no
	a := make(chan int, 1)
	go func() {
		r := db.Exec("DELETE FROM comments WHERE user_id=?", userid)
		if r.Error != nil {
			fmt.Println(r.Error)
			err = r.Error
			a <- 1
			return
		}
		a <- 1
	}()
	<-a
	if err != nil {
		return err
	}

	b := make(chan int, 1)
	go func() {
		r := db.Exec("DELETE FROM posts WHERE author_id=?", userid)
		if r.Error != nil {
			fmt.Println(r.Error)
			err = r.Error
			b <- 1
			return
		}
		b <- 1
	}()
	<-b
	if err != nil {
		return err
	}

	c := make(chan int, 1)
	go func() {
		r := db.Exec("DELETE FROM users WHERE user_id=?", userid)
		if r.Error != nil {
			fmt.Println(r.Error)
			err = r.Error
			c <- 1
			return
		}
		c <- 1
	}()
	<-c
	if err != nil {
		return err
	}

	return err
}

func DeleteUser2(userid uint64) {
	a := make(chan int, 1)
	go func() {
		db.Exec("DELETE FROM comments WHERE user_id=?", userid)
		a <- 1
	}()
	<-a
	b := make(chan int, 1)
	go func() {
		db.Exec("DELETE FROM posts WHERE author_id=?", userid)
		b <- 1
	}()
	<-b
	c := make(chan int, 1)
	go func() {
		db.Exec("DELETE FROM users WHERE user_id=?", userid)
		c <- 1
	}()
	<-c
}

func UpdatePass(pass string, userid uint64) error {
	tx := db.Begin()
	r := tx.Exec("UPDATE users SET password=? WHERE userid=?", pass, userid)
	if r.Error != nil {
		tx.Rollback()
		fmt.Println(r.Error)
		return r.Error
	} else {
		tx.Commit()
		return nil
	}
}

/*
func CheckCategory(categoryName string) (bool, uint) {
	var categoryId uint
	db.Exec("SELECT (category_id) FROM category WHERE category_name=?", categoryName).Scan(&category)
	if categoryId == nil {
		return false, 0
	}
	return true, categoryId
}*/

// creates a post
func CheckUserExists(userid uint64) bool {
	a := make(chan int, 1)
	var exists bool
	go func() {
		db.Raw("SELECT EXISTS (SELECT 1 FROM users WHERE user_id=?)", userid).Scan(&exists)
		a <- 1
	}()
	<-a
	return exists
}

// image
func AddProfilePic(userid uint64) {
}
