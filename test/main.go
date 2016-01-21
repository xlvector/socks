package main

import (
	"flag"
	"github.com/xlvector/socks"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

func main() {
	prefix := flag.String("p", "", "prefix")
	n := flag.Int("n", 1000, "num")
	port := flag.String("t", "1080", "port")
	flag.Parse()

	ips := make(chan string, 100000)

	go func() {
		genIps(*prefix, ips)
	}()

	for k := 0; k < *n; k++ {
		go func() {
			for ip := range ips {
				if tryConnect(ip+":"+*port, socks.SOCKS5) {
					log.Println("okkkkkkkkkkk 5")
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
