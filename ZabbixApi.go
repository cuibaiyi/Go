package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type zabbixapi interface {
	CreateHost(z *Zabbixobj)
	Login(z *Zabbixobj) string
}

type Zabbixobj struct {
	Url string `json:"url"`
	ContType string `json:"conttype"`
	User string `json:"user"`
	Passwd string `json:"passwd"`
}

func Newzabbixobj(u, c, user, passwd string) *Zabbixobj {
	return &Newzabbixobj{
		Url: u,
		ContType: c,
		User: user,
		Passwd: passwd,
	}
}

func main() {
	var zapi zabbixapi
	zapi = Newzabbixobj("http://127.0.0.1/api_jsonrpc.php", "application/json-rpc", "zabbix", "123456")
	CreateHost(zapi)
}

func Login(z *Zabbixobj) string {
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
	}`, z.User, z.Passwd)

	JsonData, _ := json.Marshal(data)
	resp, err := http.Post(z.Url, z.ContType, strings.NewReader(string(JsonData)))
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

func CreateHost(z *Zabbixobj) {
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
		resp, err := http.Post(z.Url, z.ContType, strings.NewReader(string(JsonData)))
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
