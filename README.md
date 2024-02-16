# What is Buzzoon?
Pure peer-to-peer distributed microblog system on NAT transparent overlay network implemented in Golang inspired by [Nostr](https://en.wikipedia.org/wiki/Nostr)

## Design Note
[here (Japanese)](https://gist.github.com/ryogrid/0ba0d825c3bb840dffa519c5ab91d4ff)

## Build
```bash
$ go build -o buzzoon main.go
```

## Buzoon command usage
```
Usage:
  buzzoon [flags]
  buzzoon [command]

Available Commands:
  help        Help about any command.
  server      Startup server.
  genkey      Generate key pair.

Flags:
  -h, --help   Help for buzzoon
```

```
Usage:
  buzzoon server [flags]

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
./buzzoon server  -l 0.0.0.0:20000 -p <public key in hex> -n origin
./buzzoon server -l 0.0.0.0:20002 -p <public key in hex> -n alice -b 127.0.0.1:20000 
./buzzoon server -l 0.0.0.0:20004 -p <public key in hex> -n bob -b 127.0.0.1:20002
./buzzoon server -l 0.0.0.0:20006 -p <public key in hex> -n tom -b 127.0.0.1:20000
```

```bash
# 4 servers distributed on different networks

# on network ryogrid.net (bind to address/port which is accessible from The Internet)
./buzzoon server  -l 0.0.0.0:20000 -p <public key in hex> -n ryo

# on network redsky.social (bind to address/port which accessible from The Internet)
./buzzoon server -l 0.0.0.0:20000 -p <public key in hex> -n alice -b ryogrid.net:20000 

# on network A (bind to address/port which is NOT accessible from The Internet)
./buzzoon server -l 0.0.0.0:20000 -p <public key in hex> -n bob -b ryogrid.net:20000

# on network B (bind to address/port which is NOT accessible from The Internet)
./buzzoon server -l 0.0.0.0:20000 -p <public key in hex> -n tom -b redsky.social:20000
```

## Examples (Client usage)
- Under construction

## About ultra simple prototype system
- [here (Japanese)](https://ryogrid.hatenablog.com/entry/2024/02/14/225619)
  - No signature validation
  - No follow feature (global TL only)
  - No data replication
  - No persistence
  - No client
    - Posts are outputed to stdout of server
    - Posting text via REST I/F which accept POST of JSON text

