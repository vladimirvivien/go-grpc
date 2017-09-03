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

	pb "github.com/vladimirvivien/go-grpc/protobuf"
	"github.com/vladimirvivien/go-grpc/util"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	server     = "127.0.0.1"
	serverPort = "50051"
	authServer = server
	authPort   = "50052"
	certFile   = "./../certs/ca.pem"
	retryMax   = 5
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

func unaryRetryIntercept(
	ctx context.Context,
	method string,
	req, reply interface{},
	conn *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) (err error) {
	for try := 0; try < retryMax; try++ {
		err = invoker(ctx, method, req, reply, conn, opts...)
		if err == nil {
			return
		}

		stat, _ := status.FromError(err)
		log.Printf("Try:%d, stat code: %d", try, stat.Code())
		switch stat.Code() {
		// do not retry on the followings,
		// these should use failfast
		case codes.Canceled,
			codes.DeadlineExceeded,
			codes.InvalidArgument,
			codes.Internal,
			codes.PermissionDenied,
			codes.Unauthenticated,
			codes.Unimplemented,
			codes.Unknown:
			return
		default:
			// delay a bit and check cancellation
			select {
			case <-time.After(time.Millisecond * 700):
				log.Println("Adding delay")
				continue
			case <-ctx.Done():
				log.Println("Cancel detected")
				if stat.Code() == codes.Unavailable {
					<-time.After(time.Millisecond * 700)
					continue
				}
				return
			}
		}

	}
	return
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
		grpc.WithUnaryInterceptor(unaryRetryIntercept),
	)
	if err != nil {
		log.Fatal(err)
	}
	authClient := pb.NewAuthClient(authConn)

	// get token from auth service
	token, err = login(authClient)
	if err != nil {
		log.Fatal(err)
	}
	// create a jwt credential with token
	jwtCreds := util.NewJwtCreds(token)

	// setup insecure connection
	conn, err := grpc.Dial(
		serverAddr,
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithPerRPCCredentials(jwtCreds),
		grpc.WithUnaryInterceptor(unaryRetryIntercept),
	)

	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewCurrencyServiceClient(conn)

	printUSD(client)

	//printEUR(client)

	//addCurrencies(client)

	//findCurrencies(client)
}
