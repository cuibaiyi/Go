package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func TickClearLog(logpath string) {
	ticker := time.Tick(time.Second * 5)
	s, _ := filepath.Glob(logpath)
	flen := len(s)
	if flen == 0 {
		log.Print("没有匹配到日志文件")
		return
	}

	for range ticker {
		now := time.Now()
		mtime := now.Format("[2006-01-02 15:04:05]")

		ch := make(chan string, flen)
		for i := range s {
			go func(file string, ch chan string) {
				if err := os.Truncate(file, 0); err != nil {
					Ecologist := fmt.Sprintf("%s %s 日志清空失败\n", mtime, file)
					ch <- Ecologist
				}else {
					Login := fmt.Sprintf("%s %s 日志清理成功\n", mtime, file)
					ch <- Login
				}
			}(s[i], ch)
		}

		f, err := os.OpenFile("./rmlog.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		for range s {
			f.WriteString(<-ch)
		}
	}
}

func main()  {
	var logpath string
	flag.StringVar(&logpath, "f", "", "请填写要清理日志的路径,支持*匹配目录和文件,填写的路径一定要加双引号括起来!")
	flag.Parse()
	if len(logpath) == 0 {
		log.Println("请使用-f参数,指定清理的日志路径!")
		return
	}
	TickClearLog(logpath)
}

