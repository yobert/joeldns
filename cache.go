package main

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
	"time"
)

type entry struct {
	expire_at int64
	message   dns.Msg
}

func (s *server) expire() {
	expired := 0
	count := 0

	now := time.Now().UTC().Unix()

	s.cache_mu.Lock()
	defer s.cache_mu.Unlock()

	for k, v := range s.cache {
		count++

		if now > v.expire_at {
			delete(s.cache, k)
			expired++
		}
	}

	if expired != 0 {
		log.Printf("expired %d/%d entries\n", expired, count)
	}
}

func (s *server) cache_key(r *dns.Msg) string {
	key := ""
	for _, q := range r.Question {
		key += fmt.Sprintf("%d_%d_%s ", q.Qtype, q.Qclass, q.Name)
	}
	return key
}

func (s *server) cache_set(req *dns.Msg, r *dns.Msg) {
	key := s.cache_key(req)
	if key == "" {
		return
	}

	s.cache_mu.Lock()
	defer s.cache_mu.Unlock()

	expire := int64(144000)
	for _, rr := range r.Answer {
		ttl := int64(rr.Header().Ttl)
		if ttl < expire {
			expire = ttl
		}
	}

	//log.Printf("cache_set %#v for %d seconds\n", key, expire)
	s.cache[key] = &entry{expire_at: time.Now().UTC().Unix() + expire, message: *r}
}

func (s *server) cache_get(req *dns.Msg) *dns.Msg {
	key := s.cache_key(req)
	if key == "" {
		return nil
	}

	s.cache_mu.RLock()
	defer s.cache_mu.RUnlock()

	if entry, ok := s.cache[key]; ok && time.Now().UTC().Unix() < entry.expire_at {
		message := entry.message
		message.Id = req.Id
		//log.Printf("cache_get %#v hit\n", key)
		return &message
	}
	//log.Printf("cache_get %#v missed\n", key)
	return nil
}
