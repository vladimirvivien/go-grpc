package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	jwt "github.com/dgrijalva/jwt-go"
	pb "github.com/vladimirvivien/go-grpc/protobuf"
	"github.com/vladimirvivien/go-grpc/util"
)

const (
	port        = ":50051"
	dataFile    = "./../curdata.csv"
	srvCertFile = "./../certs/server.crt"
	srvKeyFile  = "./../certs/server.key"
)

// CurrencyService implements the pb CurrencyServiceServer interface
type CurrencyService struct {
	ds *util.DataStore
}

func newCurrencyService(ds *util.DataStore) *CurrencyService {
	return &CurrencyService{ds: ds}
}

// GetCurrencyList searches (by Code or Number) and return CurrencyList
func (c *CurrencyService) GetCurrencyList(
	ctx context.Context,
	req *pb.CurrencyRequest,
) (*pb.CurrencyList, error) {

	if req.GetNumber() == 0 && req.GetCode() == "" {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"must provide currency number or code",
		)
	}
	items := c.ds.Search(req.GetCode(), req.GetNumber())
	return &pb.CurrencyList{Items: items}, nil
}

// GetCurrencyStream returns matching Currencies as a server stream
func (c *CurrencyService) GetCurrencyStream(
	req *pb.CurrencyRequest,
	stream pb.CurrencyService_GetCurrencyStreamServer,
) error {

	if req.GetNumber() == 0 && req.GetCode() == "" {
		return status.Errorf(
			codes.InvalidArgument,
			"must provide currency number or code",
		)
	}

	items := c.ds.Search(req.GetCode(), req.GetNumber())

	for _, cur := range items {
		if err := stream.Send(cur); err != nil {
			return err
		}
	}

	return nil
}

// SaveCurrencyStream adds Currency values from a stream to currency items
func (c *CurrencyService) SaveCurrencyStream(
	stream pb.CurrencyService_SaveCurrencyStreamServer,
) error {

	curList := new(pb.CurrencyList)
	for {
		cur, err := stream.Recv()

		if err != nil {
			// if done, close sream and return result
			if err == io.EOF {
				c.ds.Add(curList.Items)
				return stream.SendAndClose(curList)
			}
			return err
		}

		if cur.GetName() == "" ||
			cur.Code == "" ||
			cur.Number == 0 || cur.Country == "" {

			return status.Errorf(
				codes.InvalidArgument,
				"invalid request, must provide number or code",
			)
		}

		curList.Items = append(curList.Items, cur)
	}

}

// FindCurrencyStream sends a stream of CurrencyRequest while
// streaming Currency values from server.
func (c *CurrencyService) FindCurrencyStream(
	stream pb.CurrencyService_FindCurrencyStreamServer,
) error {

	for {
		req, err := stream.Recv()

		if err != nil {
			if err == io.EOF {
				return nil // we're done
			}
			return err
		}

		// validate req
		if req.GetNumber() == 0 && req.GetCode() == "" {
			return status.Errorf(
				codes.InvalidArgument,
				"invalid request, must provide number or code",
			)
		}

		items := c.ds.Search(req.GetCode(), req.GetNumber())
		for _, cur := range items {
			if err := stream.Send(cur); err != nil {
				return err
			}
		}
	}
}

// authUnaryIntercept intercepts incoming requests to validate
// jwt token from metadata header "authorization"
func authUnaryIntercept(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	if err := auth(ctx); err != nil {
		return nil, err
	}
	log.Println("authorization OK")
	return handler(ctx, req)
}

// streamAuthIntercept intercepts to validate authorization
func streamAuthIntercept(
	server interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	if err := auth(stream.Context()); err != nil {
		return err
	}
	log.Println("authorization OK")
	return handler(server, stream)
}

func auth(ctx context.Context) error {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(
			codes.InvalidArgument,
			"missing context",
		)
	}

	authString, ok := meta["authorization"]
	if !ok {
		return status.Errorf(
			codes.Unauthenticated,
			"missing authorization",
		)
	}
	// validate token algo
	log.Println("found jwt token")
	jwtToken, err := jwt.Parse(
		authString[0],
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("bad signing method")
			}
			// additional validation goes here.
			return []byte("a1b2c3d"), nil
		},
	)

	if jwtToken.Valid {
		return nil
	}
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return status.Error(codes.Internal, "bad token")
}

func main() {
	ds := util.NewDataStore(dataFile)
	if err := ds.Load(); err != nil {
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
	curService := newCurrencyService(ds)
	grpcServer := grpc.NewServer(
		grpc.Creds(tlsCreds),
		grpc.UnaryInterceptor(authUnaryIntercept),
		grpc.StreamInterceptor(streamAuthIntercept),
	)
	pb.RegisterCurrencyServiceServer(grpcServer, curService)

	// start service's server
	log.Println("starting secure currency rpc service on", port)
	if err := grpcServer.Serve(lstnr); err != nil {
		log.Fatal(err)
	}
}
