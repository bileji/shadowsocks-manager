package service

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "gopkg.in/mgo.v2"
)

type Web struct {
    Addr   string
    DB_Con mgo.Database
}

func (w *Web) Run() {
    http.HandleFunc("/addUser", addUser)
    http.ListenAndServe(w.Addr, nil)
}

// todo 添加用户
func addUser(w http.ResponseWriter, r *http.Request) {
    r.PostForm()
    if r.Method == "POST" {
        result, _ := ioutil.ReadAll(r.Body)
        r.Body.Close()
        fmt.Printf("%s\n", result)
    } else {

    }
}

// todo 移除用户

// todo 用户列表

// todo 查看某一端口流量情况

// todo 流量统计