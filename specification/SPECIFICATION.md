# Specification
## Implemented NIPs (Nostr Implementation Possibilities)
- Format of event data, filter on reqest and profile data
  - [NIP-01](https://github.com/nostr-protocol/nips/blob/master/01.md)
  - Seriaization format and transport is different in several parts
    - Not websocket and not JSON text
- Follow list
  - [NIP-02](https://github.com/nostr-protocol/nips/blob/5c796c19fd6330628a0b328bfcf5270cb2bc3aff/02.md)
- Normal post, reply, mention
  - [NIP-10](https://github.com/nostr-protocol/nips/blob/5c796c19fd6330628a0b328bfcf5270cb2bc3aff/10.md) 
- Repost, quote repost
  - [NP-18](https://github.com/nostr-protocol/nips/blob/5c796c19fd6330628a0b328bfcf5270cb2bc3aff/18.md)
- Reaction (favorite, like)
  - [NIP-25](https://github.com/nostr-protocol/nips/blob/5c796c19fd6330628a0b328bfcf5270cb2bc3aff/25.md)
    - Emoji is not supported

 ## NostrP2P specific kind
 - 40000
   - Used when request event data from client to server
   - Any kind of event data are returned according to specified filtering paramaters below
     - since
     - until
     - limit

## REST I/F for client
- Basically body is JSON text and content-type is "application/json"
- Response of "/req" endpoint only is MessagePack serialized binary and content-type is "application/octet-stream"
