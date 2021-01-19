package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"crypto/tls"
	"encoding/json"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/tidwall/gjson"
)

var (
	EndpointsApi = "https://xxx/api/v1/namespaces/aimd/endpoints/kafka-debezium-proxy"
	kafkaApi = "http://kafka-debezium-proxy.aimd.svc/connectors"
	PutConnApi = "http://kafka-debezium-proxy.aimd.svc/connectors/mys208/config"
	connapi = "http://kafka-debezium-proxy.aimd.svc/connectors/mys208"
	statusapi = "http://kafka-debezium-proxy.aimd.svc/connectors/mys208/status"
	contType = "application/json"
	pool *redis.Pool
)

// kafka conn put接口的结构体
type Putconndata struct {
	Connectorclass  string   `form:"connector.class" json:"connector.class" `
	Incrementingcolumnname string `form:"incrementing.column.name" json:"incrementing.column.name"`
	Databaseuser string `form:"database.user" json:"database.user"`
	Transformschangetopictype string `form:"transforms.changetopic.type" json:"transforms.changetopic.type"`
	Databaseserverid string `form:"database.server.id" json:"database.server.id"`
	Transformschangetopicreplacement string `form:"transforms.changetopic.replacement" json:"transforms.changetopic.replacement"`
	Databasehistorykafkabootstrapservers string `form:"database.history.kafka.bootstrap.servers" json:"database.history.kafka.bootstrap.servers"`
	Databasehistorykafkatopic string `form:"database.history.kafka.topic" json:"database.history.kafka.topic"`
	Transformschangetopicregex string `form:"transforms.changetopic.regex" json:"transforms.changetopic.regex"`
	Transforms string `form:"transforms" json:"transforms"`
	Databaseservername string `form:"database.server.name" json:"database.server.name"`
	Databaseport string `form:"database.port" json:"database.port"`
	Includeschemachanges string `form:"include.schema.changes" json:"include.schema.changes"`
	Tablewhitelist string `form:"table.whitelist" json:"table.whitelist"`
	Databasehostname string `form:"database.hostname" json:"database.hostname"`
	Databasepassword string `form:"database.password" json:"database.password"`
	Name string `form:"name" json:"name"`
	Databasehistoryskipunparseableddl string `form:"database.history.skip.unparseable.ddl" json:"database.history.skip.unparseable.ddl"`
	Transformsunwraptype string `form:"transforms.unwrap.type" json:"transforms.unwrap.type"`
	Databasewhitelist string `form:"database.whitelist" json:"database.whitelist"`
	Snapshotmode string `form:"snapshot.mode" json:"snapshot.mode"`
}

// kafka conn post接口的结构体
type Postconndata struct {
	Name string `form:"name" json:"name"`
	Config Putconndata `form:"config" json:"config"`
}

// 初始化put接口结构体
func NewPutconndata(ip, tablist string) *Putconndata {
	return &Putconndata{
		Connectorclass: "io.debezium.connector.mysql.MySqlConnector",
	    Incrementingcolumnname: "id",
	    Databaseuser: "debezium",
        Transformschangetopictype: "org.apache.kafka.connect.transforms.RegexRouter",
        Databaseserverid: "1",
        Transformschangetopicreplacement: "$1-smt",
        Databasehistorykafkabootstrapservers: ip,
        Databasehistorykafkatopic: "account_topic",
        Transformschangetopicregex: "(.*)",
        Transforms: "unwrap,changetopic",
        Databaseservername: "test",
        Databaseport: "4249",
        Includeschemachanges: "true",
        Tablewhitelist: tablist,
        Databasehostname: "xxx",
        Databasepassword: "xxx",
        Name: "mys208",
        Databasehistoryskipunparseableddl: "true",
        Transformsunwraptype: "io.debezium.transforms.UnwrapFromEnvelope",
        Databasewhitelist: "muniuser",
		Snapshotmode: "schema_only",
	}
}

// 初始化post接口结构体
func NewPostconndata(p Putconndata) *Postconndata {
	return &Postconndata{
		Name: "mys208",
		Config: p,
	}
}

