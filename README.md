��# csa_4

功能：主页、注册、登录、发表文章、为文章点赞

1.注册  127.0.0.1:9090/sign     post请求
输入username, password(Query) 在user表中查询
若用户名不存在  则注册成功 同时将用户名和密码 写入数据库user表
若用户名存在    提示：用户名已存在

2.登录  127.0.0.1:9090/login    post请求
输入username, password(Query) 在user表中查询
若查询失败      提示：用户名或密码错误
若查询成功      提示：登录成功

3.主页   127.0.0.1:9090/home    get请求
读取cookie数据
若有            提示：username你好！
若无            提示：游客你好！

4.发表文章 127.0.0.1:9090/post   post请求
读取cookie数据
若无                             提示：请先登录！ 并返回
输入title, content(Query)
若title, content为空             提示：title or content 为空  并返回
title, content都不为空,则将cookie中读取的value, title, content写入数据库blog表中  


5.为文章点赞 127.0.0.1:9090/favor put请求
读取cookie数据
若无                             提示：请先登录！
输入username, title(Query)
若username与cookie中value相同     提示：你不能给自己点赞！   并返回
在数据库blog表中查询username, title
若查询失败                       提示：点赞失败！文章不存在！  并返回
若查询成功  将blog表中对应的favor++  提示：给username点赞成功





