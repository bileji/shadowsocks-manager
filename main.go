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

var (
    MONGODB_HOST = "127.0.0.1:27017"
    MONGODB_DATABASE = "vpn"
    MONGODB_USERNAME = "shadowsocks"
    MONGODB_PASSWORD = "mlgR4evB"

    USER_COLLECTION = "users"
    FLOW_COLLECTION = "flows"

    HEARTBEAT_FREQUENCY = 30
)

func main() {
    err, Con := manager.ConnectToMgo(MONGODB_HOST, MONGODB_DATABASE, MONGODB_USERNAME, MONGODB_PASSWORD)

    if err != nil {
        panic(err)
    }

    defer Con.Session.Close()

    USock := manager.UnixSock{
        Net: "unixgram",
        LSock: "/var/run/manager.sock",
        RSock: "/var/run/shadowsocks-manager.sock",
        Con: Con,
        Collection: FLOW_COLLECTION,
    }

    USock.Listen()
    go USock.Ping()

    // 每30sec检查流量是否超标
    go USock.HeartBeat(HEARTBEAT_FREQUENCY, func() error {
        Ports := []int32{}
        Users := []manager.User{}
        Limits := make(map[int32]Limit)

        fmt.Printf("[%s] +auto update %dsec\r\n", time.Now().Format("2006-01-02 15:04:05"), HEARTBEAT_FREQUENCY)

        if USock.Con.C(USER_COLLECTION).Find(bson.M{"status": true}).All(&Users) == nil {
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
            Pipe := USock.Con.C(FLOW_COLLECTION).Pipe([]bson.M{
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
        if Message := strings.TrimLeft(string(buffer), "stat: "); strings.Compare(Message, "pong") > 0 {
            fmt.Printf("[%s] +shadowsocks-manager\r\n", time.Now().Format("2006-01-02 15:04:05"))
        } else {
            if err := json.NewDecoder(strings.NewReader(Message)).Decode(&M); err == nil {
                for k, v := range M {
                    switch vv := v.(type) {
                    case float64:
                        Port, _ := strconv.Atoi(k)
                        USock.SaveToDB(&manager.Flow{Port: int32(Port), Size: vv, Created: time.Now().Format("2006-01-02 15:04:05"), Modified: time.Now().Format("2006-01-02 15:04:05")})
                    default:
                        fmt.Printf("undefined message type: %T => %T\r\n", k, v)
                    }
                }
            }
        }
    })

    select {}
}
