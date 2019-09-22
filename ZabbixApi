package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type zabbixobj struct {
	url string
	contType string
	user string
	passwd string
}

func NewObj(u, c, user, passwd string) *zabbixobj {
	return &zabbixobj{
		url: u,
		contType: c,
		user: user,
		passwd: passwd,
	}
}

func main() {
	var zapi *zabbixobj
	zapi = NewObj("http://127.0.0.1/api_jsonrpc.php", "application/json-rpc", "zabbix", "123456")
	CreateHost(zapi)
}

func Login(z *zabbixobj) string {
	data := fmt.Sprintf(`{
		{
			"jsonrpc": "2.0",
			"method": "user.login",
			"params": {
			"user": %s,
				"password": %s
		},
			"id": 1,
			"auth": None
		}
	}`, z.user, z.passwd)

	JsonData, _ := json.Marshal(data)
	resp, err := http.Post(z.url, z.contType, strings.NewReader(string(JsonData)))
	if err != nil {
		fmt.Println("post failed, err:%v\n", err)
		return ""
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("get resp failed,err:%v\n", err)
		return ""
	}
	return string(b)
}

func CreateHost(z *zabbixobj) {
	ipList := []string{"127.0.0.1","192.168.1.1"}
	for _, ip := range ipList {
		data := fmt.Sprintf(`
		{
			"jsonrpc": "2.0",
			"method": "host.create",
			"params": {
			"host": %s,
				"name": %s,
				"interfaces": [{
				"type": 1,
				"main": 1,
				"useip": 1,
				"ip": %s,
				"dns": "",
				"port": "10050"
			}],
		"groups": [{
		"groupid": "groupid"
		}],
		"templates": [{
		"templateid": "10001"
		}],
		"inventory_mode": 0,
		"proxy_hostid": "10293"
		},
		"auth": %s,
		"id": 1
		}`,
		string(ip), string(ip), string(ip), Login(z))
		
		JsonData, _ := json.Marshal(data)
		resp, err := http.Post(z.url, z.contType, strings.NewReader(string(JsonData)))
		if err != nil {
			fmt.Println("post failed, err:%v\n", err)
			return
		}
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("get resp failed,err:%v\n", err)
			return
		}
		fmt.Println(string(b))
	}
}
