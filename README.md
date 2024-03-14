# What is NostrP2P?
Pure Peer-to-Peer Distributed Microblogging System on NAT Transparent Overlay Network Based on Idea of [Nostr](https://en.wikipedia.org/wiki/Nostr)

## Design Note
- [here](https://gist.github.com/ryogrid/fa2bfa284784c866ad88e3c38445752a)
- [Japanese version](https://gist.github.com/ryogrid/0ba0d825c3bb840dffa519c5ab91d4ff)
  - Japanese version is latest :)

## Technical Overview
- Difference with (general) Nostr
  - Relay server (Nostr)
    - Firstly, in NostrP2P, servers communicate with each other on pure peer-to-peer manner. Clients does not
      - This means that NostrP2P have client-server architecture also
    - Servers handle each recieved event data in a different way though these are not special kind (ex: replacable events) on Nostr because optimization for pure-peer-to-peer architecture is needed
    - Servers of NostrP2P are correspond to Relay server of Nostr but these are more distributed and coodinate with each other
    - Supporse large number of servers because user of NostrP2P need my server for using 
    - Client can trust server it accesses (it is important)
  - Client (Nostr)
    - Almost same role with one of Nostr but its communication protocol between server is little bit different
      - In current plan, transport is REST and data is encoded to binary. Not websocket and Not JSON text
    - Client which is used by User-A only accesses to a server which is managed by the user only  
- Common point with Nostr
  - Data structure of event data is almost same
  - Key pair format and signing method are same
  - Specification like kind number is same if it is for same functionality (at least for now)

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
  -h, --help   Help for NostrP2P
```

```
Usage:
  nostrp2p server [flags]

Flags:
  -h, --help                         Help for server
  -l, --listen-addr-port    string   Address and port to bind to (default: 127.0.0.1:20000)
  -b, --boot-peer-addr-port string   Address and port of a server which already joined NostrP2P network (optional)
  -p, --public-key          string   Your public key (required)
  -n, --nickname            string   Your nickname on NostrP2P (required) [TEMPORAL]
  -d, --debug               bool     If true, debug log is output to stderr (default: false)
```

## Examples (Generate key pair)
- Under construction

## Examples (Server launch)
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
- currently single server is running
  - ryogrid.net:8888
    - this address including port number shoud be specified at launching of your server 
- **These servers don't response to write kind REST API requests from clients. A server for yourself is also needed to use NostrP2P** 

## Client
- [here](https://github.com/ryogrid/flustr-for-nosp2p)
<img src="https://i.gyazo.com/fbed4277dcada30d22fb0c7be7401e7c.png" height="50%" width="50%" />

## Trial of current implemented featues on trial NW
- Please read [this](https://gist.github.com/ryogrid/5080ff36b6786902d40bb4b91de0766e)
  - NAT transparent overlay has been implented
  - Posting to overlay NW has been implemented
  - REST I/F using JSON text as serialized format has been implemented (TEMPORAL)
  - Simple client has been implemented with flutter
  - Event data persistence has been implemented (only server now)
  - No signature validation
  - No follow feature (global TL only)
  - No reply featue
  - No Like feature
  - No data replication

