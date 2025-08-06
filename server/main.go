package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/edfun317/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements the unary RPC
func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v ", req.GetName())
	fmt.Println("")

	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	return &pb.HelloReply{
		Message:   fmt.Sprintf("Hello, %s!", req.GetName()),
		Timestamp: time.Now().Unix(),
	}, nil
}

// SayHellos implements the server streaming RPC
func (s *server) SayHellos(req *pb.HelloRequest, stream pb.Greeter_SayHellosServer) error {
	log.Printf("Received streaming request from: %v /n", req.GetName())
	fmt.Println("")
	for i := 0; i < 5; i++ {
		msg := fmt.Sprintf("Hello %s! Message %d", req.GetName(), i+1)
		if err := stream.Send(&pb.HelloReply{
			Message:   msg,
			Timestamp: time.Now().Unix(),
		}); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	return nil
}

// ReceiveHellos implements the client streaming RPC
func (s *server) ReceiveHellos(stream pb.Greeter_ReceiveHellosServer) error {
	var messages []string
	for {
		req, err := stream.Recv()
		if err != nil {
			return stream.SendAndClose(&pb.HelloReply{
				Message:   fmt.Sprintf("Received messages: %v", messages),
				Timestamp: time.Now().Unix(),
			})
		}

		fmt.Printf("Received message from %s: %s", req.GetName(), req.GetMessage())

		fmt.Println("")
		messages = append(messages, fmt.Sprintf("%s says: %s", req.GetName(), req.GetMessage()))
	}
}

// Chat implements the bidirectional streaming RPC
func (s *server) Chat(stream pb.Greeter_ChatServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return nil
		}

		log.Printf("Received message from %s: %s", req.GetName(), req.GetMessage())

		fmt.Println("")
		reply := &pb.HelloReply{
			Message:   fmt.Sprintf("Reply to %s: %s", req.GetName(), req.GetMessage()),
			Timestamp: time.Now().Unix(),
		}
		if err := stream.Send(reply); err != nil {
			return err
		}
	}
}

/* func registerService() error {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to etcd: %v", err)
	}
	defer client.Close()

	serviceKey := "/services/greeter/localhost:50051"
	serviceValue := `{"host":"localhost","port":50051,"protocol":"grpc"}`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Put(ctx, serviceKey, serviceValue)
	if err != nil {
		return fmt.Errorf("failed to register service in etcd: %v", err)
	}

	log.Printf("Service registered in etcd: %s -> %s", serviceKey, serviceValue)
	return nil
} */

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})

	reflection.Register(s)

	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
