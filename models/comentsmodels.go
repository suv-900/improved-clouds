package models

import (
	"fmt"
	"sync"
)

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

// call responsively
func LikeAComment(commentid uint64) {
	db.Exec("UPDATE comments SET comment_likes=comment_likes+1 WHERE comment_id=?", commentid)
}

func DislikeAComment(commentid uint64) {
	db.Exec("UPDATE comments SET comment_likes=comment_likes-1 WHERE comment_id=?", commentid)
}

func Get5CommentsByPostID(postid uint64) []UsernameAndComment {
	//commentarr := make([]UsernameAndComment, 5)
	commentsvec := []UsernameAndComment{}
	//TODO OFFSET to hold a bar for next comments
	a := make(chan int, 1)
	go func() {
		sql := "SELECT comment_id,user_id,username,comment_content FROM comments WHERE post_id=? ORDER BY comment_likes DESC LIMIT 5 "
		db.Raw(sql, postid).Scan(&commentsvec)
		a <- 1
	}()
	<-a
	return commentsvec
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
}

func GetAllCommentsByPostID(postid uint64) []UsernameAndComment {
	//commentarr := make([]UsernameAndComment, 5)
	commentsvec := []UsernameAndComment{}
	//TODO OFFSET to hold a bar for next comments
	a := make(chan int, 1)
	go func() {
		sql := "SELECT comment_id,user_id,username,comment_content,comment_likes FROM comments WHERE post_id=? ORDER BY comment_likes DESC "
		db.Raw(sql, postid).Scan(&commentsvec)
		a <- 1
	}()
	<-a
	return commentsvec
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
}

func AddComment(postid uint64, userid uint64, username string, comment string) error {
	var err error
	err = nil
	a := make(chan int, 1)
	go func() {
		tx := db.Begin()
		r := tx.Exec("INSERT INTO comments (user_id,post_id,username,comment_content,comment_likes) VALUES(?,?,?,?,?) ", userid, postid, username, comment, 0)
		if r.Error != nil {
			tx.Rollback()
			err = r.Error
			a <- 1
		} else {
			tx.Commit()
			a <- 1
		}
	}()
	<-a
	return err
}

func EditComment(commentId uint64, comment string) {
	db.Exec("UPDATE comments SET comment_content=? WHERE comment_id=?", commentId, comment)
}

func RemoveComment(commentId uint64) {
	db.Exec("DELETE * FROM comments WHERE comment_id=?", commentId)
}
