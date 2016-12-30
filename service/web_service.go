package service

import (
    "net/http"
    "gopkg.in/mgo.v2"
)

type Web struct {
    Addr   string
    DB_Con mgo.Database
}

func (w *Web) Run() {
    http.ListenAndServe(w.Addr, nil)
}

// todo 添加用户

// todo 移除用户

// todo 用户列表

// todo 查看某一端口流量情况

// todo 流量统计