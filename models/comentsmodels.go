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

func LikeAComment(commentid uint64) {
	db.Exec("UPDATE comments SET comment_likes=comment_likes+1 WHERE comment_id=?", commentid)
}

func DislikeAComment(commentid uint64) {
	db.Exec("UPDATE comments SET comment_likes=comment_likes-1 WHERE comment_id=?", commentid)
}

func GetCommentsByPostID(postid uint64) []UsernameAndComment {
	commentarr := make([]UsernameAndComment, 5)
	//TODO OFFSET to hold a bar for next comments
	a := make(chan int, 1)
	go func() {
		sql := "SELECT (comment_id,user_id,username,comment_content) FROM comments WHERE post_id=? ORDER BY comment_likes DESC LIMIT 5 "
		db.Raw(sql, postid).Scan(&commentarr)
		a <- 1
	}()
	<-a
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
