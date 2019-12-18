package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/ini.v1"
)

// 接收调用网关返回的数据获取token
type Token struct {
	Access_token string `json:"access_token"`
	Token_type   string `json:"token_type"`
	Expires_in   int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type R1 struct {
	TotalRecord int    `json:"totalRecord"`
	OutputExt   string `json:"outputExt"`
	StatusCode  string `json:"statusCode"`
	ErrReason   string `json:"errReason"`
}

// 接收统一通知返回的数据
type Send struct {
	Result     R1     `json:"result"`
	RespDesc   string `json:"respDesc"`
	InstanceId string `json:"instanceId"`
	RespCode   string `json:"respCode"`
}

var (
	ConfigPath string
	SendData   string
	xml        = `
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:sms="http://www.sms.upbms.hp.com">
    <soapenv:Header/>
    <soapenv:Body>
      <sms:smsService>
         <!--Optional:-->
         <target>5400</target>
         <!--Optional:-->
         <action>00</action>
         <!--Optional:-->
         <xmldata><![CDATA[%s]]></xmldata>
         <!--Optional:-->
         <brief>%s</brief>
      </sms:smsService>
    </soapenv:Body>
    </soapenv:Envelope>
	`
	gwapi    = "http://xxx"
	contType = "application/json"
	api      = "http://xxx?"
	d        = `{
	"otherId": "1996092910",
	"smsTitle": "",
	"smsContent":"",
	"ynDefaultNum":"N",
	"sendNum":"",
	"ynTimeSend":"N",
	"timeNum":"20191202155000",
	"ynUseHeibai":"N",
	"heiOrBai":"B",
	"receiveNums":"%s",
	"ynTemplate":"Y",
	"templateNum":"%s",
	"dynaVariablle":%s,
	"ynReply":"N",
	"replyUrl":null,
	"pageSize":10,
	"currentPage":1,
	"provinceCode":"HQ",
	"appCode":"iHR",
	"inputExt":"JSON"
	}`
)

func main() {
	flag.StringVar(&ConfigPath, "f", "/tmp/config.ini", "请填写配置文件路径,路径一定要加双引号括起来!")
	flag.StringVar(&SendData, "d", "", "发送短信的数据,最好加双引号+{}括起来写内容!")
	flag.Parse()
	cfg, err := ini.Load(ConfigPath)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		return
	}
	if len(SendData) == 0 {
		fmt.Println("请使用-d参数指定发送短信的内容,最好加双引号+{}括起来写内容!")
		return
	}
	number := cfg.Section("config").Key("numberList").String()
	templateid := cfg.Section("config").Key("templateId").String()
	mesapi := "http://xxx?"
	d1 := fmt.Sprintf(d, number, templateid, SendData)
	token := GetGwToken(gwapi)
  // 拼接url
	c := url.Values{"method": {"M_ABILITY_UNINOTI_107000001"}, "format": {"json"}, "appId": {"10041"}, "version": {"V1.0"}, "accessToken": {token}, "sign": {"3D993BD22F44E6C1A638D832DB2BFB6C8BA4116BEE1B3561B4B198DB7BB6B1A0"}, "timestamp": {"20190313162400"},
		"content": {d1}}
	tzapi := fmt.Sprint(api, c.Encode())
	fmt.Println(tzapi)
	code := GetSendCode(tzapi)
	fmt.Printf("请求统一通知返回的状态码%s\n", code)
	if code != "00000" {
		directSend(number, mesapi)
	}
}

// 调用短信接口
func directSend(number, mesapi string) {
	s := `<?xml version="1.0" encoding="UTF-8"?><SmsServiceReq><SmsList><FreeCode>000000</FreeCode><Mobile>%s</Mobile><Contents>%s</Contents></SmsList></SmsServiceReq>`
	str := fmt.Sprintf(s, number, SendData)
	// fmt.Println(str)
	md5xml := md5V(fmt.Sprintf("centerhr!12300%s5400", str))
	md5xml = fmt.Sprintf(xml, str, md5xml)
	// fmt.Println(md5xml)
	req, err := http.NewRequest("POST", mesapi, strings.NewReader(md5xml))
	req.Header.Add("content-type", "application/xml;charset=utf-8")
	fmt.Println("request", err)
	response, err := http.DefaultClient.Do(req)
	fmt.Println("response", err)
	data, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	fmt.Println(string(data), err)
}

// 验证md5值
func md5V(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// 调用统一通知接口
func GetSendCode(tzapi string) string {
	resp, err := http.Post(tzapi, contType, nil)
	if err != nil {
		fmt.Printf("post failed, err:%v\n", err)
		return "no"
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("get resp failed, err:%v\n", err)
		return "no"
	}
	objsend := &Send{}
	json.Unmarshal(b, objsend)
	return objsend.RespCode
}

// 获取网关token
func GetGwToken(gwapi string) string {
	r, err := http.Get(gwapi)
	if err != nil {
		fmt.Printf("get failed, err:%v\n", err)
		return "no"
	}
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("get resp failed, err:%v\n", err)
		return "no"
	}
	obj := &Token{}
	json.Unmarshal(b, obj)
	return obj.Access_token
}
