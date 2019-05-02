package main

import (
	"log"
	pb "tensorflow_serving/apis"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8501", grpc.WithInsecure())

	if err != nil {
		log.Fatal("grpc.Dial error", err)
	}

	defer conn.Close()

	client := pb.NewPredictionServiceClient(conn)
	print(client)
}
