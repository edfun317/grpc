package main

import (
	"context"
	"io"
	"log"
	"strconv"
	"time"

	pb "github.com/edfun317/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	//default gRPC port, gRPC over TLS 50052
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)

	// Unary RPC
	log.Println("\n=== Unary RPC ===")
	testUnary(client)

	// Server Streaming RPC
	log.Println("\n=== Server Streaming RPC ===")
	testServerStreaming(client)

	// Client Streaming RPC
	log.Println("\n=== Client Streaming RPC ===")
	testClientStreaming(client)

	// Bidirectional Streaming RPC
	log.Println("\n=== Bidirectional Streaming RPC ===")
	testBidirectionalStreaming(client)
}

func testUnary(client pb.GreeterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "Alice"})
	if err != nil {
		log.Fatalf("SayHello failed: %v", err)
	}
	log.Printf("Response: %s (at %v)", resp.GetMessage(), time.Unix(resp.GetTimestamp(), 0))
}

func testServerStreaming(client pb.GreeterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	stream, err := client.SayHellos(ctx, &pb.HelloRequest{Name: "Bob"})
	if err != nil {
		log.Fatalf("SayHellos failed: %v", err)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Failed to receive: %v", err)
		}
		log.Printf("Response: %s (at %v)", resp.GetMessage(), time.Unix(resp.GetTimestamp(), 0))
	}
}

func testClientStreaming(client pb.GreeterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	stream, err := client.ReceiveHellos(ctx)
	if err != nil {
		log.Fatalf("ReceiveHellos failed: %v", err)
	}

	messages := []string{"Hi", "How are you?", "Good morning!", "ab", "dafd", "adfe"}
	count := 1
	for _, msg := range messages {
		count++
		if err := stream.Send(&pb.HelloRequest{
			Name:    "Charlie" + strconv.Itoa(count),
			Message: msg,
		}); err != nil {
			log.Fatalf("Failed to send: %v", err)
		}
		time.Sleep(time.Second)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Failed to receive response: %v", err)
	}
	log.Printf("Response: %s (at %v)", resp.GetMessage(), time.Unix(resp.GetTimestamp(), 0))
}

func testBidirectionalStreaming(client pb.GreeterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	stream, err := client.Chat(ctx)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	waitc := make(chan struct{})

	// Send messages
	go func() {
		messages := []string{"Hello!", "How's it going?", "Nice weather!"}
		for _, msg := range messages {
			if err := stream.Send(&pb.HelloRequest{
				Name:    "David",
				Message: msg,
			}); err != nil {
				log.Fatalf("Failed to send: %v", err)
			}
			time.Sleep(time.Second)
		}
		if err := stream.CloseSend(); err != nil {
			log.Printf("Failed to close send: %v", err)
		}
	}()

	// Receive messages
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive: %v", err)
			}
			log.Printf("Response: %s (at %v)", resp.GetMessage(), time.Unix(resp.GetTimestamp(), 0))
		}
	}()

	<-waitc
}
