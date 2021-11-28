package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
)

type user struct {
	Username string
	Password string
}

var db *sql.DB

func initDB() (err error) {
	dsn := "user:123456@tcp(127.0.0.1:3306)/sql_test"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	return nil
}

//登录时查询
func queryRow(u user) bool{
	sqlStr := "select username, password from user where username=? and password=?"
	var uTemp user
	err := db.QueryRow(sqlStr, u.Username, u.Password).Scan(&uTemp.Username, &uTemp.Password)
	if err != nil {
		fmt.Printf("scan failed, err:%v\n", err)
		return false
	}
	fmt.Printf("username:%s passwprd:%s\n", uTemp.Username, uTemp.Password)
	return true
}


//注册时插入
func insertRow(u user) {
	sqlStr := "insert into user(username, password) values (?,?)"
	ret, err := db.Exec(sqlStr, u.Username, u.Password)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}
	id, err1 := ret.LastInsertId()
	if err1 != nil {
		fmt.Printf("get lastinsert ID fauled, err:%v\n", err)
		return
	}
	fmt.Printf("insert success, the id is %d.\n", id)
}

//中间件 提示用户是否登录
func mid() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("username")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "游客你好!",
			})
			c.Next()
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"message": cookie.Value + "你好！",
		})
		c.Abort()
	}
}


//登录
func login(c *gin.Context) {
	var u user
	u.Username = c.Query("username")
	u.Password = c.Query("password")
	if queryRow(u) {
		cookie := &http.Cookie{
			Name: "username",
			Value: u.Username,
			MaxAge: 1000,
			Path: "/",
			HttpOnly: true,
		}
		http.SetCookie(c.Writer, cookie)
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"message": "登录成功！",
		})
	} else {
		c.JSON(http.StatusForbidden, gin.H{
			"status": http.StatusForbidden,
			"message": "用户名与密码有误！",
		})
	}
}

//注册
func sign(c *gin.Context) {
	var u user
	u.Username = c.Query("username")
	u.Password = c.Query("password")
	if queryRow(u) {
		c.JSON(http.StatusForbidden, gin.H{
			"status": http.StatusForbidden,
			"message": "已存在该用户名！",
		})
	} else {
		insertRow(u)
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"message": "注册成功！",
		})
	}
}

//发布文章
func postBlog(c *gin.Context) {
	username, err := c.Cookie("username")
	if err != nil{
		c.JSON(http.StatusForbidden, gin.H{
			"status": http.StatusForbidden,
			"message": "请先登录！",
		})
		return
	}
	title := c.Query("title")
	content := c.Query("content")
	if content == "" || title == ""{
		c.JSON(http.StatusForbidden, gin.H{
			"status": http.StatusForbidden,
			"message": "发表失败",
			"reason": "title or content 为空",
		})
		return
	}
	sqlStr := "insert into blog(username, title, content) values (?,?,?)"
	_ , err = db.Exec(sqlStr, username, title, content)
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"message": "发表失败！",
			"err": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"message": "发表成功！",
		"username": username,
		"title": title,
		"content": content,
	})
}

//给其他人点赞
func likeOther(c *gin.Context) {
	u, erro := c.Cookie("username")
	if erro != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"status": http.StatusForbidden,
			"message": "请先登录！",
		})
		return
	}
	username := c.Query("username")
	title := c.Query("title")

	if u == username{
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"message": "你不能给自己点赞！",
		})
		return
	}

	sqlStr := "update blog set favor=favor+1 where username=? and title = ?"
	ret, err := db.Exec(sqlStr, username, title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"message": "点赞失败！",
			"err": err,
		})
		return
	}
	n, err1 := ret.RowsAffected()
	if err1 != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"message": "点赞失败！",
		})
		return
	}
	if n == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"status": http.StatusForbidden,
			"message": "点赞失败！",
			"reason": "文章不存在!",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "给" + username + "点赞成功！",
		"title": title,
	})
}


func main() {
	err := initDB()
	if err != nil {
		fmt.Printf("init db failed,err:%v\n", err)
		return
	}
	r := gin.Default()

	//主页
	r.GET("/home", mid(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"message": "主页",
		})
	})

	//登录
	r.POST("/login", login)

	//注册
	r.POST("/sign", sign)

	//发布文章
	r.POST("/post", postBlog)

	//点赞文章
	r.PUT("/favor", likeOther)

	r.Run(":9090")
}