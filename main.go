package main

import (
    "fmt"
    "time"
    "strconv"
    "strings"
    "encoding/json"
    "shadowsocks-manager/manager"
    "gopkg.in/mgo.v2/bson"
)

func main() {
    err, Connector := manager.ConnectToMgo("127.0.0.1", "vpn", "shadowsocks", "mlgR4evB")

    if err != nil {
        panic(err)
    }

    defer Connector.Close()

    USock := manager.UnixSock{
        Net: "unixgram",
        LSock: "/var/run/manager.sock",
        RSock: "/var/run/shadowsocks-manager.sock",
        Collection: Connector.DB("vpn").C("flows"),
    }

    USock.Listen()
    go USock.Ping()

    // todo每1分钟检查流量是否超标
    go USock.HeartBeat(5, func() error {
        Ports := []int32{}
        Users := []manager.User{}

        UserModel := Connector.DB("vpn").C("users")

        if UserModel.Find(bson.M{"Port": int32(8388)}).Select(bson.M{"Port": 1}).All(&Users) == nil {
            fmt.Println(Users)
            for _, User := range Users {
                Ports = append(Ports, User.Port)
            }
        }

        if len(Ports) > 0 {
            StartTime, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
            fmt.Println(StartTime.Format("2006-01-02 15:04:05"))
            FlowModel := Connector.DB("vpn").C("flows")
            Pipe := FlowModel.Pipe([]bson.M{
                {
                    "$match": bson.M{"Created": bson.M{"$gt": StartTime.Format("2006-01-02 15:04:05")}},
                },
                {
                    "$group": bson.M{"_id": "$Port", "total": bson.M{"$sum": "$Size"}},
                },
            })

            Resp := []bson.M{}
            if Pipe.All(&Resp) != nil {
                // todo print err info
            }

            fmt.Println(Resp)
        } else {
            fmt.Println("collection users is null")
        }
        return nil
    })

    // 监听各端口流量情况
    go USock.Rec(func(buffer []byte) {
        M := make(map[string]interface{})
        if Message := strings.TrimLeft(string(buffer), "stat: "); strings.EqualFold(Message, "pong") {
            fmt.Println("start the program: shadowsocks-manager")
        } else {
            if err := json.NewDecoder(strings.NewReader(Message)).Decode(&M); err == nil {
                for k, v := range M {
                    switch vv := v.(type) {
                    case float64:
                        Port, _ := strconv.Atoi(k)
                        USock.SaveToDB(&manager.Flow{Port: int32(Port), Size: vv, Created: time.Now().Format("2006-01-02 15:04:05"), Modified: time.Now().Format("2006-01-02 15:04:05")})
                    default:
                        fmt.Printf("undefined message type: %T => %T", k, v)
                    }
                }
            }
        }
    })

    select {}
}
