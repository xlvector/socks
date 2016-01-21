package main

import (
	"flag"
	"fmt"
	"github.com/xlvector/socks"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func socksClient(ip string) *http.Client {
	dialSocksProxy := socks.DialSocksProxy(socks.SOCKS5, ip, time.Second*10)
	tr := &http.Transport{Dial: dialSocksProxy, ResponseHeaderTimeout: time.Second * 10}
	return &http.Client{Transport: tr}
}

func httpProxyClient(ip string) *http.Client {
	proxy, _ := url.Parse("http://" + ip)
	transport := &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			timeout := time.Duration(10) * time.Second
			deadline := time.Now().Add(timeout)
			c, err := net.DialTimeout(netw, addr, timeout)
			if err != nil {
				return nil, err
			}
			c.SetDeadline(deadline)
			return c, nil
		},
		Proxy: http.ProxyURL(proxy),
		ResponseHeaderTimeout: time.Second * 10,
	}

	client := &http.Client{Transport: transport}
	return client
}

func check(ip, tp, target, buf string) bool {
	var c *http.Client
	if tp == "socks5" {
		c = socksClient(ip)
	} else if tp == "http" {
		c = httpProxyClient(ip)
	} else {
		return false
	}

	resp, err := c.Get(target)
	if err != nil || resp == nil || resp.Body == nil {
		return false
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	//log.Println(string(b))
	return strings.Contains(string(b), buf)
}

func main() {
	fn := flag.String("f", "", "file name")
	target := flag.String("t", "baidu", "target")
	flag.Parse()

	b, err := ioutil.ReadFile(*fn)
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(b), "\n")
	tasks := make(chan string, 100000)
	for _, line := range lines {
		tasks <- line
	}

	for i := 0; i < 100; i++ {
		go func() {
			line := <-tasks
			tks := strings.Split(strings.TrimSpace(line), "\t")
			log.Println(tks)

			if *target == "taobao" && check(tks[0]+":"+tks[1], tks[2], "https://www.taobao.com/", "<title>淘宝网 - 淘！我喜欢</title>") {
				fmt.Println(">>", tks[0], tks[1], "https://www.taobao.com/")
			}
			if *target == "baidu" && check(tks[0]+":"+tks[1], tks[2], "http://www.baidu.com/", "<title>百度一下，你就知道</title>") {
				fmt.Println(">>", tks[0], tks[1], "http://www.baidu.com/")
			}

		}()
	}

	tc := time.NewTicker(time.Second * 10)
	for t := range tc.C {
		if len(tasks) == 0 {
			break
		}
		log.Println(t, len(tasks))
	}
	time.Sleep(120)
}
