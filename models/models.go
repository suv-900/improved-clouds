package models

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	//"gorm.io/plugin/prometheus"
	//	_ "github.com/lib/pq"
)

type Users struct {
	ID         uint64    `gorm:"primaryKey"`
	Username   string    `db:"username"`
	Email      string    `db:"email"`
	Password   string    `db:"password"`
	isLoggedIn bool      `db:"sessionValid"`
	CreatedAt  time.Time `db:"createdAt"`
	UpdatedAt  time.Time `db:"updatedAt"`
}

type Posts struct {
	ID            uint64    `gorm:"primaryKey"`
	Authorid      uint64    `db:"authorid"`
	Post_title    string    `db:"post_title"`
	Post_tldr     string    `db:"post_tldr"`
	Post_content  string    `db:"post_content"`
	Post_likes    uint      `db:"post_likes"`
	Post_category string    `db:"post_category"`
	CreatedAt     time.Time `db:"createdAt"`
	UpdatedAt     time.Time `db:"updatedAt"`
}

type Category struct {
	CategoryId    uint64 `gorm:"primarykey"`
	Category_Name string `db:"category_name"`
	Post_count    uint   `db:"post_count"`
}

type Comment struct {
	Comment_id      uint64    `gorm:"primarykey"`
	Post_id         uint64    `db:"post_id"`
	User_id         uint64    `db:"user_id"`
	Username        string    `db:"username"`
	Comment_content string    `db:"comment_content"`
	Comment_likes   uint64    `db:"comment_likes"`
	CreatedAt       time.Time `db:"createdAt"`
	UpdatedAt       time.Time `db:"updatedAt"`
}

type UserAndPost struct {
	User  Users
	Posts []Posts
}
type UsernameAndPost struct {
	Username string
	Post     Posts
}

type UsernameAndComment struct {
	User_id         uint64
	Username        string
	CommentID       uint64
	Comment_content string
	Comment_likes   uint64
}
type Passanduserid struct {
	Password string `db:"password"`
	User_id  uint64 `db:"user_id"`
}

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

/*
	db.Use(prometheus.New(prometheus.Config{
			DBName:          "cloud",
			RefreshInterval: 10,
			PushAddr:        "prometheus pusher address",
			StartServer:     true,
			HTTPServerPort:  8080,
			MetricsCollector: []prometheus.MetricsCollector{
				&prometheus.Postgres{
					VariableNames: []string{"Thread running"},
				},
			},
		}))
*/
//DONE
func FindUser(username string) bool {
	var res int32
	db.Raw("SELECT COUNT(username) FROM users WHERE username=?", username).Scan(&res)
	return res != 0
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
			sql := "INSERT INTO users (username,email,password,isLoggedIn,createdat,updatedat) VALUES(?,?,?,?,?) RETURNING user_id"
			res := tx.Raw(sql, user.Username, user.Email, user.Password, true, user.CreatedAt, user.UpdatedAt).Scan(&userid)
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
	db.Raw("SELECT EXISTS (SELECT 1 FROM users WHERE username=?)", username).Scan(&check)
	if check {
		p := make(chan int, 1)
		go func() {
			db.Raw("SELECT user_id,password FROM users WHERE username= ?", username).Scan(&res)
			p <- 1
		}()
		<-p
		go func() {
			db.Raw("UPDATE users(active) VALUES(?) WHERE user_id=?", true, res.User_id)
			p <- 1
		}()
		<-p
		return res.Password, true, res.User_id
	}
	return "", false, 0
}

// DONE
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
	p := make(chan bool, 1)
	go func() {
		tx := db.Begin()
		r := tx.Raw("UPDATE users(active) VALUES(?) WHERE user_id=?", "", false, userid)
		if r.Error != nil {
			fmt.Println("error while logging out user")
			p <- false
		}
		p <- true
	}()
	s := <-p
	return s
}

// TODO
func GetUserDetails(username string) Users {
	var user Users
	db.Raw("SELECT (user_id,username,userabout) FROM users WHERE username=?", username).Scan(&user)
	return user
}

