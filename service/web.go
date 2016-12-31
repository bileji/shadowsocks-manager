package service

import (
    "fmt"
    "net/http"
    "gopkg.in/mgo.v2"
    "encoding/json"
    //"io/ioutil"
)

type Web struct {
    Addr   string
    DB_Con mgo.Database
}

func (w *Web) Run() {
    http.HandleFunc("/addUser", addUser)
    http.ListenAndServe(w.Addr, nil)
}

type AddUserParams  struct {
    Username string `json:"username"`
    Port     string `json:"port"`
    Password string `json:"password"`
}

// todo 添加用户
func addUser(w http.ResponseWriter, r *http.Request) {
    fmt.Println(r.PostFormValue("port"), r.PostForm)

    if r.Method == "POST" {
        var Params AddUserParams
        Args, err := json.Marshal(r.PostForm)
        r.ParseForm()
        if err == nil {
            if json.Unmarshal(Args, &Params) == nil {
                fmt.Println(Params)
            }
        }
    }

    //fmt.Println(r.PostFormValue("port"), r.Method, r.Form)
    //fmt.Printf("type: %T, value: %s", r.PostFormValue("port"), r.PostFormValue("port"))
    //r.ParseForm()
    //if r.Method == "POST" {
    //    result, _ := ioutil.ReadAll(r.Body)
    //    r.Body.Close()
    //
    //    var Params AddUserParams
    //    json.Unmarshal([]byte(result), &Params)
    //
    //    fmt.Println(Params)
    //} else {
    //
    //}
}

// todo 移除用户

// todo 用户列表

// todo 查看某一端口流量情况

// todo 流量统计