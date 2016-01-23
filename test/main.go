package main

import (
	"flag"
	"github.com/xlvector/socks"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func genIpsFromFile(fn string, ips chan string) {
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Println(err)
	}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		ips <- strings.TrimSpace(line)
	}
}

func genIps(prefix string, ips chan string) {
	tks := strings.Split(prefix, ".")
	if len(tks) == 3 {
		for i := 1; i < 256; i++ {
			buf := prefix + "." + strconv.Itoa(i)
			ips <- buf
		}
	}

	if len(tks) == 2 {
		for i := 1; i < 256; i++ {
			buf := prefix + "." + strconv.Itoa(i)
			for j := 1; j < 256; j++ {
				buf2 := buf + "." + strconv.Itoa(j)
				ips <- buf2
			}
		}
	}

	if len(tks) == 1 {
		for i := 1; i < 256; i++ {
			buf := prefix + "." + strconv.Itoa(i)
			for j := 1; j < 256; j++ {
				buf2 := buf + "." + strconv.Itoa(j)
				for k := 1; k < 256; k++ {
					buf3 := buf2 + "." + strconv.Itoa(k)
					ips <- buf3
				}
			}
		}
	}
}

func socksClient(ip string, tp int) *http.Client {
	dialSocksProxy := socks.DialSocksProxy(tp, ip, time.Second*2)
	tr := &http.Transport{Dial: dialSocksProxy}
	return &http.Client{Transport: tr}
}

func tryConnect(ip string, tp int) bool {
	if rand.Intn(100) == 0 {
		log.Println(ip)
	}
	client := socksClient(ip, tp)

	link := "http://121.201.28.254:8999/?ip=" + ip
	if tp == socks.SOCKS4 {
		link += "&type=socks4"
	} else if tp == socks.SOCKS5 {
		link += "&type=socks5"
	}
	_, err := client.Get(link)
	if err != nil {
		return false
	}
	return true
}

func tryHttpProxy(ip string) bool {
	proxy, _ := url.Parse("http://" + ip)
	transport := &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			timeout := time.Duration(5) * time.Second
			deadline := time.Now().Add(timeout)
			c, err := net.DialTimeout(netw, addr, timeout)
			if err != nil {
				return nil, err
			}
			c.SetDeadline(deadline)
			return c, nil
		},
		Proxy: http.ProxyURL(proxy),
	}

	client := &http.Client{Transport: transport}
	link := "http://121.201.28.254:8999/?ip=" + ip + "&type=http"
	_, err := client.Get(link)
	if err != nil {
		return false
	}
	return true
}

func main() {
	prefix := flag.String("p", "", "prefix")
	fn := flag.String("f", "", "file name")
	n := flag.Int("n", 1000, "num")
	port := flag.String("t", "", "port")
	tag := flag.String("tag", "socks5", "tag")
	flag.Parse()

	ips := make(chan string, 100000)

	if len(*prefix) > 0 {
		go genIps(*prefix, ips)
	}

	if len(*fn) > 0 {
		go genIpsFromFile(*fn, ips)
	}

	for k := 0; k < *n; k++ {
		go func() {
			for ip := range ips {
				if len(*port) > 0 {
					if *tag == "socks5" && tryConnect(ip+":"+*port, socks.SOCKS5) {
						log.Println("okkkkkkkkkkk 5")
					}

					if *tag == "http" && tryHttpProxy(ip+":"+*port) {
						log.Println("okkkkkkkkkkk http")
					}
				} else {
					if *tag == "socks5" && tryConnect(ip, socks.SOCKS5) {
						log.Println("okkkkkkkkkkk 5")
					}

					if *tag == "http" && tryHttpProxy(ip) {
						log.Println("okkkkkkkkkkk http")
					}
				}
			}
		}()
	}

	t := time.NewTicker(time.Second)
	for t := range t.C {
		log.Println(t, len(ips))
		if len(ips) == 0 {
			time.Sleep(time.Second * 5)
			break
		}
	}
}
