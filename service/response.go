package service

import (
    "net/http"
    "encoding/json"
)

type Response struct {
    Code    int32 `json:"code"`
    Data    map[string]interface{} `json:"data"`
    Message string `json:"message"`
}

func (r Response) Json(w http.ResponseWriter) {
    w.Header().Set("Content-type", "application/json")
    bytes, _ := json.Marshal(r)
    w.Write(bytes)
}
