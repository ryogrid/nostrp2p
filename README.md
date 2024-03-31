- [What is NostrP2P?](#what-is-nostrp2p)
- [Design Note](#design-note)
- [Technical Overview](#technical-overview)
  - [Difference and Commpon Points with (General) Nostr](#difference-and-commpon-points-with-general-nostr)
  - [(General) Nostr Architecture](#general-nostr-architecture)
  - [NostrP2P Architecture](#nostrp2p-architecture)
- [Build](#build)
- [NostrP2P Command Usage](#nostrp2p-command-usage)
- [Examples](#examples)
  - [Generate key pair](#generate-key-pair)
  - [Server launch](#server-launch)
- [Bootstrap Server](#bootstrap-server)
- [Client](#client)
- [Trial of Current Implemented Featues on Dedicated NW](#trial-of-current-implemented-featues-on-dedicated-nw)
- [Trial of Current Implemented Features with Nostr Client (Not NostrP2P Client) Using a Protcol Bridge Server](#trial-of-current-implemented-features-with-nostr-client-not-nostrp2p-client-using-a-protcol-bridge-server)

# What is NostrP2P?
- Pure Peer-to-Peer Distributed Microblogging System on NAT Transparent Overlay Network Based on Idea of [Nostr](https://en.wikipedia.org/wiki/Nostr)
- Distributed Microblogging System by All User's Contribution

# Design Note
- [here (English)](https://gist.github.com/ryogrid/fa2bfa284784c866ad88e3c38445752a)
  - English version is latest :)
- [Japanese version](https://gist.github.com/ryogrid/0ba0d825c3bb840dffa519c5ab91d4ff)


# Technical Overview
## Difference and Commpon Points with (General) Nostr
- Difference with (general) Nostr
  - Server (Relay server)
    - Firstly, in NostrP2P, servers communicate with each other on pure peer-to-peer manner. Clients does not
      - This means that NostrP2P have client-server architecture also
    - Servers of NostrP2P are correspond and similar to Relay server of Nostr but these are more distributed and coodinate with each other
    - Servers handle each recieved event data in a different way though these are not special kind one (ex: replacable events) on Nostr because optimization for pure peer-to-peer network architecture is needed
    - Supporse large number of servers because each user of NostrP2P need my server
    - Client can trust server it accesses to
      - This is important for optimization of client's network resource consumption and power consumption
    - Powerful machine is not needed for server because it handles only one user's requests and amount of managing data is not large compred to (general) Nostr
      - If we considered server (Relay server) as a kind of database system, above is obvious 
  - Client
    - Almost same role with one of Nostr but its communication protocol between server is little bit different
      - In current plan, transport is REST and data is encoded to binary. Not websocket and Not JSON text
    - Client which is used by User-A only accesses to a server which is managed by the user only
  - In all of design
    - Microbrogging application specific
      - This means architecture and protcorl of NostrP2P is not for general purpose unlike Nostr
        - "Nostr" is a name of architecture and protocol and not name of SNS and Microblogging system
      - In other words, "NostrP2P" is a name of an microblogging system like "Bluesky"
- Common point with (general) Nostr
  - Data structure of event data is almost same
  - Key pair format and signing method are same
  - Specification like kind number is same if it is for same functionality (at least for now)
  - Functionality realization led by Clients
    - (flexibility may be low compared with general Nostr...)
  
　　  
## (General) Nostr Architecture

```mermaid

classDiagram
    RelayServerA <-- ClientX : Raed/Write
    RelayServerA <-- ClientY : Raed/Write    
    RelayServerB <-- ClientX : Raed/Write
    RelayServerB <-- ClientY : Raed/Write
    RelayServerB <-- ClientZ : Raed/Write
    RelayServerC <-- ClientX : Raed/Write    
    RelayServerC <-- ClientZ : Raed/Write
    RelayServerD <-- ClientZ : Raed/Write
    RelayServerD <-- ClientY : Raed/Write

    namespace Internet {
      class RelayServerA{
      }
      class RelayServerB{
      }
      class RelayServerC{
      }
      class RelayServerD{
      }
    }
    class ClientX{

    }
    class ClientY{

    }
    class ClientZ{

    }
```
  
## NostrP2P Architecture
```mermaid
classDiagram
    ServerA <|--|> ServerB : coodinate
    ServerA <|--|> ServerC : coodinate
    ServerA <|--|> ServerD : coodinate
    ServerB <|--|> ServerC : coodinate
    ServerB <|--|> ServerD : coodinate
    ServerA <-- ClientA : Raed/Write
    ServerB <-- ClientB : Read/Write
    ServerC <-- ClientC : Read/Write
    ServerD <-- ClientD : Read/Write(VPN)
    namespace Internet {
      class ServerA{
      }
      class ServerB{
      }
    }
    namespace NW-C____________ {
      class ServerC{
      }
      class ClientC{
      }      
    }
    namespace NW-D______________ {
      class ServerD{
      }
    }
    namespace NW-A {
      class ClientA{
      }
    }
    namespace NW-B_________ {
      class ClientB{
      }
    }
    namespace NW-X_________ {
      class ClientD{      
      }
    }
    
```


# Build
```bash
$ go build -o nostrp2p main.go
```

# NostrP2P Command Usage
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
# Examples
## Generate key pair
- Under construction

## Server launch
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

# Bootstrap Server
- currently, running server which is accessible from The Internet is below
  - ryogrid.net:8888
    - this address including port number shoud be specified at launching of your server with -b option
- **These servers don't response to write kind REST API requests from clients. A server for yourself is also needed to use NostrP2P!** 

# Client
- [here](https://github.com/ryogrid/flustr-for-nosp2p)
<img src="https://i.gyazo.com/fbed4277dcada30d22fb0c7be7401e7c.png" height="50%" width="50%" />

# Trial of Current Implemented Featues on Dedicated NW
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

# Trial of Current Implemented Features with Nostr Client (Not NostrP2P Client) Using a Protcol Bridge Server
- Please read [this](https://gist.github.com/ryogrid/5080ff36b6786902d40bb4b91de0766e#file-nostrp2p_demo_v3_procedure-md)
