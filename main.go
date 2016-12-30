package main

import (
    "fmt"
    "time"
    "strconv"
    "strings"
    "encoding/json"
    "shadowsocks-manager/manager"
    "gopkg.in/mgo.v2/bson"
    "flag"
    "os"
)

type Limit struct {
    AllowSize float64
    Password  string
}

type Options struct {
    DBHost             string
    DBName             string
    DBUsername         string
    DBPassword         string
    HeartbeatFrequency int
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

func Header() {
    //fmt.Printf("[%s] +%s\r\n", time.Now().Format("2006-01-02 15:04:05"), "welcome to use ss-manager ^_^____")
    //fmt.Printf("    author: %s\r\n", "shuc324@gmail.com")
    //fmt.Printf("    time: %s\r\n", "2016-12-30 00:00:00")
    //fmt.Printf("    version: %s\r\n", "1.0")

    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s -h | [-f FREQUENCY] [-h DB_HOST] [-d DB_NAME] [-u USERNAME] [-p PASSWORD]\n\n" +
            "Welcome to use ss-manager ^_^____\r\n\r\n" +
            "Options:\n", os.Args[0])
        flag.PrintDefaults()
    }

    flag.String("name", "gerry", "input ur name")
    flag.String("age", "30", "input ur age")

    flag.Parse()
    return &Options{

    }
}

func main() {
    go Header()

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

    // 正在监听的端口
    ListenPorts := manager.New()

    USock.Listen()
    go USock.Ping()

    // 每30sec检查流量是否超标
    go USock.HeartBeat(HEARTBEAT_FREQUENCY, func() error {
        Ports := manager.New()
        Users := []manager.User{}
        Limits := make(map[int32]Limit)

        fmt.Printf("[%s] +auto update %dsec\r\n", time.Now().Format("2006-01-02 15:04:05"), HEARTBEAT_FREQUENCY)

        if USock.Con.C(USER_COLLECTION).Find(bson.M{"status": true}).All(&Users) == nil {
            for _, User := range Users {
                if User.Port != 0 {
                    Ports.Add(User.Port)
                    Limits[User.Port] = Limit{
                        Password: User.Password,
                        AllowSize: User.AllowSize,
                    }
                }
            }
        }

        for _, Port := range manager.Minus(ListenPorts, Ports).List() {
            USock.Del(Port)
        }

        if !Ports.Empty() {
            StartTime, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
            Pipe := USock.Con.C(FLOW_COLLECTION).Pipe([]bson.M{
                {
                    "$match": bson.M{
                        "port": bson.M{"$in": Ports.List()},
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

            for _, Item := range Resp {
                Port := int32(Item["_id"].(int))
                AllowSize := Item["total"].(float64)
                if _, ok := Limits[Port]; !ok {
                    _, err := USock.Del(Port)
                    if err == nil {
                        ListenPorts.Remove(Port)
                        fmt.Printf("    -del: %d\r\n", Port)
                    }
                } else {
                    if Limits[Port].AllowSize != float64(0) && Limits[Port].AllowSize < AllowSize {
                        _, err := USock.Del(Port)
                        if err == nil {
                            ListenPorts.Remove(Port)
                            fmt.Printf("    -del: %d\r\n", Port)
                            delete(Limits, Port)
                        }
                    }
                }
            }

            for Port, item := range Limits {
                if !ListenPorts.Has(Port) {
                    _, err := USock.Add(Port, string(item.Password))
                    if err == nil {
                        ListenPorts.Add(Port)
                        fmt.Printf("    +add: %d\r\n", Port)
                    }
                } else {
                    fmt.Printf("    *lis: %d\r\n", Port)
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
