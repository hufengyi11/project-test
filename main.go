package main

// mongodb+srv://dbUser:dbUserPassword@cluster0.kjucuqb.mongodb.net/?retryWrites=true&w=majority

import (
	"context"
	"fmt"
	"log"
	"net"

	gen "test-project/buf/gen/go/proto"
	impl "test-project/server"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	gRPC "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/mgo.v2/bson"
)

// type UserServiceServer struct {
// 	gen.UnimplementedUserServiceServer
// }

func main() {

	server := getNetListener(8080)
	gRPCServer := gRPC.NewServer()

	reflection.Register(gRPCServer)

	gen.RegisterUserServiceServer(gRPCServer, &impl.NewUserServiceImpl{})

	if err := gRPCServer.Serve(server); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

	// mongodb
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb+srv://dbUser:dbUserPassword@cluster0.kjucuqb.mongodb.net/?retryWrites=true&w=majority"))
	if err != nil {
		fmt.Printf("Connect Error: %v \n", err)
	}
	defer client.Disconnect(context.Background())

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		fmt.Printf("Ping Error: %v \n", err)
	}

	database := client.Database("resourceManagerTest")
	// log.Println(database)
	usersCollection := database.Collection("users_test")
	// client.Database("resourceManager").Collection("projects")

	usersCollection.InsertOne(context.Background(), bson.D{
		{Name: "Name", Value: "TestUser"},
	})

	log.Fatalln(gRPCServer.Serve(server))

}

func getNetListener(port uint) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		panic(fmt.Sprintf("failed to listen: %v", err))
	}

	return lis
}
