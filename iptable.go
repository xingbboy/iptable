package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gogf/gf/net/gipv4"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"os"
	"regexp"
)

type ip_info struct {
	Region string `json:"region"`
}

type ip_data struct {
	ip4  uint32
	info string
}

var (
	ip_datas []ip_data
	addr     = flag.String("addr", ":8080", "TCP address to listen to")
)

func LoadIPFile(filename string) []ip_data {
	file, err := os.Open(filename)
	if err != nil {
		logrus.WithFields(logrus.Fields{}).Fatal("open ip.txt fail.")
	}
	defer file.Close()

	var datas []ip_data
	scanner := bufio.NewScanner(file)
	r, _ := regexp.Compile("(\\d+.\\d+.\\d+.\\d+)\\s+(\\d+.\\d+.\\d+.\\d+)\\s+(\\S+)")
	for scanner.Scan() {
		var line = scanner.Text()
		var data = r.FindStringSubmatch(line)
		datas = append(datas, ip_data{ip4: gipv4.Ip2long(data[1]), info: data[3]})
	}
	ip_datas = datas
	logrus.WithFields(logrus.Fields{}).Info("open ip.txt fail.")
	return datas
}

func getIPInfo(ip string) string {
	var ip4 = gipv4.Ip2long(ip)
	var info = ip_info{Region: "地球某处"}
	var l = 1
	var r = len(ip_datas)
	var m = (l + r + 1) / 2
	for {
		if l > r {
			break
		}
		if ip_datas[m].ip4 <= ip4 {
			l = m
			if (ip_datas[m].ip4 == ip4) || (ip_datas[m+1].ip4 > ip4) {
				info.Region = ip_datas[m].info
				break
			}
		} else {
			r = m
			if ip_datas[m-1].ip4 <= ip4 {
				info.Region = ip_datas[m-1].info
				break
			}
		}
		m = (l + r + 1) / 2
	}
	data, _ := json.Marshal(info)
	return string(data)
}

func getinfo(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ip := r.Form.Get("ip")
	data := getIPInfo(ip)
	w.Write([]byte(data))
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func getinfo1(ctx *fasthttp.RequestCtx) {
	args := ctx.QueryArgs()

	data := getIPInfo(string(args.Peek("ip")))
	fmt.Fprintf(ctx, data)
	fmt.Fprintf(ctx, "\r\n\r\n")
	ctx.SetConnectionClose()
}

func ping1(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "pong\r\n\r\n")
	ctx.SetConnectionClose()
}

func main() {
	// 设置日志输出到屏幕
	logrus.SetOutput(os.Stdout)
	// 设置日志格式为json格式
	logrus.SetFormatter(&logrus.TextFormatter{})
	// 设置日志级别
	logrus.SetLevel(logrus.InfoLevel)
	LoadIPFile("ip.txt")

	//http.HandleFunc("/getinfo", getinfo)
	//http.HandleFunc("/ping", ping)
	//
	//s := http.Server{
	//	Addr:         ":8080",
	//	Handler:      nil,
	//	ReadTimeout:  500 * time.Millisecond,
	//	WriteTimeout: 500 * time.Millisecond,
	//}
	//
	//logger.Log(logger.InfoLevel, "start listen 8080.")
	//s.ListenAndServe()

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/getinfo":
			getinfo1(ctx)
		default:
		}
	}

	if err := fasthttp.ListenAndServe(*addr, requestHandler); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
