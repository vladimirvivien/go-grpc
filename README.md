# Go and the gRPC Framework
If you are reading this, chances are you have some familiarity with Go and are looking to start working with gRPC.  Or, maybe you are have been working with gRPC in a different language and are looking to start using it with Go.  Either way, welcome!

This repository is a collection of code samples that showcases several features of the gRPC-go framework.  Before we jump in too deep, let us start from the beginning with an overview of gRPC, exploring the nuts and bolts of its components.

## Why gRPC
To explain gRPC, let us establish a scenario where we have a financial provider that wants to create a *Currency Service* that allows lookup and validation of currency info (i.e. name, code, country, and ISO number).  The service is spec'd to have the following non-functional properties:
- Be accessible from mobile front-ends (Java, Objective-C)
- Accessible from backend other backends (Python, Java)
- Accessible from Node.js backend to suppor JS front-end
- Provides desktop access (#C) for a reporting tool

Your first reaction, to implement such service, may be to reach for JSON+HTTP (REST) to implement an HTTP-based API server.  And, that would be a good choice as these technologies are mature and well-understood by the developer community.  JSON was a response to the shortcomings of SOAP-era technologies and its soup of acronyms (let's not even mentioned the stuff SOAP replaced).  Similarly, now gRPC is an evolution of this service-based approach that uses RPC to fill the technological gaps left by the REST including:
- JSON offers weak data types
- Lack of standardized machine-enforced interface contracts in JSON
- JSON is a flexible but inefficient text-based encoding 
- Services are exposed as requests/responses of JSON docs causing a lost of semantics (HTTP only has GET,PUT,POST,DELETE, and PATCH)
- Services code and clients are generally manually generated 
- Versioning, update, backward compatibility can be problems

If you are about to say SOAP, stop.  While SOAP-based technologies introduced the machin-readable and enforceable contracts, it had its own shortmings:
- Cumbersome contract definitions that were meant for machine, but often created by human
- Encoding and decoding SOAP was so resource intensive, that it became its own market with vendors selling hardware to do it
- XML over the wire is inefficient for low-bandwidth devices (mobile)
- Tooling varied by platform and language

gRPC attempts to incorporate best practices from these technlogies by leveraging the efficiencies of protocol boffers and HTTP/2.  It also introduce features not found in prior statck like bi-directional streaming allowing the creation of data intensive applications.

## gRPC Overview
gRPC is a "high performance, open-source universal RPC framework".  Specifically gRPC provides all the tools necessary to write services and clients, in a variety of languages, that can communicate by expressing remote services as transparent native methods on the client.

gRPC is designed to work efficiently with usage ranging from datacenter computing to small IoT devices.  It continues where REST and SOAP left off and provides the following features:
- Uses Protocol Buffers as a typed and efficient binary wire format
- Machine-to-machine contracts are defined using a simple interface definition language (IDL)
- Tools to generate code, from IDL, used to implement service methods and client code to invoke those methods remotely
- Uses HTTP/2 which multiplexes long-lived connections for fast and efficient communication
- Support for bi-directional streaming between client and servers
- Extensive middleware API to control structural concerns such as security, logging, and service policy

## Exploring Potocol Buffers
gRPC's efficiency is partly due to its use of protocol buffers (or protobuf).  It is a language-neutral and platform-neutral technology to efficiently serialize data.  Protocol buffers can be used independently of gRPC as a binary wire or storage format.

The first step to using protocol buffers is to define `messages` which are structures that consist of strongly-typed fields representing the data to be encoded.  Protocol buffer supports fields of diverse tyes including numeric, string, boolean, enums, or other messages. 

The following is a simple protobuf definition with two messages: `Currency` and `CurrencyList`:

```protobuf
syntax = "proto3";
package curproto;

message Currency {
    string code = 1; 
    string name = 2;
    int32 number = 3;
    string country = 4;
}

message CurrencyList {
    repeated Currency items = 1;
}
```
*Protocol buffers file [pb-examples/curproto/currency.proto](https://github.com/vladimirvivien/go-grpc/blob/master/pb-examples/curproto/currency.proto)*

Messages are defined in a file with a `.proto` extension (convention) where they can be arranged as complex and nested data structures.

### Compile the .proto file
The protobuf file by itself is not much use.  The next step is to compile the file using the protocol buffers compiler (`protoc`) located at https://developers.google.com/protocol-buffers/.  The compiler does the followings:
- Create source code containing data structres based on the protobuf messages
- Create code that can serialize and deserialize data into the generated structures

The compiler can generate code into several languages using a pluggable architecture. To generate Go code, download the Go generator for protoc using:

```shell
$> go get github.com/golang/protobuf/protoc-gen-go
```
To generate the Go code from the protobuf file, we can use the following command:
```sh
$ protoc --go_out=./curproto ./curproto/currency.proto
```
The previous command uses parameter `--go_out` to specify Go code generation.  It will compile file `currency.proto` in directory `curproto` and place the generated Go code there as well.  The compilation step will generate Go source file `currency.pb.go` which contains code for serialization, deserialization, and struct types matching the messages defined in the .proto file:
```go
type Currency struct {
	Code    string `protobuf:"bytes,1,opt,name=code" json:"code,omitempty"`
	Name    string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Number  int32  `protobuf:"varint,3,opt,name=number" json:"number,omitempty"`
	Country string `protobuf:"bytes,4,opt,name=country" json:"country,omitempty"`
}
...
type CurrencyList struct {
	Items []*Currency `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
}
```
*Generated Go file [pb-examples/curproto/currency.pb.go](https://github.com/vladimirvivien/go-grpc/blob/master/pb-examples/curproto/currency.pb.go)*

### Using Protobuf directly
Once we have the ability to serialize and deserialize our data, we can use it for any purpose where binary encoding can be applied.  For instance, the following example shows how to use protocol buffer as an effeicient format for storage.  The source snippet below loads data from a CSV file and saves it as a protocol buffer encoded file.

```go
import (
    ...
    "github.com/golang/protobuf/proto"
    "github.com/vladimirvivien/go-grpc/pg-example/curproto"
)

const fileName = "data.pb"

func main() {
	currencyItems, err := createPbFromCsv("../curdata.csv")
	if err != nil {
		log.Fatalf("failed to load csv: %v\n", err)
	}

	// encode data as protobuf binary
	data, err := proto.Marshal(currencyItems)
	if err != nil {
		log.Fatal(err)
	}

	// save the encoded binary to file
	if err := ioutil.WriteFile(fileName, data, 0644); err != nil {
		log.Fatal(err)
	}
	log.Println("data file saved as", fileName)
}

// read csv content and return rows as *curproto.CurrencyList
// a type generated from protobuf
func createPbFromCsv(path string) (*curproto.CurrencyList, error) {
	items := make([]*curproto.Currency, 0)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// create CSV reader from file
	reader := csv.NewReader(file)
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		var num int32
		if i, err := strconv.Atoi(row[3]); err == nil {
			num = int32(i)
		}
		// copy row data into protobuf-generated type
		c := &curproto.Currency{
			Country: row[0],
			Name:    row[1],
			Code:    row[2],
			Number:  num,
		}
		items = append(items, c)
	}
	return &curproto.CurrencyList{Items: items}, err
}
```
*Protobuf example file [pb-examples/encode_pb.go](https://github.com/vladimirvivien/go-grpc/blob/master/pb-examples/encode_pb.go)*

When the code (above) is executed, it will produce file `data.pb` with the data encoded using the protocol buffer binary format.  Doing something similar using JSON ([source code](https://github.com/vladimirvivien/go-grpc/blob/master/pb-examples/encode_json.go)), we can compare the resulting data file sizes for a rough comparison in efficiency as shown below:
```shell
-rw-r--r--  1 vvivien  staff    20K  data.js
-rw-r--r--  1 vvivien  staff    10K  data.pb
```
This simple test reveals that the protobuf-encoded file is half the size of the JSON-encoded file. This saving can be even more pronounced when the data is composed of mostly numeric data.

## Creating Services with Protobuf and gRPC
Earlier we have seen how protocol buffers work.  Now, let us see how to use protobuf and the gPRC framework to build efficient and fast RPC services.  When creating services with gRPC, there are three general steps that must be followed:

> 1. Create a protobuf file (IDL) to define messages and service methods
> 2. Compile the protobuf IDL into code to generate types and service interfaces
> 3. Implement the code for the service remote methods


### 1. Define Protocol Buffers IDL
Using the currency service scenario, presented earlier, let us define the protobuf file that contains the messages and the service definition:

```protobuf
syntax = "proto3";
  
message Currency {
    string code = 1; 
    string name = 2;
    int32 number = 3;
    string country = 4;
}

message CurrencyList {
    repeated Currency items = 1;
}

message CurrencyRequest {
    string code = 1;
    int32 number = 2;
}

// CurrencyService exposes methods to call
service CurrencyService {
    rpc GetCurrencyList(CurrencyRequest) returns (CurrencyList){}
}
```
*Protocol Buffers IDL [protobuf/currency.proto](https://github.com/vladimirvivien/go-grpc/blob/master/protobuf/currency.proto)*

The protocol buffer defines the messages that we saw earlier. However, it now also contains a `service` block which defines one or more `rpc` methods.  In our example, service `CurrencyService` :
  - Offers method `GetCurrencyList` 
  - The method takes `CurrencyRequest` as input paremeter 
  - And, returns `CurrencyList` to the client

### 2. Compile the IDL File
As before, the protobuf IDL file needs to be compiled into source code. To generate code for gPRC, however, we need to specify additional protoc parameters to trigger the gRPC plugin which will generate the code necessary to bootstrap the RPC services and remote methods. 

Assuming the IDL file above is located in a folder called `./protobuf`, the following will generate, in addition to the message types, the gRPC code needed to implement remote server methods and the client stubs to call them:
```sh
 $ protoc -I=./protobuf --go_out=plugins=grpc:./protobuf ./protobuf/currency.proto
```
> Notice the additional parameter value in `--go_out`

The compiler will genearate file `currency.pb.go` in the `./protobuf` directory.  The generated source contains the message struct types (as before), but also includes the service methods as a Go interface to be implemented:
```go
type Currency struct {
	Code    string `protobuf:"bytes,1,opt,name=code" json:"code,omitempty"`
	Name    string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Number  int32  `protobuf:"varint,3,opt,name=number" json:"number,omitempty"`
	Country string `protobuf:"bytes,4,opt,name=country" json:"country,omitempty"`
}

type CurrencyList struct {
	Items []*Currency `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
}

type CurrencyRequest struct {
	Code   string `protobuf:"bytes,1,opt,name=code" json:"code,omitempty"`
	Number int32  `protobuf:"varint,2,opt,name=number" json:"number,omitempty"`
}

// service interface for CurrencyService
type CurrencyServiceServer interface {
	GetCurrencyList(context.Context, *CurrencyRequest) (*CurrencyList, error)
}
```
### 3. Implement the service
The last general step is to implement remote methods for the service.  For our this example, we will implement method `GetCurrencyList()` which return a value of type `CurrencyList` as defined in the IDL.

```go
import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/vladimirvivien/go-grpc/protobuf"
	"github.com/vladimirvivien/go-grpc/util"
)

type CurrencyService struct {
	data []*pb.Currency
}

func newCurrencyService(data []*pb.Currency) *CurrencyService {
	return &CurrencyService{data: data}
}

// GetCurrencyList searches (by Code or Number) and return CurrencyList
func (c *CurrencyService) GetCurrencyList(
	ctx context.Context,
	req *pb.CurrencyRequest,
) (*pb.CurrencyList, error) {

	var items []*pb.Currency
	for _, cur := range c.data {
		if cur.GetNumber() == req.GetNumber() || cur.GetCode() == req.GetCode() {
			items = append(items, cur)
		}
	}

	return &pb.CurrencyList{Items: items}, nil
}

func main() {

	// load data into protobuf structures
	data, err := util.LoadPbFromCsv("./../curdata.csv")
	if err != nil {
		log.Fatal(err) // dont start
	}

    lstnr, err := net.Listen("tcp", ":50050")
	if err != nil {
		log.Fatal("failed to start server:", err)
	}

	// setup and register currency service
	curService := newCurrencyService(data)
	grpcServer := grpc.NewServer()
	pb.RegisterCurrencyServiceServer(grpcServer, curService)

	// start service's server
	log.Println("starting currency rpc service on", port)
	if err := grpcServer.Serve(lstnr); err != nil {
		log.Fatal(err)
	}
}
```
*gRPC server file [grpc/server.go](https://github.com/vladimirvivien/go-grpc/blob/master/grpc/server.go)*

In function `main`, the code uses package `grpc` to register the service implementation and bootstrap the server to expose remote method `GetCurrencyList()`.  When the code is executed, an HTTP/2 server will start listening for requests on `TCP` port `50050`.  This is the general pattern that is often used to get a gRPC service up and running.  

### Using the service
At this point, we are ready to write a simple client which can invoke the remote method defined earlier.  The snippet below shows a Go client that calls the service.

>It should be noted that the client stubs used to call the remote service are (can be) generated when the protocol buffer definition file is compiled (see above).

```go
// printUSD demonstrates simple binary call from client
func printUSD(client pb.CurrencyServiceClient) {
	curReq := &pb.CurrencyRequest{Code: "USD"}
	curList, err := client.GetCurrencyList(context.Background(), curReq)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nUSD Countries")
	fmt.Println("-------------")
	for _, cur := range curList.Items {
		fmt.Printf("%-50s%-10s\n", cur.GetCountry(), cur.GetCode())
	}
}

func main() {
	serverAddr := net.JoinHostPort("localhost", "50050")

	// setup insecure connection
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewCurrencyServiceClient(conn)

	printUSD(client)
}
```
Go client file [grpc/client.go](https://github.com/vladimirvivien/go-grpc/blob/master/grpc/client.go)

In function `main()` the code uses package `grpc` to setup connection to the RPC server.  The client stub is generated during protoc compilation and provides an extensive API to communicate with the server.  Note in function `printUSD` the call to `client.GetCurrencyList()` looks like it is a local call.  However, its an abstraction that hides the complicated dance of serialization and deserialization of protocol buffers to communicat with the server.

## Other gRPC Examples
This repository contains an extensive list of gRPC examples and Go.  You may find some of the followings useful:
- [grpc_auth](https://github.com/vladimirvivien/go-grpc/tree/master/grpc_auth): example of implementation of JWT token-based authorization.
- [grpc_err](https://github.com/vladimirvivien/go-grpc/tree/master/grpc_err): shows how to do error handling in gRPC including the use of complex error objects.
- [grpc_intrcpt](https://github.com/vladimirvivien/go-grpc/tree/master/grpc_intrcpt): introduction to intercept for logging.
- [grpc_limits](https://github.com/vladimirvivien/go-grpc/tree/master/grpc_limits): shows how add preventive measures to guard your gRPC service and clients from failures by specifying limits. 
- [grpc_rate](https://github.com/vladimirvivien/go-grpc/tree/master/grpc_rate): shows how to setup limits on the rate at which a service can be called for a given period of times.
- [grpc_retry](https://github.com/vladimirvivien/go-grpc/tree/master/grpc_retry): shows how to use intecerptors to implement a simple retry logic.
- [grpc_tls](https://github.com/vladimirvivien/go-grpc/tree/master/grpc_tls):shows how to setup TLS-based auth on both client and the server.
- [grpc_to](https://github.com/vladimirvivien/go-grpc/tree/master/grpc_to): shows how to use context timeout to indicate to the framework how long a request should take.
