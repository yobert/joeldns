UPDATE
------
This tool is useless garbage. You should use CoreDNS which has plugins to do exactly this, and it will actually be reliable.

# joeldns
The way I think a DNS server should be

I hate DNS server settings. I think they're stupid.
What I want in a DNS server:

`/etc/hosts` is the input file for host -> ip lookups

`/etc/resolv.conf` is where you send requests not in /etc/hosts

And that's it!  To build:
```
go get github.com/yobert/joeldns
cd ${GOPATH}/src/github.com/yobert/joeldns
go get ./...
go build .
```

To run:
```
sudo ./joeldns
```
or
```
./joeldns :1234
```
