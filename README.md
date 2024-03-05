# What is NostrP2P?
Pure peer-to-peer distributed microblog system on NAT transparent overlay network based on idea of [Nostr](https://en.wikipedia.org/wiki/Nostr)

## Design Note
- [here](https://gist.github.com/ryogrid/fa2bfa284784c866ad88e3c38445752a)
- [Japanese version](https://gist.github.com/ryogrid/0ba0d825c3bb840dffa519c5ab91d4ff)
  - Japanese version is latest :)

## Build
```bash
$ go build -o nostrp2p main.go
```

## NostrP2P command usage
```
Usage:
  nostrp2p [flags]
  nostrp2p [command]

Available Commands:
  help        Help about any command.
  server      Startup server.
  genkey      Generate key pair.

Flags:
  -h, --help   Help for buzzoon
```

```
Usage:
  nostrp2p server [flags]

Flags:
  -h, --help                         Help for server
  -l, --listen-addr-port    string   Address and port to bind to (default: 127.0.0.1:20000)
  -b, --boot-peer-addr-port string   Address and port of a server which already joined buzzoon network (optional)
  -p, --public-key          string   Your public key (required)
  -n, --nickname            string   Your nickname on Buzzoon (required) [TEMPORAL]
  -d, --debug               bool     If true, debug log is output to stderr (default: false)
```

## Examples (Generate key pair)
- Under construction

## Examples (Server lauch)
```bash
# 4 servers network on local network (4 shells are needed...)
./nostrp2p server  -l 0.0.0.0:20000 -p <public key in hex> -n origin
./nostrp2p server -l 0.0.0.0:20002 -p <public key in hex> -n alice -b 127.0.0.1:20000 
./nostrp2p server -l 0.0.0.0:20004 -p <public key in hex> -n bob -b 127.0.0.1:20002
./nostrp2p server -l 0.0.0.0:20006 -p <public key in hex> -n tom -b 127.0.0.1:20000
```

```bash
# 4 servers distributed on different networks

# on network ryogrid.net (bind to address/port which is accessible from The Internet)
./nostrp2p server  -l 0.0.0.0:20000 -p <public key in hex> -n ryo

# on network redsky.social (bind to address/port which accessible from The Internet)
./nostrp2p server -l 0.0.0.0:7777 -p <public key in hex> -n alice -b ryogrid.net:9999 

# on network A (bind to address/port which is NOT accessible from The Internet)
./nostrp2p server -l 0.0.0.0:20000 -p <public key in hex> -n bob -b ryogrid.net:8888

# on network B (bind to address/port which is NOT accessible from The Internet)
./nostrp2p server -l 0.0.0.0:20000 -p <public key in hex> -n tom -b redsky.social:7777
```

## Bootstrap Server
- currently two servers are running
  - ryogrid.net:9999
  - ryogrid.net:8888
- **These servers don't response to REST API requests from clients. A server for yourself is also needed to use NostrP2P.** 

## Examples (Client)
- [here](https://github.com/ryogrid/flustr-for-nosp2p/tree/for-nostrp2p)
  - Under construction :)

## <del>About ultra simple prototype system</del> (not work in current impl)
- [here (Japanese)](https://ryogrid.hatenablog.com/entry/2024/02/14/225619)
  - NAT transparent overlay has been implented
  - Posting to overlay NW has been implemented
  - No signature validation
  - No follow feature (global TL only)
  - No data replication
  - No data persistence
  - (Almost) No client
    - Posts are outputed to stdout of server
    - Posting text via REST I/F which accepts HTTP POST of JSON text

