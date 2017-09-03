# gRPC Auth Example
This package shows how to use token-based authorization.
The client first logs in via a service auth (see authsvc).
Upon succesful login, the auth service returns a jwt token
which is then used to validate authorization with each
subsequent request to the server.