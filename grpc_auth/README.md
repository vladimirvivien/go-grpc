# gRPC Auth Example
This package shows how to use token-based authorization.
The client first logs in via a service auth (see authsvc).
Upon succesful login, the auth service returns a jwt token
which is then used to validate authorization with each
subsequent request to the server.

The token is sent to the server using two approaches:

#### 1) Injection via Interceptor
This approach (in `client_auth.go`) uses an interceptor to 
manually inject the token into a gRPC metadata header. The
injected token is retreived from the server using a server-side
interceptor.

#### 2) Using `credentials.PerRPCCredentials`
In this approach, the code implements `credentials.PerRPCCredentials` interface
(see `client_auth2.go`). This allows the JWT key to be automatically injected
into the metadata headers with each gRPC request.  The injected token is 
retrieved from the server-side using an interceptor.

#### Run Example
```sh
// start auth server
$> cd authsvc
$> go run *.go

// start currency server
$> cd grpc_auth
$> go run serv_auth.go

// run client
$> go run client_auth.go

// or
$> go run client_auth2.go
```