# 博客系统

## 功能：主页、注册、登录、发表文章、为文章点赞

### 1.注册 127.0.0.1:9090/sign   post请求

```go
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
```



输入username, password(Query) 在user表中查询
若用户名不存在 则注册成功 同时将用户名和密码 写入数据库user表

![sign注册成功](D:\GoProjects\src\csa_4\测试图\sign注册成功.png)

若用户名存在  提示：用户名已存在

![sign注册失败](D:\GoProjects\src\csa_4\测试图\sign注册失败.png)

### 2.登录 127.0.0.1:9090/login  post请求

```go
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
```

输入username, password(Query) 在user表中查询
若查询失败   提示：用户名或密码错误

![login登录失败](D:\GoProjects\src\csa_4\测试图\login登录失败.png)

若查询成功   提示：登录成功

![login登录成功](D:\GoProjects\src\csa_4\测试图\login登录成功.png)



3.主页  127.0.0.1:9090/home  get请求

```go
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
```

读取cookie数据
若有      提示：username你好！

![home登录后](D:\GoProjects\src\csa_4\测试图\home登录后.png)

若无      提示：游客你好！

![home未登录时](D:\GoProjects\src\csa_4\测试图\home未登录时.png)

4.发表文章 127.0.0.1:9090/post  post请求

```go
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
```

读取cookie数据
若无               提示：请先登录！ 并返回

![post未登录时](D:\GoProjects\src\csa_4\测试图\post未登录时.png)

输入title, content(Query)
若title, content为空       提示：title or content 为空 并返回

![post登录后发表文章失败](D:\GoProjects\src\csa_4\测试图\post登录后发表文章失败.png)

title, content都不为空,则将cookie中读取的value, title, content写入数据库blog表中 

![post登录后发表文章成功](D:\GoProjects\src\csa_4\测试图\post登录后发表文章成功.png)



### 5.为文章点赞 127.0.0.1:9090/favor put请求

```go
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
```

读取cookie数据
若无               提示：请先登录！

![favor未登录时](D:\GoProjects\src\csa_4\测试图\favor未登录时.png)

输入username, title(Query)
若username与cookie中value相同   提示：你不能给自己点赞！  并返回

![favor点赞失败1](D:\GoProjects\src\csa_4\测试图\favor点赞失败1.png)

在数据库blog表中查询username, title
若查询失败            提示：点赞失败！文章不存在！ 并返回

![favor点赞失败2](D:\GoProjects\src\csa_4\测试图\favor点赞失败2.png)

若查询成功 将blog表中对应的favor++     提示：给username点赞成功

![favor点赞成功](D:\GoProjects\src\csa_4\测试图\favor点赞成功.png)



