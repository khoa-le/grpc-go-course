package main

import (
	"context"
	"github.com/khoa-le/grpc-go-course/sum/sumpb"
	"fmt"
	"net"
	"log"
	"google.golang.org/grpc"
	"io"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"math"
	"google.golang.org/grpc/reflection"
)

type server struct {
}

func (*server) Sum(ctx context.Context, req *sumpb.SumRequest) (*sumpb.SumResponse, error) {
	fmt.Printf("Sum service was invoke with %v", req)

	result := req.Sum.A + req.Sum.B

	res := &sumpb.SumResponse{
		Result: result,
	}
	return res, nil
}

func (*server) PrimeNumberDecomposition(req *sumpb.PrimeNumberDecompositionRequest, stream sumpb.SumService_PrimeNumberDecompositionServer) error {
	fmt.Printf("PrimeNumberDecomposition service was invoke with %v", req)

	number := req.GetNumber()

	divisor := int64(2)

	for number > 1 {
		if number%divisor == 0 {
			stream.Send(&sumpb.PrimeNumberDecompositionResponse{
				PrimeFactor: divisor,
			})
			number = number / divisor
		} else {
			divisor++
			fmt.Printf("\nDivisor has increase to %v", divisor)
		}
	}
	return nil
}
func (*server) ComputeAverage(stream sumpb.SumService_ComputeAverageServer) error {
	fmt.Printf("ComputeAverage service was invoke with a streaming request")
	total := int64(0)
	count := 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			//finish stream
			average := float64(total) / float64(count)
			return stream.SendAndClose(&sumpb.ComputeAverageResponse{
				Result: average,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}

		total += req.GetNumber()
		count++
	}
	return nil
}

func (*server) FindMaximum(stream sumpb.SumService_FindMaximumServer) error {
	maximum := int32(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}

		number := req.GetNumber()
		if number > maximum {
			maximum = number
			sendErr := stream.Send(&sumpb.FindMaximumResponse{
				Maximum: maximum,
			})
			if sendErr != nil {
				log.Fatalf("Error while sending from server: %v", err)
			}
		}

	}
	return nil
}

func (*server) SquareRoot(ctx context.Context, req *sumpb.SquareRootResquest) (*sumpb.SquareRootResponse, error) {
	fmt.Println("Received Square Root RPC")
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Received a negative number: %v", number))
	}
	return &sumpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	},nil
}

func main() {
	fmt.Println("Hello world")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()

	sumpb.RegisterSumServiceServer(s, &server{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