// DONE
func DeleteUser(userid uint64) error {
	//delete user and posts
	r := db.Raw("DELETE FROM users WHERE userid=? ", userid)
	if r.Error != nil {
		fmt.Println(r.Error)
		return r.Error
	}
	r = db.Raw("DELETE FROM posts WHERE authorid=?", userid)
	if r.Error != nil {
		fmt.Println(r.Error)
		return r.Error
	}
	return nil
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

// POST
func CreatePost(post Posts) (uint64, error) {
	var postid uint64
	pipe := make(chan bool, 1)
	var err error

	go func() {
		tx := db.Begin()
		result := tx.Raw("INSERT INTO posts (post_title,post_content,author_id,post_likes,post_category) VALUES(?,?,?,?,?,?) RETURNING post_id", post.Post_title, post.Post_content, post.Authorid, 0, post.Post_category).Scan(&postid)
		if result.Error != nil {
			tx.Rollback()
			err = result.Error
			postid = 0
			pipe <- false
			return
		} else {
			tx.Commit()
			err = nil
			pipe <- true
			return
		}
	}()
	<-pipe
	return postid, err
}
func DeletePost(postid uint64) error {

	r := db.Exec("DELETE FROM posts WHERE postid=?", postid)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

func UpdatePost(postid uint64, post Posts) error {

	tx := db.Begin()
	r := tx.Exec("UPDATE posts SET post_content=? post_title=? WHERE postid=?", post.Post_content, post.Post_title, postid)
	if r.Error != nil {
		tx.Rollback()
		return r.Error
	} else {
		tx.Commit()
		return nil
	}

}
func GetPostsByUserId(userId uint64) []Posts {
	var posts []Posts
	db.Raw("SELECT (post_title,post_content) FROM posts WHERE authorid=? LIMIT 5", userId).Scan(&posts)
	return posts
}
func PostById(postid uint64) (Posts, string) {
	var post Posts
	var username string
	db.Raw("SELECT (post_title,post_content,post_category,post_tldr,authorid) FROM posts WHERE post_id=?", postid).Scan(&post)
	db.Raw("SELECT username FROM users WHERE user_id=?", post.Authorid).Scan(&username)
	return post, username
}

func FeedGenerator(userid uint64) []Posts {

	var categories []string

	db.Raw(`SELECT c.category_name
  FROM post_likes pl JOIN posts p ON pl.post_id=p.post_id
  JOIN categories c ON c.category_id=p.category_id
  WHERE pl.user_id=?
  GROUP BY c.category_name
  LIMIT 5`, userid).Scan(&categories)

	var posts []Posts
	db.Raw(` 
  SELECT p.author_id,p.post_title,p.post_content,COUNT(pl.like_id) as likes_num
  FROM post_likes pl JOIN post p ON pl.post_id=p.post_id
  JOIN categories c ON c.category_id=p.category_id
  WHERE c.category_name=?
  GROUP BY p.post_id,p.author_id,p.post_title,p.post_content
  ORDER BY likes_num DESC
  LIMIT 10
  `, categories).Scan(&posts)
	return posts
}

// COMMENT

func LikePost(postid uint64, userid uint64) bool {
	var worker sync.WaitGroup

	worker.Add(1)
	errchannel := make(chan bool, 1)
	go func() {
		defer worker.Done()
		tx := db.Begin()
		r := tx.Exec("INSERT INTO like_posts (user_id,post_id) VALUES(?,?)", userid, postid)
		if r.Error != nil {
			tx.Rollback()
			fmt.Println("Error occured while inserting like to like_posts %w", r.Error)
			errchannel <- true
			return
		}
		errchannel <- false
	}()

	if <-errchannel {
		return true
	}

	worker.Add(1)
	go func() {
		defer worker.Done()
		db.Exec("UPDATE posts SET num_likes=num_likes+1 WHERE post_id=?", postid)
	}()

	worker.Wait()
	return false
}

func LikeAComment(commentid uint64) {
	db.Exec("UPDATE comments SET comment_likes=comment_likes+1 WHERE comment_id=?", commentid)
}

func DislikeAComment(commentid uint64) {
	db.Exec("UPDATE comments SET comment_likes=comment_likes-1 WHERE comment_id=?", commentid)
}

func GetCommentsByPostID(postid uint64) []UsernameAndComment {
	commentarr := make([]UsernameAndComment, 5)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		db.Raw(
			` SELECT (comment_id,user_id,username,comment_content) FROM post_comments WHERE post_id=? 
           GROUP BY comment_likes DESC 
           LIMIT 5 OFFSET
        `, postid).Scan(&commentarr)
	}()
	/*
	   		rawComment := UsernameAndComment{
	   			UserID:          comment.User_id,
	   			Username:        username,
	         CommentID:       comment.Comment_id,
	   			Comment_content: comment.Comment_content,
	   		}
	   		commentarr = append(commentarr, rawComment)
	   	}
	*/
	fmt.Println(commentarr)
	return commentarr
}

func AddComment(postid uint64, userid uint64, comment string) error {

	tx := db.Begin()
	r := tx.Exec("INSERT INTO post_comments (user_id,post_id,comment_content) VALUES(?,?,?)", userid, postid, comment)
	if r.Error != nil {
		tx.Rollback()
		return r.Error
	}
	return nil
	//var commentid uint
	//db.Raw("SELECT comment_id FROM comments WHERE post_id=? user_id=? comment_content=?", postid, userid, comment).Scan(&commentid)
	//return commentid, nil
}

func EditComment(commentId uint64, comment string) {
	db.Exec("UPDATE comments SET comment_content=? WHERE comment_id=?", commentId, comment)
}

func RemoveComment(commentId uint64) {
	db.Exec("DELETE * FROM comments WHERE comment_id=?", commentId)
}

// image
func AddProfilePic(userid uint64) {
}
