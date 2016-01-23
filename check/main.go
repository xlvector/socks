package main

import (
	"flag"
	"fmt"
	"github.com/xlvector/socks"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
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

func check(ip, tp, target string) []byte {
	var c *http.Client
	if tp == "socks5" {
		c = socksClient(ip)
	} else if tp == "http" {
		c = httpProxyClient(ip)
	} else {
		return nil
	}

	resp, err := c.Get(target)
	if err != nil || resp == nil || resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	return b
}

func main() {
	proxy := flag.String("p", "", "proxy")
	tp := flag.String("a", "", "type")
	target := flag.String("t", "", "target")
	flag.Parse()

	b := check(*proxy, *tp, *target)
	if b != nil {
		fmt.Println(string(b))
	}
}
