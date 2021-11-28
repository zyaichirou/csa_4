//@Title		main.go
//@Description	实现一个简单的博客系统
//@Author		zy
//@Update		2021.11.28


package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
)

//User 	用户对象
type user struct {
	Username string		//用户名
	Password string		//密码
}

//定义一个全局对象db方便调用
var db *sql.DB

//@title		initDB()
//@description	连接数据库
//@author		zy
//@param
//@return		err error
func initDB() (err error) {
	dsn := "user:123456@tcp(127.0.0.1:3306)/sql_test"
	db, err = sql.Open("mysql", dsn)			//打开指定的数据库
	if err != nil {
		return err
	}
	err = db.Ping()		//是否成功连接
	if err != nil {
		return err
	}
	return nil
}

//@title		queryRow()
//@description	查询数据库相应的数据
//@author		zy
//@param		u user
//@return		bool
func queryRow(u user) bool{
	sqlStr := "select username, password from user where username=? and password=?"		//sql语句
	var uTemp user
	err := db.QueryRow(sqlStr, u.Username, u.Password).Scan(&uTemp.Username, &uTemp.Password)  //调用QueryRow进行插入
	if err != nil {
		fmt.Printf("scan failed, err:%v\n", err)
		return false
	}
	fmt.Printf("username:%s passwprd:%s\n", uTemp.Username, uTemp.Password)
	return true
}

//@title		insertRow()
//@description	注册时将信息插入到数据库相应的表中
//@author		zy
//@param		u user
//@return
func insertRow(u user) {
	sqlStr := "insert into user(username, password) values (?,?)"	//sql语句
	ret, err := db.Exec(sqlStr, u.Username, u.Password)				//执行一次命令
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

//@title		mid()
//@description	中间件 提示用户是否登录
//@author		zy
//@param
//@return		gin.HandlerFunc
func mid() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("username")		//读取cookie
		if err != nil {											//若当前无cookie相关内容则跳转到下一个
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "游客你好!",
			})
			c.Next()											//跳转
			return
		}
		c.JSON(http.StatusOK, gin.H{							//若当前有cookie相关内容
			"code": 200,
			"message": cookie.Value + "你好！",
		})
		c.Abort()												//终止
	}
}


//@title		login()
//@description	登录界面  查询数据库中相应表的数据，若存在，则登录成功同时写入cookie 反之，提示登录失败
//@author		zy
//@param		c *gin.Context
//@return
func login(c *gin.Context) {
	var u user
	u.Username = c.Query("username")				//请求数据
	u.Password = c.Query("password")				//请求数据
	if queryRow(u) {									//若存在则写入cookie
		cookie := &http.Cookie{
			Name: "username",
			Value: u.Username,
			MaxAge: 1000,
			Path: "/",
			HttpOnly: true,
		}
		http.SetCookie(c.Writer, cookie)				//写cookie
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"message": "登录成功！",
		})
	} else {											//提示错误
		c.JSON(http.StatusForbidden, gin.H{
			"status": http.StatusForbidden,
			"message": "用户名与密码有误！",
		})
	}
}

//@title		sign()
//@description	注册  查询数据库中相应表， 若已存在则提示用户已存在  反之将数据插入到数据库中相应的表
//@author		zy
//@param		c *gin.Context
//@return
func sign(c *gin.Context) {
	var u user
	u.Username = c.Query("username")
	u.Password = c.Query("password")
	if queryRow(u) {							//若查询到已存在该用户名
		c.JSON(http.StatusForbidden, gin.H{
			"status": http.StatusForbidden,
			"message": "已存在该用户名！",
		})
	} else {
		insertRow(u)							//将用户注册的信息插入到数据库
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"message": "注册成功！",
		})
	}
}

//@title		postBlog()
//@description	发表文章
//@author		zy
//@param		c *gin.Context
//@return
func postBlog(c *gin.Context) {
	username, err := c.Cookie("username")
	if err != nil{									//当前无cookie相关内容  提示用户登录
		c.JSON(http.StatusForbidden, gin.H{
			"status": http.StatusForbidden,
			"message": "请先登录！",
		})
		return
	}
	title := c.Query("title")
	content := c.Query("content")
	if content == "" || title == ""{				//内容或标题为空      提示错误
		c.JSON(http.StatusForbidden, gin.H{
			"status": http.StatusForbidden,
			"message": "发表失败",
			"reason": "title or content 为空",
		})
		return
	}
	sqlStr := "insert into blog(username, title, content) values (?,?,?)"			//sql语句
	_ , err = db.Exec(sqlStr, username, title, content)								//将信息插入到数据库
	if err != nil{																	//若插入有误  提示错误
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"message": "发表失败！",
			"err": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{													//发表成功
		"status": http.StatusOK,
		"message": "发表成功！",
		"username": username,
		"title": title,
		"content": content,
	})
}

//@title		likeOther()
//@description	给博客点赞			以username和title为key
//@author		zy
//@param		c *gin.Context
//@return
func likeOther(c *gin.Context) {
	u, erro := c.Cookie("username")
	if erro != nil {								//当前无cookie相关内容  提示用户登录
		c.JSON(http.StatusForbidden, gin.H{
			"status": http.StatusForbidden,
			"message": "请先登录！",
		})
		return
	}
	username := c.Query("username")
	title := c.Query("title")

	if u == username{								//读取的cookie与请求数据中的username一致  即给自己点赞 不允许
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"message": "你不能给自己点赞！",
		})
		return
	}

	sqlStr := "update blog set favor=favor+1 where username=? and title = ?"			//sql语句
	ret, err := db.Exec(sqlStr, username, title)										//更新对应的值——favor++
	if err != nil {																		//存在错误 提示用户
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"message": "点赞失败！",
			"err": err,
		})
		return
	}
	n, err1 := ret.RowsAffected()														//判断更新了几行相应的值 即判断username和title是否存在
	if err1 != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"message": "点赞失败！",
		})
		return
	}
	if n == 0 {																			//没有相应的博客
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
