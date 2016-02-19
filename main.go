package main

import (
	"github.com/miekg/dns"
	"log"
	"net"
	"os"
	"strings"
)

type server struct {
	client *dns.Client
	config *dns.ClientConfig
	hosts  *hostcache
}

func (s *server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {

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

	upstream_r, _, err := s.client.Exchange(r, net.JoinHostPort(s.config.Servers[0], s.config.Port))
	if err != nil {
		log.Println(err)
		return
	}

	upstream_r.SetReply(r)

	w.WriteMsg(upstream_r)
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

	s := &server{
		config: config,
		client: client,
		hosts:  hosts,
	}

	addr := ":53"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	log.Println("serving dns on udp " + addr)
	log.Fatal(dns.ListenAndServe(addr, "udp", s))
}
