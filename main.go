package main

import (
	"github.com/miekg/dns"
	"log"
	"net"
	"os"
	"strings"
)

type server struct {
	hosts  *hostcache
}

func (s *server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {

	m := new(dns.Msg)
	m.SetReply(r)

	for _, q := range r.Question {
		if q.Qtype != dns.TypeA {
			continue
		}

		dom := q.Name
		search := strings.Trim(dom, ".")

		a, ok := s.hosts.names[search]
		if !ok {
			continue
		}

		for _, addr := range a {
			rr := new(dns.A)
			rr.Hdr = dns.RR_Header{Name: dom, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60}
			rr.A = net.ParseIP(addr).To4()

			m.Answer = append(m.Answer, rr)
		}
	}

	w.WriteMsg(m)
	return
}

func main() {
	hosts, err := loadhosts()
	if err != nil {
		log.Fatal(err)
	}

	s := &server{
		hosts:  hosts,
	}

	addr := ":53"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	log.Println("serving dns on udp " + addr)
	log.Fatal(dns.ListenAndServe(addr, "udp", s))
}
