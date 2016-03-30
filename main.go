package main

import (
	"github.com/miekg/dns"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type server struct {
	client *dns.Client
	config *dns.ClientConfig
	hosts  *hostcache

	cache_mu sync.RWMutex
	cache    map[string]*entry
}

func (s *server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {

	if hit := s.cache_get(r); hit != nil {
		w.WriteMsg(hit)
		return
	}

	if r.Question[0].Qtype == dns.TypeA {
		dom := r.Question[0].Name
		search := strings.Trim(dom, ".")

		a, ok := s.hosts.names[search]
		if ok && len(a) > 0 {

			m := new(dns.Msg)
			m.SetReply(r)

			for _, addr := range a {
				rr := new(dns.A)
				rr.Hdr = dns.RR_Header{Name: dom, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60}
				rr.A = net.ParseIP(addr).To4()

				m.Answer = append(m.Answer, rr)
			}

			w.WriteMsg(m)
			return
		}
	}

	first := rand.Intn(len(s.config.Servers))
	i := first

	for {
		upstream_r, _, err := s.client.Exchange(r, net.JoinHostPort(s.config.Servers[i], s.config.Port))
		if err == nil {
			s.cache_set(r, upstream_r)
			w.WriteMsg(upstream_r)
			return
		}

		i++
		if i == len(s.config.Servers) {
			i = 0
		}

		if i == first {
			log.Println(err)
			return
		}
	}
}

func main() {
	hosts, err := loadhosts()
	if err != nil {
		log.Fatal(err)
	}

	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		log.Fatal(err)
	}

	client := new(dns.Client)

	client.ReadTimeout = time.Second
	client.WriteTimeout = time.Second

	s := &server{
		config: config,
		client: client,
		hosts:  hosts,
		cache:  make(map[string]*entry),
	}

	go func() {
		for {
			time.Sleep(time.Second * 10)
			s.expire()
		}
	}()

	addr := ":53"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	log.Println("serving dns on udp " + addr)
	log.Fatal(dns.ListenAndServe(addr, "udp", s))
}
