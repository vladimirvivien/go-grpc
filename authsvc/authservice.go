package main

import (
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"golang.org/x/crypto/bcrypt"

	jwt "github.com/dgrijalva/jwt-go"
	pb "github.com/vladimirvivien/go-grpc/protobuf"
)

const (
	uname       = "vector"
	name        = "Vic Vector"
	port        = ":50052"
	secret      = "a1b2c3d"
	pwd         = "abc123"
	srvCertFile = "./../certs/server.crt"
	srvKeyFile  = "./../certs/server.key"
)

type user struct {
	uname,
	name string
	pwd []byte
}

type AuthService struct {
	*user
}

func newAuthService() *AuthService {
	return new(AuthService)
}

// loadUser creates 1 user which will be used
// for all authentication.
func (s *AuthService) loadUser() error {
	password := "abc123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u := &user{
		uname: "vector",
		name:  "Vic Vector",
		pwd:   hash,
	}
	s.user = u
	return nil
}

// Login For this example, the credentials are stored in
// in-memory in variabe uname and only works for 1 user.
// In a realworld example, a database or some store would
// be used for storing the username and hashed password.
func (s *AuthService) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	log.Println("Authorizing user", req.GetUname())
	if req.GetUname() == "" || req.GetPwd() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing uname or password")
	}
	if req.GetUname() != uname {
		return nil, status.Error(codes.PermissionDenied, "invalid user")
	}
	if err := bcrypt.CompareHashAndPassword(s.user.pwd, []byte(req.GetPwd())); err != nil {
		return nil, status.Error(codes.PermissionDenied, "auth failed")
	}

	// create jwt token
	// see reserved claims https://tools.ietf.org/html/rfc7519#section-4.1
	// see jwt example here https://godoc.org/github.com/dgrijalva/jwt-go#example-New--Hmac
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"exp":  15000,
			"sub":  uname,
			"iss":  "authservice",
			"aud":  "user",
			"name": name,
		},
	)

	// this example uses a simple string secret. You can also
	// use JWT package to specify an RSA public cert here as well.
	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Internal, "internal login problem")
	}

	log.Printf("User %s logged in OK with toke %s\n", uname, tokenString)
	return &pb.AuthResponse{Token: tokenString}, nil
}

func main() {
	authService := newAuthService()
	if err := authService.loadUser(); err != nil {
		log.Fatal(err)
	}

	lstnr, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("failed to start server:", err)
	}

	tlsCreds, err := credentials.NewServerTLSFromFile(srvCertFile, srvKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	// setup and register currency service
	authServer := grpc.NewServer(grpc.Creds(tlsCreds))
	pb.RegisterAuthServer(authServer, authService)

	// start service's server
	log.Println("starting auth service on", port)
	if err := authServer.Serve(lstnr); err != nil {
		log.Fatal(err)
	}
}
