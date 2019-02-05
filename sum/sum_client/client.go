package main

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"context"
	"github.com/khoa-le/grpc-go-course/sum/sumpb"
	"io"
	"time"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

func main() {
	fmt.Println("Hello, I am a client")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer cc.Close()

	s := sumpb.NewSumServiceClient(cc)

	//doUnary(s)

	doErrorUnary(s)

	//doServerStreaming(s)

	//doClientStream(s)

	//doBiDiStreaming(s)


}

func doUnary(s sumpb.SumServiceClient) {
	fmt.Println("Starting to do an Unary RPC...")
	req := &sumpb.SumRequest{
		Sum: &sumpb.Sum{
			A: 4,
			B: 5,
		},
	}
	res, err := s.Sum(context.Background(), req)
	if err != nil {
		log.Fatal("error while calling Sum RPC %v", err)
	}
	log.Printf("Response from Sum : %v", res.Result)
}

func doServerStreaming(s sumpb.SumServiceClient) {
	fmt.Println("Starting to do an Streaming PrimeNumberDecomposition RPC...")
	req := &sumpb.PrimeNumberDecompositionRequest{
		Number: 123589674600,
	}
	stream, err := s.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatal("error while calling Sum RPC %v", err)
	}
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}

		fmt.Println(msg.GetPrimeFactor())
	}
}

func doClientStream(s sumpb.SumServiceClient) {
	fmt.Println("Starting to do a client streaming RPC...")

	numbers := []int64{1, 2, 3, 4}

	stream, err := s.ComputeAverage(context.Background())

	if err != nil {
		log.Fatalf("Error while calling ComputeAverage")
	}

	for _, number := range numbers {
		fmt.Println("Sending request : %v", number)
		stream.Send(&sumpb.ComputeAverageRequest{
			Number: number,
		})
		time.Sleep(100 * time.Millisecond)
	}
	res, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("Error while close and recv from server")
	}
	fmt.Printf("Response ComputeAverage: %v", res)

}
func doBiDiStreaming(s sumpb.SumServiceClient) {
	fmt.Println("Starting to do a FindMaximum Bidi streaming RPC...")

	stream, err := s.FindMaximum(context.Background())

	if err != nil {
		log.Fatalf("Error while opening stream and calling FindMaximum: %v", err)
	}

	waitc := make(chan struct{})

	//send to go routine
	go func() {
		numbers := []int32{4, 5, 43, 7, 19, 9, 32}
		for _, number := range numbers {
			stream.Send(&sumpb.FindMaximumRequest{Number: number})
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	//receive go routine

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalf("Problem while reading server stream: %v", err)
			}

			maximum := res.GetMaximum()

			fmt.Printf("\nReceived a new maximum of ...: %v", maximum)
		}
		close(waitc)
	}()

	<-waitc
}
func doErrorUnary(s sumpb.SumServiceClient){
	fmt.Println("Starting to do a SquareRoot Unary RPC ")
	doErrorCall(s)
}

func doErrorCall(s sumpb.SumServiceClient){
	number := int32(-10)
	res, err :=s.SquareRoot(context.Background(),&sumpb.SquareRootResquest{Number:number})
	if err !=nil{
		respErr, ok :=status.FromError(err)
		if ok{
			//actualy error from RPC (user error)
			fmt.Println(respErr.Message())
			fmt.Println(respErr.Code())
			if respErr.Code() == codes.InvalidArgument{
				fmt.Println("We probably sent a negative number!")
			}
		}else{
			log.Fatalf("Big error calling SquareRoot: %v", err)
		}
	}
	fmt.Printf("Result of square root of %v: %v",number,res.GetNumberRoot())
}