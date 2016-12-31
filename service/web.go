package service

import (
    "net/http"
    "gopkg.in/mgo.v2"
    "strconv"
    "encoding/json"
    "shadowsocks-manager/manager"
    "time"
)

var (
    FAILED int32 = 0
    SUCCESS int32 = 200
)

type Web struct {
    Addr   string
    DB_Con mgo.Database
}

type Res struct {
    Code    int32 `json:"code"`
    Data    map[string]interface{} `json:"data"`
    Message string `json:"message"`
}

func (w *Web) Run() {
    http.HandleFunc("/addUser", w.addUser)
    http.ListenAndServe(w.Addr, nil)
}

// 添加用户
func (web *Web) addUser(w http.ResponseWriter, r *http.Request) {

    type Params struct {
        Username  string
        Port      int32
        Password  string
        AllowSize float64
    }

    if r.Method == "POST" {
        P, _ := strconv.ParseInt(r.PostFormValue("port"), 10, 64)
        A, _ := strconv.ParseFloat(r.PostFormValue("allowsize"), 64)
        Params := &Params{
            Username: r.PostFormValue("username"),
            Port: int32(P),
            Password: r.PostFormValue("password"),
            AllowSize: A * 1000000,
        }

        if len(Params.Username) == 0 || len(Params.Password) == 0 || Params.Port == 0 {
            D, _ := json.Marshal(Res{
                Code: FAILED,
                Data: make(map[string]interface{}),
                Message: "required field username/password/port",
            })
            w.Write(D)
        } else {
            // todo 判断

            err := web.DB_Con.C("users").Insert(manager.User{
                Username: Params.Username,
                Port: Params.Port,
                Status: true,
                Password: Params.Password,
                AllowSize: Params.AllowSize,
                Created: time.Now().Format("2006-01-02 15:04:05"),
                Modified: time.Now().Format("2006-01-02 15:04:05"),
            })
            if err != nil {
                D, _ := json.Marshal(Res{
                    Code: FAILED,
                    Data: make(map[string]interface{}),
                    Message: "save failed",
                })
                w.Write(D)
            }
        }

        D, _ := json.Marshal(Res{
            Code: SUCCESS,
            Data: make(map[string]interface{}),
            Message: "add success",
        })
        w.Write(D)
    } else {
        D, _ := json.Marshal(Res{
            Code: FAILED,
            Data: make(map[string]interface{}),
            Message: "required method post",
        })
        w.Write(D)
    }
}

// todo 移除用户

// todo 用户列表

// todo 查看某一端口流量情况

// todo 流量统计