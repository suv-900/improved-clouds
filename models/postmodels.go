package models

import (
	"errors"
	"fmt"
)

func CreatePost(post Posts) (uint64, error) {
	var postid uint64
	pipe := make(chan bool, 1)
	var err error
	var userexists bool

	a := make(chan int, 1)
	go func() {
		r := CheckUserExists(post.Author_id)
		userexists = r
		a <- 1
	}()
	<-a
	if !userexists {
		fmt.Println("User doesnt exists.Post Creation Failed")
		return 0, errors.New("user doesnt exists.Post Creation Failed")
	}

	go func() {
		tx := db.Begin()
		result := tx.Exec("INSERT INTO posts (post_title,post_content,author_id,post_likes) VALUES(?,?,?,?) RETURNING post_id", post.Post_title, post.Post_content, post.Author_id, 0).Scan(&postid)
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

// deletes a post
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
	db.Raw("SELECT post_title,post_content,author_id FROM posts WHERE post_id=?", postid).Scan(&post)
	db.Raw("SELECT username FROM users WHERE user_id=?", post.Author_id).Scan(&username)
	return post, username
}

func LikePostByID(postid uint64) {
	db.Exec("UPDATE posts SET post_likes=post_likes+1 WHERE post_id=?", postid)
}
func DislikePostByID(postid uint64) {
	db.Exec("UPDATE posts SET post_likes=post_likes-1 WHERE post_id=?", postid)
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