// 自定义字符串切割分隔符
func SplitByMoreStr(r rune) bool {
	return r == '[' || r == ']' || r == ',' || r == '"'
}

// 初始化redis连接池
func initRedisClinet() (p *redis.Pool) {
	return &redis.Pool {
		MaxIdle:     1,  //最初的连接数量
		MaxActive:   0,  //连接池最大连接数量,不确定可以用0（0表示自动定义），按需分配
		IdleTimeout: 120, //连接关闭时间 300秒 （300秒不使用自动关闭）
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

// 获取redis数据
func GetRedisData(p *redis.Pool, keyname string) string {
	data, err := redis.String(p.Get().Do("GET", keyname))
	if err != nil {
		fmt.Printf("get redis failed, err:%v\n", err)
		return "no"
	}
	return data
}

// 获取kafka集群ip
func GetEndpointsApi() (v, strerr string) {
	defer GoRecover()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(EndpointsApi)
	if err != nil {
		fmt.Printf("get Endpoints failed, err:%v\n", err)
		return "", "no"
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("get Endpoints data failed, err:%v\n", err)
		return "", "no"
	}
	data := gjson.Get(string(b), "subsets.#.addresses.#.ip").String()
	if data == "" {
		fmt.Println("gjson解析endpoint ip失败")
		return "", "no"
	}
	sdata := strings.FieldsFunc(data, SplitByMoreStr)
	strdata := strings.Join(sdata, ",")
	return strdata, ""
}

// 获取kafka监听信息
func GetKafkaConnData() (connip, conntablelist, strerr string) {
	defer GoRecover()
	resp, err := http.Get(connapi)
	if err != nil {
		return "", "", "no"
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
    connip = gjson.Get(string(b), "config.database\\.history\\.kafka\\.bootstrap\\.servers").String()
	conntablelist = gjson.Get(string(b), "config.table\\.whitelist").String()
	if connip == "" || conntablelist == "" {
		return "", "", "no"
	}
	strerr = "ok"
	return
}

// 配置kafka监听
func Setkafkaconn(p *redis.Pool, strip string) bool {
	listip := strings.Split(strip, ",")
	clusterip := fmt.Sprintf("%s:9092,%s:9092,%s:9092", listip[0], listip[1], listip[2])
	fmt.Printf("kafka cluster ip is %s\n", clusterip)
	tablist := GetRedisData(p, "kafkatablist")
	if tablist == "no" {
		fmt.Println("获取redis kafkatablist失败")
		return false
	}
	putconndata := NewPutconndata(clusterip, tablist)
	data := NewPostconndata(*putconndata)
	JsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("序列化json失败,配置kafka监听失败, err:%v\n", err)
		return false
	}
	fmt.Printf("JSON序列化后的格式:%s\n", JsonData)
	defer GoRecover()
	response, err := http.Post(kafkaApi, contType, strings.NewReader(string(JsonData)))
	if err != nil {
		fmt.Printf("请求kafka监听接口失败,err:%v\n", err)
		return false
	}
	defer response.Body.Close()
	resp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("get response data failed, err:%v\n", err)
		return false
	}
	fmt.Printf("配置kafka监听成功response:%s\n", string(resp))
	return true
}

// 更新kafka监听
func Putkafkaconn(p *redis.Pool, strip string) bool {
	listip := strings.Split(strip, ",")
	clusterip := fmt.Sprintf("%s:9092,%s:9092,%s:9092", listip[0], listip[1], listip[2])
	tablist := GetRedisData(p, "kafkatablist")
	if tablist == "no" {
		fmt.Println("获取redis kafkatablist失败")
		return false
	}
	data := NewPutconndata(clusterip, tablist)
	defer GoRecover()
	JsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("序列化json失败,更新kafka监听失败, err:%v\n", err)
		return false
	}
	fmt.Printf("JSON序列化后的格式:%s\n", JsonData)
	res, err := http.NewRequest("PUT", PutConnApi, strings.NewReader(string(JsonData)))
	if err != nil {
		fmt.Printf("更新kafka监听失败, err:%v\n", err)
		return false
	}
	res.Header.Add("content-type", contType)
	response, _ := http.DefaultClient.Do(res)
	defer response.Body.Close()
	resp, _ := ioutil.ReadAll(response.Body)
	fmt.Printf("更新kafka监听成功response:%s\n", string(resp))
	return true
}

