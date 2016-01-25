package main

import (
	"errors"
	"flag"
	"github.com/xlvector/socks"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func socksClient(ip string) *http.Client {
	dialSocksProxy := socks.DialSocksProxy(socks.SOCKS5, ip, time.Second*10)
	tr := &http.Transport{Dial: dialSocksProxy, ResponseHeaderTimeout: time.Second * 10}
	return &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			log.Println("redirect to:%s", req.URL.String())
			return errors.New("does not allow redirect")
		},
	}
}

func loadLines(fn string, c chan string) {
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Println(err)
		return
	}
	ret := strings.Split(string(b), "\n")
	for _, line := range ret {
		c <- strings.TrimSpace(line)
	}
}

func download(ip, link string) []byte {
	log.Println("begin download", link, ip)
	c := socksClient(ip)
	resp, err := c.Get(link)
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

func name(link string) string {
	tks := strings.Split(link, "/")
	return tks[len(tks)-1]
}

func main() {
	ps := flag.String("p", "", "proxy file name")
	ls := flag.String("l", "", "links")
	folder := flag.String("d", "./", "folder")
	n := flag.Int("n", 10, "n")
	flag.Parse()

	proxies := make(chan string, 10000)
	links := make(chan string, 10000)

	go loadLines(*ps, proxies)
	go loadLines(*ls, links)

	for i := 0; i < *n; i++ {
		go func() {
			for link := range links {
				p := <-proxies
				kv := strings.Split(p, "\t")
				if len(kv) != 2 {
					continue
				}
				b := download(kv[0]+":"+kv[1], link)
				if b != nil {
					log.Println("success download", link, kv[0]+":"+kv[1])
					log.Println("save to", *folder+"/"+name(link))
					err := ioutil.WriteFile(*folder+"/"+name(link), b, 0655)
					if err != nil {
						log.Fatalln(err)
					}
				} else {
					log.Println("fail download", link, kv[0]+":"+kv[1])
				}
				log.Print(len(proxies), len(links))
			}
		}()
	}

	tc := time.NewTicker(time.Second * 10)
	for t := range tc.C {
		if len(proxies) == 0 || len(links) == 0 {
			break
		}
		log.Println(t, len(proxies), len(links))
	}
}
