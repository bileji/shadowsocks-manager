package service

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "gopkg.in/mgo.v2"
    "encoding/json"
)

type Web struct {
    Addr   string
    DB_Con mgo.Database
}

func (w *Web) Run() {
    http.HandleFunc("/addUser", addUser)
    http.ListenAndServe(w.Addr, nil)
}

type AddU struct {
    Port     int32
    Password string
}

// todo 添加用户
func addUser(w http.ResponseWriter, r *http.Request) {

    r.ParseForm()
    if r.Method == "POST" {
        fmt.Println(r.PostFormValue("port"))
        fmt.Printf("type: %T, value: %s", r.PostFormValue("port"), r.PostFormValue("port"))

        result, _ := ioutil.ReadAll(r.Body)
        r.Body.Close()

        var Params AddU
        json.Unmarshal([]byte(result), &Params)

        fmt.Println(Params)
    } else {

    }
}

// todo 移除用户

// todo 用户列表

// todo 查看某一端口流量情况

// todo 流量统计