// 删除kafka监听
func Delkafkaconn() bool {
	defer GoRecover()
	res, err := http.NewRequest("DELETE", connapi, nil)
	if err != nil {
		fmt.Printf("删除kafka监听失败, err:%v\n", err)
		return false
	}
	response, _ := http.DefaultClient.Do(res)
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)
	fmt.Printf("删除kafka监听成功response:%s\n", string(data))
	return true
}

// 检测kafka监听状态
func CheckStatus(p *redis.Pool, ip string ) bool {
	defer GoRecover()
	resp, err := http.Get(statusapi)
	// 捕获异常操作
	if err != nil || resp.StatusCode == 404 {
		fmt.Println("kafka集群重建或监听被删除,正在重新配置kafka监听...")
		// 配置监听
		if ok := Setkafkaconn(p, ip);!ok {
			fmt.Println("配置kafka监听失败")
			return false
		}
		return true
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	ConnStatus := gjson.Get(string(b), "connector.state").String()
    TasksStatus := gjson.Get(string(b), "tasks.0.state").String()
	if ConnStatus != "RUNNING" || TasksStatus != "RUNNING" {
		fmt.Println("kafka集群ip发送飘逸,正在重新配置kafka监听ip...")
		// 删除监听
	    if ok := Delkafkaconn();!ok {
	    	return false
	    }
	    // 配置监听
	    if ok := Setkafkaconn(p, ip);!ok {
	    	fmt.Println("配置kafka监听失败")
	    	return false
	    }
	}
	return true
}

// 检测kafka监听是否正常及是否有数据库表更新操作
func CheckData(p *redis.Pool) bool {
	strip, strerr := GetEndpointsApi()
	if strerr == "no" {
		fmt.Println("调用Endpoints接口失败!")
		return false
	}
	if len(strings.Split(strip, ",")) != 3 {
		fmt.Println("kafka集群节点数不等于3,需等待kafka集群恢复...")
		return false
	}
	fmt.Printf("当前最新的kafka集群节点ip:%s\n", strip)
	if ok := CheckStatus(p, strip);!ok {
		return false
	}
	connip, table, ok := GetKafkaConnData()
	if ok == "no" {
		return false
    }
	fmt.Printf("kafka监听配置ip是:%s\n", connip)
	fmt.Println("kafka监听正常,正在检测kafka监听表是否有更新操作...")
	// 动态感知redis kafkatablist 发送变化，执行PUT操作更新监听配置
	newtable := GetRedisData(p, "kafkatablist")
	if newtable == "no" {
		fmt.Println("获取redis kafkatablist失败")
		return false
	}
	sufstr := strings.Split(newtable, ",")
	if strings.Contains(table, sufstr[len(sufstr)-1]) {
		if len(sufstr) != len(strings.Split(table, ",")) {
			// 更新kafka监听配置
	        if ok := Putkafkaconn(p, strip);!ok {
				fmt.Println("更新kafka监听配置操作失败!")
			    return false
	        }
		    fmt.Println("kafka监听mysql表更新成功!")
		    return true
		}
		fmt.Println("没有更新操作!")
		return true	
	}
	// 更新kafka监听配置
	if ok := Putkafkaconn(p, strip);!ok {
		fmt.Println("更新kafka监听配置操作失败!")
		return false
	}
	fmt.Println("kafka监听mysql表更新成功!")
	return true
}

// 捕获异常
func GoRecover() {
	if err := recover(); err != nil {
		return
	}
}

func main()  {
	pool := initRedisClinet()
	defer pool.Close()
	ticker := time.Tick(time.Second * 180)
	for range ticker {
		now := time.Now()
		if ok := CheckData(pool);!ok {
			fmt.Printf("%s:执行失败,进行下一个循环检测\n", now.Format("2006-01-02-15-04"))
		} else {
			fmt.Printf("%s:执行成功\n", now.Format("2006-01-02-15-04"))
		}
	}
}
