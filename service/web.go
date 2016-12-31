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
    http.HandleFunc("/static/multi", w.staticMulti)
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

// 查看单端口流量
func (web *Web) staticSingle(w http.ResponseWriter, r *http.Request) {

    type Params struct {
        Port           int32
        StartTimestamp string
        EndTimestamp   string
        Limit          int
    }

    if r.Method == "POST" {
        P, _ := strconv.ParseInt(r.PostFormValue("port"), 10, 64)
        L, _ := strconv.Atoi(r.PostFormValue("limit"))
        Params := Params{
            Port: int32(P),
            StartTimestamp: r.PostFormValue("start_timestamp"),
            EndTimestamp: r.PostFormValue("end_timestamp"),
            Limit: L,
        }

        if Params.Limit == 0 {
            Params.Limit = 5000
        }

        if len(Params.EndTimestamp) == 0 {
            Params.EndTimestamp = time.Now().Format("2006-01-02 15:04:05")
        }

        Resp := []bson.M{}
        if web.DB_Con.C("flows").Find(bson.M{
            "port": Params.Port,
            "created": bson.M{"$gte": Params.StartTimestamp, "$lte": Params.EndTimestamp},
        }).Select(bson.M{"size": true, "created": true}).Sort("-created").Limit(Params.Limit).All(&Resp) == nil {
            SumSize := float64(0)
            for K, Item := range Resp {
                Item["timestamp"] = Item["created"]
                delete(Item, "_id")
                delete(Item, "created")
                SumSize += Item["size"].(float64)
                Resp[K] = Item
            }

            D, _ := json.Marshal(Res{
                Code: SUCCESS,
                Data: map[string]interface{}{
                    "list": Resp,
                    "sum_size": SumSize,
                    "len": len(Resp),
                },
                Message: "flow static of port: " + strconv.Itoa(int(Params.Port)),
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

// 多端口流量统计
func (web *Web) staticMulti(w http.ResponseWriter, r *http.Request) {

    type Params struct {
        StartTimestamp string
        EndTimestamp   string
    }

    if r.Method == "POST" {
        Params := Params{
            StartTimestamp: r.PostFormValue("start_timestamp"),
            EndTimestamp: r.PostFormValue("end_timestamp"),
        }

        if len(Params.EndTimestamp) == 0 {
            Params.EndTimestamp = time.Now().Format("2006-01-02 15:04:05")
        }

        Pipe := web.DB_Con.C("flows").Pipe([]bson.M{
            {
                "$match": bson.M{"created": bson.M{"$gte": Params.StartTimestamp, "$lte": Params.EndTimestamp}},
            },
            {
                "$group": bson.M{"_id": "$port", "size": bson.M{"$sum": "$size"}},
            },
        })

        Resp := []bson.M{}
        if err := Pipe.All(&Resp); err != nil {
            D, _ := json.Marshal(Res{
                Code: FAILED,
                Data: make(map[string]interface{}),
                Message: "pipe error",
            })
            w.Write(D)
            return
        }

        D, _ := json.Marshal(Res{
            Code: SUCCESS,
            Data: map[string]interface{}{
                "list": Resp,
                "port_num": len(Resp),
            },
            Message: "success",
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