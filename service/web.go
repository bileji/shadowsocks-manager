package service

import (
    "net/http"
    "gopkg.in/mgo.v2"
    "strconv"
    "encoding/json"
)

type Web struct {
    Addr   string
    DB_Con mgo.Database
}

type Res struct {
    Code    int32 `json:"code"`
    Message string `json:"message"`
    Data    map[string]interface{} `json:"data"`
}

func (w *Web) Run() {
    http.HandleFunc("/addUser", addUser)
    http.ListenAndServe(w.Addr, nil)
}

// 添加用户
func addUser(w http.ResponseWriter, r *http.Request) {

    type Params struct {
        Username string
        Port     int32
        Password string
    }

    if r.Method == "POST" {
        P, _ := strconv.Atoi(r.PostFormValue("port"))
        Params := &Params{
            Username: r.PostFormValue("username"),
            Port: int32(P),
            Password: r.PostFormValue("password"),
        }

        if len(Params.Username) == 0 || len(Params.Password) == 0 || Params.Port == 0 {
            D, _ := json.Marshal(Res{
                Code: 0,
                Message: "required field username/password/port",
                Data: make(map[string]interface{}),
            })
            w.Write(D)
        } else {

        }

        //Args, err := json.Marshal(r.PostForm)
        //r.ParseForm()
        //if err == nil {
        //    err = json.Unmarshal(Args, &Params)
        //    fmt.Println(err)
        //    if err == nil {
        //        fmt.Println(Params)
        //    }
        //}
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