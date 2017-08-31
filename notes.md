# Protobuf and gRPC Notes
Goal for presentation:
- Compare JSON
  - Highlight JSON +'s : humand readable, portable
  - Higlight its -'s : weakly typed, text based, can be bloated
- Make case for better encoding with protobuf
  - uses a formal definition document (IDL)
  - strongly typed
  - binary
  - efficiently encoded
  - portable - available across many languages
  - introduction to Go generator
- Demostrate protobuf encoding:
  - create simple protobuf
  - Show code that saves data in a file
  - Show code that reads data into memory
  - Compare generated file sizes
  - Show graph that compares sizes of protobuf-encoded vs json-encoded data
- Make the case for RPC with protobuf
  - Example of transporting protobuf-encoded payload over HTTP
  - Explain shortcoming of such approach
  - explain lack of robustness
  - Show gRPC framework features
- introduce gRPC
  - introduction to generator tool for gRPC
  - defining rpc services in IDL
  - Using tools to generate code
  - explore code generated
- Demonstrate gRPC
  - demo/discuss unary method sevice calls
  - demo/discuss server and client stream
  - demo/discuss bi-directional
- Discuss/demo good practices and freatures
  [x] Use getter to avoid nil exception (specially in server)
  [x] Raising and propagating erros prooperly 
  [x] Send complex data with errors
  [x] Secure service with TLS
  [x] Service call timeouts
  [ ] Inteceptors - logging 
  [ ] Inteceptors - 

Proactive protection
  [ ] Serviice authorization with tokens
  - Validation deadlines (useful with streams)
  - Context Metadata 
  - Call retry
  - Message size limiting (specially for repeated fields)
  - Connection Limits (limit # of concurrent connections)
    - concurrent stream limit (server side)
    - listener limit (server side)
    - Or, use tapHandler to customize limit
  - TapHandlers
  - Rate Limiting using TapHandler
  - Load balancing
  - Client connection retries with backoff

# Protobuf 
https://developers.google.com/protocol-buffers/

## Go and Protocol Buffers
Go's support for protocol buffer is provided by the following project:

```
https://github.com/golang/protobuf
```
Setting up local Go development environment involves the following steps:
1. Setup `protoc`, the protocol buffer compiler (see https://developers.google.com/protocol-buffers/)
2. Grab the Go protoc generator with `go get github.com/golang/protobuf/protoc-gen-go`

The protocol generator is a protoc plugin that is used to compile the protocol buffer IDL files into Go code.  When it is fetched, it will also pull down package `github.com/golang/proto` which is the library resposible for marshalling and unmarshalling protobuf-encoded binary payloads.

### Compiling protocol buffer IDLs
Given a Go project repository with protobuf IDL files localted in directory `protobuf`, the following commands can be used to compile the files:
```sh
$ cd protobuf
$ protoc --go_out=. *.proto
```

The previous command uses parameter `--go_out` to trigger the `protoc-gen-go` tool (given both `protoc` and `protoc-gen-go` are on the systems $PATH).  The command will compile all files `*.proto` and place the generated Go code in the current directory.  If we want the file to be placed in a different (relative) directory, the following can be used for instance:
```sh
$ protoc --go_out=./goproto ./protobuf/*.proto
```

When the protobuf IDL declares `rpc` based services, the `protoc-gen-go` tool can also be used to generate gRPC framework code as shown below:
```sh
 $ protoc --go_out=plugins=grpc:./goproto ./protobuf/*.proto
```

## Encoding and decoding protobuf with Go
The following notes are based on:
```
https://developers.google.com/protocol-buffers/docs/gotutorial
```

Package `github.com/golang/protobuf/proto` contains runtime types to encode and decode 