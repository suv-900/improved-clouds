CREATE TABLE users (
    user_id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    profile_pic BYTEA,  
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    userabout TEXT,
    
	createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE categories (
    category_id BIGSERIAL PRIMARY KEY,
    category_name VARCHAR(50) NOT NULL
);

CREATE TABLE posts (
    post_id BIGSERIAL PRIMARY KEY,
    user_id INT,
    category_id INT,
	post_title varchar(100) NOTNULL,
	post_content TEXT,
	num_likes INT DEFAULT 0,
	createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (category_id) REFERENCES categories(category_id)
);

CREATE TABLE post_likes (
    like_id BIGSERIAL PRIMARY KEY,
    user_id INT,
    post_id INT,
	createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (post_id) REFERENCES posts(post_id)
);

CREATE TABLE comments(
	comment_id BIGSERIAL PRIMARY KEY,
	user_id INT REFERENCES users(user_id),
	username VARCHAR(100) REFERENCES users(username),
	post_id INT REFERENCES posts(post_id),
	comment_content TEXT,
	comment_likes INT,
	createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
);

CREATE TABLE comment_likes(
	commentlike_id BIGSERIAL PRIMARY KEY,
	comment_id REFERENCES comments(comment_id),
	createdAt TIMESTAMP CURRENT_TIMESTAMP,
);


CREATE TABLE comment_replies(
	reply_id BIGSERIAL PRIMARY KEY,
	comment_id INT REFERENCES post_comments(comment_id),
	user_id INT REFERENCES users(user_id),
	reply_content TEXT,
	createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	
);