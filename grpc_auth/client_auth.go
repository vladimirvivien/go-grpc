package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	pb "github.com/vladimirvivien/go-grpc/protobuf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	server     = "127.0.0.1"
	serverPort = "50051"
	authServer = server
	authPort   = "50052"
	certFile   = "./../certs/ca.pem"
)

var (
	token string
)

// call auth service to get token
func login(client pb.AuthClient) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	user := &pb.AuthRequest{Uname: "vector", Pwd: "abc123"}
	authResp, err := client.Login(ctx, user)
	if err != nil {
		return "", err
	}
	return authResp.GetToken(), nil
}

// printUSD demonstrates simple binary call from client
func printUSD(client pb.CurrencyServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	curReq := &pb.CurrencyRequest{Code: "USD"}
	curList, err := client.GetCurrencyList(ctx, curReq)
	if err != nil {
		fmt.Println("error in printUSD:", err)
		return
	}

	fmt.Println("\nUSD Countries")
	fmt.Println("-------------")
	for _, cur := range curList.Items {
		fmt.Printf("%-50s%-10s\n", cur.GetCountry(), cur.GetCode())
	}
}

// printEUR demonstrates server stream call from client
func printEUR(client pb.CurrencyServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	curReq := &pb.CurrencyRequest{Code: "EUR"}
	stream, err := client.GetCurrencyStream(ctx, curReq)
	if err != nil {
		log.Fatal("error in printEUR:", err)
	}

	fmt.Println("\nEUR Countries")
	fmt.Println("-------------")

	for {
		// since the service is long-running,
		// this call will return a deadline exceeded error
		cur, err := stream.Recv()

		if err != nil {
			if err == io.EOF {
				break // we're done
			}
			if stat, ok := status.FromError(err); ok {
				switch stat.Code() {
				case codes.InvalidArgument:
					fmt.Println("error in printEUR:", err)
					return
				default:
					// other err type, do something with it
					fmt.Println(err)
					return
				}
			}
		}
		fmt.Printf("%-50s%-10s\n", cur.GetCountry(), cur.GetCode())
	}
}

// addCurrencies demonstrates client to server stream
func addCurrencies(client pb.CurrencyServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	currencies := []*pb.Currency{
		&pb.Currency{Country: "HAITI", Name: "Gourde", Code: "HTG", Number: 332},
		&pb.Currency{Country: "MARTINIQUE", Name: "Euro", Code: "EUR", Number: 978},
		&pb.Currency{Country: "CUBA", Name: "Cuban Peso", Code: "CUP", Number: 192},
		&pb.Currency{Country: "JAMAICA", Name: "Jamaican Dollar", Code: "JMD", Number: 388},
	}

	// setup server stream (remember, not calling server.SaveCurrencyStream yet)
	stream, err := client.SaveCurrencyStream(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// stream Currency values to the server.
	// The streamed data is not yet being saved.
	for _, cur := range currencies {
		if err := stream.Send(cur); err != nil {
			fmt.Println(err)
			return
		}
	}

	curList, err := stream.CloseAndRecv()

	if err != nil {
		if stat, ok := status.FromError(err); ok {
			switch stat.Code() {
			case codes.InvalidArgument:
				fmt.Println("error in addCurrencies:", err)
				return
			default:
				// handle other errors here
				fmt.Println(err)
				return
			}
		}
	}

	fmt.Println("\nSaved currencies")
	fmt.Println("-----------------")
	for _, cur := range curList.Items {
		fmt.Printf("%-50s%-10s\n", cur.GetCountry(), cur.GetCode())
	}
}

// findCurrencies demonstrates bi-directional stream: one direction streams
// requests to the server while receiving replies from the server.
func findCurrencies(client pb.CurrencyServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reqs := []*pb.CurrencyRequest{
		&pb.CurrencyRequest{Code: "CDF"},
		&pb.CurrencyRequest{Code: "AZN"},
		&pb.CurrencyRequest{Number: 392},
		&pb.CurrencyRequest{Code: "QAR"},
		&pb.CurrencyRequest{Number: 949},
	}

	// setup stream
	stream, err := client.FindCurrencyStream(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// goroutine to stream outbound requests to server
	go func() {
		for _, req := range reqs {
			if err := stream.Send(req); err != nil {
				log.Fatal(err)
			}
		}
		if err := stream.CloseSend(); err != nil {
			log.Fatal(err)
		}
	}()

	// handle incoming Currency reponses from stream
	fmt.Println("\nFound Currencies")
	fmt.Println("-----------------")
	for {
		cur, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			if stat, ok := status.FromError(err); ok {
				switch stat.Code() {
				case codes.InvalidArgument:
					fmt.Println("error in findCurrencies:", err)
					return
				default:
					// other err type, do something with it
					fmt.Println(err)
					return
				}
			}
		}
		fmt.Printf("%-50s%-10s\n", cur.GetCountry(), cur.GetCode())
	}
}

// authUnaryIntercept injects authorization metadata header
// into request before it is sent to the server.
func authUnaryIntercept(
	ctx context.Context,
	method string,
	req, reply interface{},
	conn *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	md := metadata.Pairs("authorization", token)
	ctx = metadata.NewOutgoingContext(ctx, md)
	log.Println("*** injected auth header ***")
	return invoker(ctx, method, req, reply, conn, opts...)
}

// streamLogIntecept implements StreamClientInterceptor and is invoked
// right before the client stream object is created.
func authStreamIntercept(
	ctx context.Context,
	desc *grpc.StreamDesc,
	conn *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	md := metadata.Pairs("authorization", token)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return streamer(ctx, desc, conn, method, opts...)
}

func main() {
	authAddr := net.JoinHostPort(authServer, authPort)
	serverAddr := net.JoinHostPort(server, serverPort)

	// setup tls creds
	tlsCreds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		log.Fatal(err)
	}

	authConn, err := grpc.Dial(
		authAddr,
		grpc.WithTransportCredentials(tlsCreds),
	)
	if err != nil {
		log.Fatal(err)
	}
	authClient := pb.NewAuthClient(authConn)

	// NOTE: token is a global mutable, goroutine-unsafe,
	// and intented for 1 user For multi-tenant setup,
	// care must be taken not to share the token.
	token, err = login(authClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("login ok: token %s\n", token)

	// setup insecure connection
	conn, err := grpc.Dial(
		serverAddr,
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithUnaryInterceptor(authUnaryIntercept),
		grpc.WithStreamInterceptor(authStreamIntercept),
	)

	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewCurrencyServiceClient(conn)

	printUSD(client)

	printEUR(client)

	addCurrencies(client)

	findCurrencies(client)
}
