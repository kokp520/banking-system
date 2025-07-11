# bank system

## 前言

因為時間有限打算用比較熟悉的語言 Golang，同時可以應付高併發請求負載。
為了快速產出以及需求上可能會有擴充的需求，以及Docker部署問題，決定用Gin框架快速開發。
>會參考 go-gin-example架構



## system design

### 必須：
- docker
- unit test, integration test
- 原子級別transcation
- logger(when, who, how much)
- func for Transfer balance to another account
- func withdraw
- func deposit
- func Create (name,balance)
- balance cannot be negative
- restful api

### Great to have

- rate limit
- JWT

### 流程

Web(client) -> Nginx -> Gin server -> redis -> MySql

