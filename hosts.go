package main

import (
	"os"
	"bufio"
	"strings"
	"fmt"
	"log"
	"net"
	"io"
)

type hostcache struct {
	names map[string][]string
	addrs map[string][]string
}

func loadhosts() (*hostcache, error) {
	fpath := "/etc/hosts"

	fh, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	r := bufio.NewReader(fh)

	hs := make(map[string][]string)
	is := make(map[string][]string)

	linecount := 0
	for {
		linecount++
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if i := strings.IndexByte(line, '#'); i >= 0 {
			// Discard comments.
			line = line[0:i]
		}
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		addr := f[0]
		ip := net.ParseIP(addr)
		if ip == nil {
			return nil, fmt.Errorf("Parse error at %s line %d: %#v is not a valid IP paddress", fpath, linecount, addr)
		}
		if ip.To4() == nil {
			log.Printf("Parse warning at %s line %d: IPv6 not supported yet (%#v)\n", fpath, linecount, addr)
			continue
		}
		for i := 1; i < len(f); i++ {
			name := strings.ToLower(f[i])
			hs[name] = append(hs[name], addr)
			is[addr] = append(is[addr], name)
		}
	}

	log.Println(fpath + " loaded")

	return &hostcache{
		names: hs,
		addrs: is,
	}, nil
}

