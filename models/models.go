package models

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	//	_ "github.com/go-sql-driver/mysql"
	//	"github.com/jmoiron/sqlx"
	//	_ "github.com/lib/pq"
)

 

type Users struct {
	ID        uint64    `gorm:"primaryKey"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	Password  string    `db:"passwordhash"`
	CreatedAt time.Time `db:"createdAt"`
	UpdatedAt time.Time `db:"updatedAt"`
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
	Comment_content string    `db:"comment_content"`
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
	UserID          uint64
	Username        string
	Comment_content string
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

func CreateUser(user Users) (uint64, error) {

	var userid uint64
	tx := db.Begin()
	result := tx.Exec("INSERT INTO users (username,email,password,createdAt,updatedAt) VALUES(?,?,?,?,?)", user.Username, user.Email, user.Password, user.CreatedAt, user.UpdatedAt).Scan(&userid)

	if result.Error != nil {
		fmt.Println(result.Error)
		return 0, result.Error
	} else {
		tx.Commit()
		return userid, nil
	}

	// db.Exec("SELECT userid FROM users WHERE username=?",user.Username).Scan()
}

func FindUser(username string) bool {
	var check bool
	db.Raw("SELECT EXISTS ( SELECT 1 FROM users WHERE username=?  )", username).Scan(&check)
  return check
}

func GetUserDetails(username string) Users {
	var user Users
	db.Raw("SELECT (user_id,username,userabout) FROM users WHERE username=?", username).Scan(&user)
	return user
}

func LoginUser(username string) (string, bool, uint64) {
	var dbpassword string
	var userid uint64
  var check bool
	db.Raw("SELECT EXISTS (SELECT 1 FROM users WHERE username=?)",username).Scan(&check)
  if check{
    db.Raw("SELECT (password,userid) FROM users WHERE username=?", username).Scan(&dbpassword).Scan(&userid)
    return dbpassword,true,userid
  }	
	return "",false,0
}

func DeleteUser(userid uint64) error {
	//delete user and posts

	r := db.Exec("DELETE FROM users WHERE userid=? ",userid)
	if r.Error != nil {
		fmt.Println(r.Error)
		return r.Error
	}
  r=db.Exec("DELETE FROM posts WHERE authorid=?", userid)
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
	tx := db.Begin()
	result := tx.Exec("INSERT INTO posts (post_title,post_tldr,post_content,authorid,post_likes,post_category) VALUES(?,?,?,?,?,?)", post.Post_title, post.Post_tldr, post.Post_content, post.Authorid, 0, post.Post_category).Scan(&postid)
	if result.Error != nil {
		tx.Rollback()
		return 0, result.Error
	} else {
		tx.Commit()
		return postid, nil
	}

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

func FeedGenerator(userid uint64)[]Posts{

  var categories []string

  db.Raw(`SELECT c.category_name
  FROM post_likes pl JOIN posts p ON pl.post_id=p.post_id
  JOIN categories c ON c.category_id=p.category_id
  WHERE pl.user_id=?
  GROUP BY c.category_name
  LIMIT 5`,userid).Scan(&categories)

  var posts []Posts
  db.Raw(` 
  SELECT p.author_id,p.post_title,p.post_content,COUNT(pl.like_id) as likes_num
  FROM post_likes pl JOIN post p ON pl.post_id=p.post_id
  JOIN categories c ON c.category_id=p.category_id
  WHERE c.category_name=?
  GROUP BY p.post_id,p.author_id,p.post_title,p.post_content
  ORDER BY likes_num DESC
  LIMIT 10
  ` ,categories).Scan(&posts) 
  return posts
}


// COMMENT

func LikePost(postid uint64,userid uint64)bool{
  var worker sync.WaitGroup
  
  worker.Add(1)
  errchannel:=make(chan bool,1)
  go func(){
    defer worker.Done()
    tx:=db.Begin()
    r:=tx.Exec("INSERT INTO like_posts (user_id,post_id) VALUES(?,?)",userid,postid)
    if r.Error!=nil{
      tx.Rollback()
      fmt.Println("Error occured while inserting like to like_posts %w",r.Error)
      errchannel<-true
      return
    }
    errchannel<-false
  }()

  if !(<-errchannel){
    return true
  }
  
  worker.Add(1)
  go func(){
    defer worker.Done() 
    db.Exec("UPDATE posts SET num_likes=num_likes+1 WHERE post_id=?",postid)
  }()

  worker.Wait()
  return false 
}

func GetCommentsByPostID(postid uint) []UsernameAndComment {
	commentarr := make([]UsernameAndComment, 6)

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)

		var comment Comment
		go func() {
			defer wg.Done()
			db.Raw("SELECT (comment_content,user_id) FROM comments WHERE post_id=?", postid).Scan(&comment)
		}()

		wg.Add(1)
		var username string
		go func() {
			defer wg.Done()
			db.Raw("SELECT username FROM users WHERE user_id=?", comment.User_id).Scan(&username)
		}()

		rawComment := UsernameAndComment{
			UserID:          comment.User_id,
			Username:        username,
			Comment_content: comment.Comment_content,
		}
		commentarr = append(commentarr, rawComment)
	}
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

func FetchComments(postid uint64) []Comment {
	var comments []Comment
	db.Raw("SELECT (comment_content,user_id) FROM comments WHERE post_id=?", postid).Scan(&comments)
	return comments
}

