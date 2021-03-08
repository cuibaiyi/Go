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
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"gopkg.in/ini.v1"
)

type Token struct {
	Access_token string `json:"access_token"`
	Token_type   string `json:"token_type"`
	Expires_in   int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type SLogData struct {
	logdata string
	logbody string
	result string
}

var (
	date = now.Format("2006年01月02日15:04")
	// 验证 start
	SendData string
	DataMap map[string]string
	n = "0"
	totol string
	tag uint8
	CountTag uint8
	// end
	datasplice []SLogData
    totolmap map[int]string
	ConfigPath string
	now        = time.Now()
	timestr    = fmt.Sprintf("%s%s", strconv.FormatInt(now.Unix(), 10), "000")
	toTime, fromTime = myt()
	id = "xxx"
	logapi   = "http://xxx/search/xxx"
	gwapi    = "http://xxx/aopoauth/xxx"
	api      = "http://xxx/rest/xxx?"
	contType = "application/json"
	d        = `{
	"otherId":"xxx",
	"smsTitle":"",
	"smsContent":""
	}`
)

func myt() (string, string) {
	toTime := now.Add(time.Minute * -5).Format("2006-01-02 15:04")
	fromTime := now.Add(time.Minute * -60).Format("2006-01-02 15:04")
	return fmt.Sprintf("%s:00", toTime), fmt.Sprintf("%s:00", fromTime)
}

func main() {
	flag.StringVar(&ConfigPath, "f", "./log4xsend.ini", "请填写配置文件路径,路径一定要加双引号括起来!")
	flag.Parse()
	cfg, err := ini.Load(ConfigPath)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		return
	}
	number := cfg.Section("config").Key("numberList").String()
	templateid := cfg.Section("config").Key("templateId").String()
	if demo(number, templateid) {
		fmt.Printf("时间戳:%s FROM:%s TO:%s\n", timestr, fromTime, toTime)
	} else {
		fmt.Printf("%s LOGINMAX值为0或空值,不发生短信!\n", date)
	}
}

func Count() string {
	//成功数
	sdata := fmt.Sprintf(`{"searchParamHead": {"queryTime": %s}, "searchParamBody": {"requestBody": "{\"indexGroup\": \"osp\", \"queryString\": \"* AND success:true AND appNodeRemark:统一待办| stats dc(traceId)\", \"timeType\":2, \"from\":\"%s\", \"to\":\"%s\"}"}}`, timestr, fromTime, toTime)
	sbody := fmt.Sprintf(`{"indexGroup": "osp", "queryString": "* AND success:true AND appNodeRemark:统一待办| stats dc(traceId)", "timeType":2, "from":"%s", "to":"%s"}`, fromTime, toTime)
	obj1 := SLogData{
		logdata: sdata,
		logbody: sbody,
		result: "result.result.0.count(traceId)",
	}

        //失败数
	sdata = fmt.Sprintf(`{"searchParamHead": {"queryTime": %s}, "searchParamBody": {"requestBody": "{\"indexGroup\": \"osp\", \"queryString\": \"* AND success:false AND appNodeRemark:统一待办| stats dc(traceId) as countnum\", \"timeType\":2, \"from\":\"%s\", \"to\":\"%s\"}"}}`, timestr, fromTime, toTime)
	sbody = fmt.Sprintf(`{"indexGroup": "osp", "queryString": "* AND success:false AND appNodeRemark:统一待办| stats dc(traceId) as countnum", "timeType":2, "from":"%s", "to":"%s"}`, fromTime, toTime)
	obj2 := SLogData{
		logdata: sdata,
		logbody: sbody,
		result: "result.result.0.count(traceId)",
	}

	datasplice = append(datasplice, obj1, obj2)
	totolmap = make(map[int]string)
	for i, data := range datasplice {
		totolmap[i] = zoo(data.logdata, data.logbody, data.result)
	}
	zoodata := fmt.Sprintf(`{"DATE":"%s", "OK":"%s", "ERR":"%s"}`, date, totolmap[0], totolmap[1])
	return zoodata
}

func zoo(logdata, logbody, result string) string {
	md5body := strings.ToUpper(md5log(logbody))
	md5str := strings.ToUpper(md5log(fmt.Sprintf("%s%s%s", md5body, timestr, id)))
	keyid := fmt.Sprintf("%s:%s", md5str, "admin")
	res, err := http.NewRequest("POST", logapi, strings.NewReader(logdata))
	if err != nil {
		panic(err)
	}
	res.Header.Add("Content-Type", "application/json")
	res.Header.Add("Log4xAuthorization", keyid)
	fmt.Println(res.Header.Get("Log4xAuthorization"))
	fmt.Println(logdata)
	response, err := http.DefaultClient.Do(res)
	fmt.Println("response", err)
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("get resp failed, err:%v\n", err)
		return "no"
	}
	response.Body.Close()
	fmt.Println(string(data))
	v := gjson.Get(string(data), result)
	return v.String()
}

func md5log(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func validation(tag uint8) string {
	if tag > CountTag {
		return n
	}
	SendData = Count()
	DataMap = make(map[string]string)
	if err := json.Unmarshal([]byte(SendData), &DataMap); err != nil {
		fmt.Println("SendData:json parsing err")
		return "SendData json parsing err"
	}
	totol = DataMap["LOGINMAX"]
	if totol == n || totol == "" {
		tag++
		fmt.Println("totol=0 or 空字符串,restart Count()...")
		time.Sleep(60 * time.Second)
		SendData = validation(tag)
		return SendData
	}
	return SendData
}

func demo(numberstr, templateid string) bool {
	SendData := validation(tag)
	if SendData == n {
		return false
	}
	fmt.Println(SendData)
	d1 := fmt.Sprintf(d, numberstr, templateid, SendData)
	token := GetGwToken(gwapi)
	c := url.Values{"method": {"xxx"}, "format": {"json"}, "appId": {"10009"}, "version": {"Vx"}, "accessToken": {token}, "sign": {"xxx"}, "timestamp": {"20190313162400"},
		"content": {d1}}
	tzapi := fmt.Sprint(api, c.Encode())
	fmt.Println(tzapi)
	code, errstr := GetSendCode(tzapi)
	fmt.Printf("rep返回的状态码%s,错误信息:%s\n", code, errstr)
	return true
}

func GetSendCode(tzapi string) (string, string) {
	resp, err := http.Post(tzapi, contType, nil)
	if err != nil {
		fmt.Printf("post failed, err:%v\n", err)
		return "no", "no"
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("get resp failed, err:%v\n", err)
		return "no", "no"
	}
	fmt.Println(string(b))
	v := gjson.Get(string(b), "result.statusCode").String()
	errstr := gjson.Get(string(b), "result.errReason").String()
	return v, errstr
}

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
