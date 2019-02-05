package main

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"github.com/khoa-le/grpc-go-course/greet/greetpb"
	"context"
	"io"
	"time"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
)

func main() {
	fmt.Println("Hello, I am a client")
	tls := false
	opts := grpc.WithInsecure()
	if tls {
		certFile := "ssl/ca.crt"
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			log.Fatalf("Failed loading CA trust certificate: %v", sslErr)
		}
		opts = grpc.WithTransportCredentials(creds)
	}

	cc, err := grpc.Dial("localhost:50051", opts)

	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)

	//doUnary(c)

	//doServerStreaming(c)

	//doClientStreaming(c)

	//doBidiStreaming(c)

	doUnaryWithDeadline(c, 5*time.Second)
	doUnaryWithDeadline(c, 1*time.Second)
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do an Unary RPC...")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Khoa",
			LastName:  "Le",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Greet RPC %v", err)
	}
	log.Printf("Response from Greet : %v", res)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a Server Streaming RPC...")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Khoa",
			LastName:  "Le",
		},
	}
	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling GreetManyTimes RPC %v", err)
	}
	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}
		log.Printf("Response from GreetManyTimes : %v", msg.GetResult())
	}
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a client streaming RPC...")

	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Khoa",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "An",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Minh",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error when calling LongGreet: %v", err)

	}

	for _, req := range requests {
		fmt.Println("Sending request : %v", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("error while recv stream from LongGreetingResponse")
	}

	fmt.Printf("LongGreet Response: %v", res.GetResult())
}

func doBidiStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a client streaming RPC...")

	stream, err := c.GreetEveryOne(context.Background())

	if err != nil {
		log.Fatalf("Error while creating stream : %v", err)
		return
	}

	requests := []*greetpb.GreetEveryOneRequest{
		&greetpb.GreetEveryOneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Khoa",
			},
		},
		&greetpb.GreetEveryOneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "An",
			},
		},
		&greetpb.GreetEveryOneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Minh",
			},
		},
	}
	waitc := make(chan struct{})
	go func() {
		for _, req := range requests {
			fmt.Printf("\nSending message: %v", req)
			stream.Send(req)
			time.Sleep(2 * time.Second)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving: %v", err)
				break
			}

			fmt.Printf("\nReceive: %v", res.GetResult())
		}
		close(waitc)

	}()

	<-waitc
}
func doUnaryWithDeadline(c greetpb.GreetServiceClient, seconds time.Duration) {
	fmt.Println("Starting to do an Unary RPC...")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Khoa",
			LastName:  "Le",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), seconds)
	defer cancel()
	res, err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit! Deadline we exceeded")
			} else {
				fmt.Printf("unexpected error: %v", statusErr)
			}
		} else {
			log.Fatalf("error while calling GreetWithDeadline RPC: %v", err)
		}
	}
	log.Printf("Response from GreetWithDeadline : %v", res)
}
