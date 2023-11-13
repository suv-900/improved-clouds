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
func GetPostsByUserId(userid uint64) []Posts {
	var posts []Posts
	db.Raw("SELECT (post_title,post_content) FROM posts WHERE authorid=? LIMIT 5", userid).Scan(&posts)
	return posts
}

func GetPostAndUserPreferences(postid uint64, userid uint64) (Posts, string, error) {
	var post Posts
	var err error
	var username string
	a := make(chan int, 1)
	go func() {
		r := db.Raw("SELECT post_title,post_content,author_id,post_likes FROM posts WHERE post_id=?", postid).Scan(&post)
		err = r.Error
		a <- 1
	}()
	<-a
	if err != nil {
		return post, username, err
	}

	b := make(chan int, 1)
	go func() {
		r := db.Raw("SELECT username FROM users WHERE user_id=?", post.Author_id).Scan(&username)
		err = r.Error
		b <- 1
	}()
	<-b
	if err != nil {
		return post, username, err
	}

	return post, username, nil
}

func Check_if_user_likedPost(userid uint64, postid uint64) (bool, bool, error) {
	var userLikedPost bool
	var userDislikedPost bool
	var err error
	a := make(chan int, 1)
	go func() {

		r := db.Raw("SELECT liked from posts_liked_by_user WHERE user_id=? AND post_id=?", userid, postid).Scan(&userLikedPost)
		if r.Error != nil {
			err = r.Error
			a <- 1
			return
		}
		db.Raw("SELECT disliked FROM posts_disliked_by_user WHERE user_id=? AND post_id=?", userid, postid).Scan(&userDislikedPost)
		err = r.Error
		a <- 1
	}()
	<-a

	return userLikedPost, userDislikedPost, err
}
func PostById(postid uint64) (Posts, string, error) {
	var post Posts
	var username string
	r := db.Raw("SELECT post_title,post_content,author_id,post_likes FROM posts WHERE post_id=?", postid).Scan(&post)
	if r.Error != nil {
		return post, username, r.Error
	}
	r = db.Raw("SELECT username FROM users WHERE user_id=?", post.Author_id).Scan(&username)
	return post, username, r.Error
}

func LikePostByID(userid uint64, postid uint64) error {
	var err error

	c := make(chan int, 1)
	go func() {
		r := db.Exec("DELETE FROM posts_disliked_by_user WHERE user_id=? AND post_id=?", userid, postid)
		if r.Error != nil {
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

	a := make(chan int, 1)
	go func() {
		tx := db.Begin()
		r := tx.Exec("UPDATE posts SET post_likes=post_likes+1 WHERE post_id=?", postid)
		if r.Error != nil {
			err = r.Error
			a <- 1
			tx.Rollback()
			return
		}
		tx.Commit()
		a <- 1
	}()
	<-a
	if err != nil {
		return err
	}

	b := make(chan int, 1)
	go func() {
		tx := db.Begin()
		r := tx.Exec("INSERT INTO posts_liked_by_user (user_id,post_id,liked) VALUES(?,?,?)", userid, postid, true)
		if r.Error != nil {
			err = r.Error
			b <- 1
			tx.Rollback()
			return
		}
		tx.Commit()
		b <- 1
	}()
	<-b
	return err
}

func RemoveLikeFromPost(userid uint64, postid uint64) error {
	r := db.Exec("UPDATE posts SET post_likes=post_likes-1 WHERE post_id=?", postid)
	if r.Error != nil {
		return r.Error
	}
	r = db.Exec("DELETE FROM posts_disliked_by_user WHERE post_id=? AND user_id=?", postid, userid)
	if r.Error != nil {
		return r.Error
	}
	return nil
}
func DislikePostByID(userid uint64, postid uint64) error {
	var err error

	c := make(chan int, 1)
	go func() {
		r := db.Exec("DELETE FROM posts_liked_by_user WHERE user_id=? AND post_id=?", userid, postid)
		if r.Error != nil {
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

	a := make(chan int, 1)
	go func() {
		tx := db.Begin()
		r := tx.Exec("UPDATE posts SET post_likes=post_likes-1 WHERE post_id=?", postid)
		if r.Error != nil {
			err = r.Error
			tx.Rollback()
			a <- 1
			return
		}
		tx.Commit()
		a <- 1
	}()
	<-a

	if err != nil {
		return err
	}

	b := make(chan int, 1)
	go func() {
		tx := db.Begin()
		r := tx.Exec("INSERT INTO posts_disliked_by_user (post_id,user_id,disliked) VALUES(?,?,?)", postid, userid, true)

		if r.Error != nil {
			err = r.Error
			tx.Rollback()
			b <- 1
			return
		}
		tx.Commit()
		b <- 1

	}()
	<-b
	return err
}
func RemoveDislikeFromPost(userid uint64, postid uint64) error {
	r := db.Exec("UPDATE posts SET post_likes=post_likes+1 WHERE post_id=?", postid)
	if r.Error != nil {
		return r.Error
	}
	r = db.Exec("DELETE FROM posts_disliked_by_user WHERE post_id=? AND user_id=?", postid, userid)
	if r.Error != nil {
		return r.Error
	}
	return nil

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
