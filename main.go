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

type Limit struct {
    AllowSize float64
    Password  string
}

func main() {
    err, Con := manager.ConnectToMgo("127.0.0.1:27017", "vpn", "shadowsocks", "mlgR4evB")

    if err != nil {
        panic(err)
    }

    Con.C("users").Insert(manager.User{
        Username: "邓羽浩",
        Port: 8390,
        Status: true,
        Password:"5dae3cdc",
        AllowSize: 1000000,
        Created: time.Now().Format("2006-01-02 15:04:05"),
        Modified: time.Now().Format("2006-01-02 15:04:05"),
    })

    defer Con.Session.Close()

    USock := manager.UnixSock{
        Net: "unixgram",
        LSock: "/var/run/manager.sock",
        RSock: "/var/run/shadowsocks-manager.sock",
        Con: Con,
        Collection: "flows",
    }

    USock.Listen()
    go USock.Ping()

    // todo每1分钟检查流量是否超标
    go USock.HeartBeat(30, func() error {
        Ports := []int32{}
        Users := []manager.User{}
        Limits := make(map[int32]Limit)

        fmt.Printf("[%s] +auto update %dsec\r\n", time.Now().Format("2006-01-02 15:04:05"), 30)

        if USock.Con.C("users").Find(bson.M{"status": true}).All(&Users) == nil {
            for _, User := range Users {
                if User.Port != 0 {
                    Ports = append(Ports, User.Port)
                    Limits[User.Port] = Limit{
                        Password: User.Password,
                        AllowSize: User.AllowSize,
                    }
                }
            }
        }

        if len(Ports) > 0 {
            StartTime, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
            Pipe := USock.Con.C("flows").Pipe([]bson.M{
                {
                    "$match": bson.M{
                        "port": bson.M{"$in": Ports},
                        "created": bson.M{"$gt": StartTime.Format("2006-01-02 15:04:05")},
                    },
                },
                {
                    "$group": bson.M{"_id": "$port", "total": bson.M{"$sum": "$size"}},
                },
            })

            Resp := []bson.M{}
            if Pipe.All(&Resp) != nil {
                // todo print err info
            }

            for _, item := range Resp {
                Port := int32(item["_id"].(int))
                AllowSize := item["total"].(float64)
                if _, ok := Limits[Port]; !ok {
                    _, err := USock.Del(Port)
                    if err == nil {
                        fmt.Printf("    -del: %d\r\n", Port)
                    }
                } else {
                    if Limits[Port].AllowSize != float64(0) && Limits[Port].AllowSize < AllowSize {
                        _, err := USock.Del(Port)
                        if err == nil {
                            fmt.Printf("    -del: %d\r\n", Port)
                            delete(Limits, Port)
                        }
                    }
                }
            }

            for port, item := range Limits {
                _, err := USock.Add(port, string(item.Password))
                if err == nil {
                    fmt.Printf("    +add: %d\r\n", port)
                }
            }
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
