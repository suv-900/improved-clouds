create table users(
    user_id bigserial primary key,
    username varchar(20)  unique,
    email varchar(40) not null,
    passwordhash varchar(30) not null,
    userabout varchar(100),
    createdAt timestamp,
    updatedAt timestamp,
)
create table posts(
    post_id bigserial primary key,
    authorid bigint foreign key,
    post_title varchar(20) not null,
    post_content text,
    post_category varchar(60) not null,
--    likes int,
    createdAt timestamp,
    updatedAt timestamp,

  )
create table comments(
    comment_id bigserial primary key,
    user_id int,
    post_id int,
    comment_content text not null,
    createdAt timestamp,
    updatedAt timestamp
)
alter table comment add foreign_key "user_id" references users("user_id")
alter table comment add foreign_key "post_id" references posts("post_id")



create table category(
    category_id bigserial primary key,
    category_name varchar(60) ,
    post_count int,
    createdAt timestamp,
    updatedAt timestamp,
)
insert into category(catergory_name,post_count) values("technolodgy",0)
insert into category(catergory_name,post_count) values("technolodgy",0)
insert into category(catergory_name,post_count) values("technolodgy",0)
insert into category(catergory_name,post_count) values("technolodgy",0)
insert into category(catergory_name,post_count) values("technolodgy",0)

alter table posts add foreign_key ""

alter table posts add foreign_key "authorid" references 
users("user_id");
