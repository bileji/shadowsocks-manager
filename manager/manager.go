package manager

import (
    "os"
    "net"
    "strconv"
    "gopkg.in/mgo.v2"
    "github.com/noaway/heartbeat"
    "gopkg.in/mgo.v2/bson"
    "fmt"
    "time"
    "strings"
    "encoding/json"
)

type Options struct {
    DBHost             string
    DBName             string
    DBUsername         string
    DBPassword         string
    HeartbeatFrequency int
}

type User struct {
    Username  string
    Port      int32
    Status    bool
    Password  string
    AllowSize float64
    Created   string
    Modified  string
}

type Limit struct {
    AllowSize float64
    Password  string
}

type Flow struct {
    Port     int32
    Size     float64
    Created  string
    Modified string
}

type UnixSock struct {
    Net         string
    LSock       string
    RSock       string
    UConn       *net.UnixConn
    Con         *mgo.Database
    FlowC       string
    UserC       string
    Args        *Options
    ListenPorts *Ports
}

func ConnectToMgo(host string, db string, username string, password string) *mgo.Database {
    session, err := mgo.Dial(host)
    if err != nil {
        panic(err)
    }
    if session.DB(db).Login(username, password) != nil {
        panic(err)
    }
    return session.DB(db)
}

func (us *UnixSock) Listen() {
    os.Remove(us.LSock)
    rAddr, err := net.ResolveUnixAddr(us.Net, us.RSock)
    if err != nil {
        panic(err)
    }
    lAddr, err := net.ResolveUnixAddr(us.Net, us.LSock)
    if err != nil {
        panic(err)
    }
    us.UConn, err = net.DialUnix(us.Net, lAddr, rAddr)
    if err != nil {
        panic(err)
    }
}

func (us *UnixSock) Ping() (int int, err error) {
    int, err = us.UConn.Write([]byte("ping"))
    return
}

func (us *UnixSock) Add(port int32, pwd string) (int int, err error) {
    str := "add: {\"server_port\": " + strconv.FormatInt(int64(port), 10) + ", \"password\": \"" + pwd + "\"}"
    int, err = us.UConn.Write([]byte(str))
    return
}

func (us *UnixSock) Del(port int32) (int int, err error) {
    str := "remove: {\"server_port\": " + strconv.FormatInt(int64(port), 10) + "}"
    int, err = us.UConn.Write([]byte(str))
    return
}

func (us *UnixSock) Rec(fn func(res []byte)) {
    for {
        buffer := make([]byte, 128)
        _, err := us.UConn.Read(buffer)
        if err == nil {
            fn(buffer)
        }
    }
}

func (us *UnixSock) HeartBeat(spec int, fn func() error) error {
    ht, err := heartbeat.NewTast("task", spec)
    if err != nil {
        return err
    }
    ht.Start(fn)

    return nil
}

func (us *UnixSock) Monitor() error {
    Ports := New()
    Users := []User{}
    Limits := make(map[int32]Limit)

    fmt.Printf("[%s] +auto update %dsec\r\n", time.Now().Format("2006-01-02 15:04:05"), us.Args.HeartbeatFrequency)

    if us.Con.C(us.UserC).Find(bson.M{"status": true}).All(&Users) == nil {
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

    for _, Port := range Minus(us.ListenPorts, Ports).List() {
        us.Del(Port)
    }

    if !Ports.Empty() {
        StartTime, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
        Pipe := us.Con.C(us.FlowC).Pipe([]bson.M{
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
                _, err := us.Del(Port)
                if err == nil {
                    us.ListenPorts.Remove(Port)
                    fmt.Printf("    -del: %d\r\n", Port)
                }
            } else {
                if Limits[Port].AllowSize != float64(0) && Limits[Port].AllowSize < AllowSize {
                    _, err := us.Del(Port)
                    if err == nil {
                        us.ListenPorts.Remove(Port)
                        fmt.Printf("    -del: %d\r\n", Port)
                        delete(Limits, Port)
                    }
                }
            }
        }

        for Port, item := range Limits {
            if !us.ListenPorts.Has(Port) {
                _, err := us.Add(Port, string(item.Password))
                if err == nil {
                    us.ListenPorts.Add(Port)
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
}

func (us *UnixSock) SaveToDB(buffer []byte) {
    M := make(map[string]interface{})
    if Message := strings.TrimLeft(string(buffer), "stat: "); strings.EqualFold(Message, "pong") {

    } else {
        if err := json.NewDecoder(strings.NewReader(Message)).Decode(&M); err == nil {
            fmt.Println(M)
            for k, v := range M {
                switch Size := v.(type) {
                case float64:
                    Port, _ := strconv.Atoi(k)
                    us.Con.C(us.FlowC).Insert(&Flow{
                        Port: int32(Port),
                        Size: Size,
                        Created: time.Now().Format("2006-01-02 15:04:05"),
                        Modified: time.Now().Format("2006-01-02 15:04:05"),
                    })
                default:
                    fmt.Printf("undefined message type: %T => %T\r\n", k, v)
                }
            }
        }
    }
}