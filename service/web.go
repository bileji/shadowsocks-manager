package service

import (
    "net/http"
    "gopkg.in/mgo.v2"
    "strconv"
    "encoding/json"
    "shadowsocks-manager/manager"
    "time"
    "gopkg.in/mgo.v2/bson"
)

var (
    FAILED int32 = 0
    SUCCESS int32 = 200
)

type Web struct {
    Addr   string
    DB_Con *mgo.Database
}

type Res struct {
    Code    int32 `json:"code"`
    Data    map[string]interface{} `json:"data"`
    Message string `json:"message"`
}

func (w *Web) Run() {
    http.HandleFunc("/user/add", w.addUser)
    http.HandleFunc("/static/single", w.staticSingle)
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
            return
        } else {
            // todo 判断

            Count, _ := web.DB_Con.C("users").Find(bson.M{
                "$or": []bson.M{
                    {
                        "port": Params.Port,
                    },
                    {
                        "username": Params.Username,
                    },
                },
            }).Count()
            if Count > 0 {
                D, _ := json.Marshal(Res{
                    Code: FAILED,
                    Data: make(map[string]interface{}),
                    Message: "the username/port is occupied",
                })
                w.Write(D)
                return
            }

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
                return
            }
        }

        D, _ := json.Marshal(Res{
            Code: SUCCESS,
            Data: make(map[string]interface{}),
            Message: "add success",
        })
        w.Write(D)
        return
    } else {
        D, _ := json.Marshal(Res{
            Code: FAILED,
            Data: make(map[string]interface{}),
            Message: "required method post",
        })
        w.Write(D)
        return
    }
}

// todo 移除用户
func (web *Web) stopUser() {

}

// todo 用户列表
func (web *Web) editUser() {

    type Params struct {
        Username  string
        Port      int32
        Password  string
        AllowSize float64
        Status    bool
    }

}

// todo 查看某一端口流量情况
func (web *Web) staticSingle(w http.ResponseWriter, r *http.Request) {

    type Params struct {
        Port           int32
        StartTimestamp string
        EndTimestamp   string
    }

    if r.Method == "POST" {
        P, _ := strconv.ParseInt(r.PostFormValue("port"), 10, 64)
        Params := Params{
            Port: int32(P),
            StartTimestamp: r.PostFormValue("start_timestamp"),
            EndTimestamp: r.PostFormValue("end_timestamp"),
        }

        Resp := []bson.M{}
        if web.DB_Con.C("flows").Find(bson.M{
            "port": Params.Port,
            "created": bson.M{"$gte": Params.StartTimestamp, "$lte": Params.EndTimestamp},
        }).Select(bson.M{"size": true, "created": true}).All(&Resp) == nil {

            SumSize := float64(0)
            //Static := make(map[string]interface{})

            for K, Item := range Resp {
                Item["timestamp"] = Item["created"]
                delete(Item, "_id")
                delete(Item, "created")
                SumSize += Item["size"].(float64)
                Resp[K] = Item
            }

            //Static["sum"] = SumSize
            //Static["list"] = Resp

            D, _ := json.Marshal(Res{
                Code: SUCCESS,
                Data: map[string]interface{}{
                    "list": Resp,
                    "sum_size": SumSize,
                },
                Message: "flow static of port(" + strconv.Itoa(int(Params.Port)) + ")",
            })
            w.Write(D)
            return
        } else {
            D, _ := json.Marshal(Res{
                Code: FAILED,
                Data: make(map[string]interface{}),
                Message: "query failed",
            })
            w.Write(D)
            return
        }
    } else {
        D, _ := json.Marshal(Res{
            Code: FAILED,
            Data: make(map[string]interface{}),
            Message: "required method post",
        })
        w.Write(D)
        return
    }
}

// todo 流量统计
func (web *Web) staticMulti() {

}