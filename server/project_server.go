package impl

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gen "test-project/buf/gen/go/proto"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NewUserServiceImpl struct {
	gen.UnimplementedUserServiceServer
}

var db *mongo.Client
var projectdb *mongo.Collection

type UserDetail struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
}

func mongoNewClient() (*mongo.Client, *mongo.Collection, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb+srv://dbUser:dbUserPassword@cluster0.kjucuqb.mongodb.net/?retryWrites=true&w=majority"))
	if err != nil {
		return nil, nil, err
	}

	resourceManagerDB := client.Database("ResourceManagement")
	projectsCollection := resourceManagerDB.Collection("People_Service")

	return client, projectsCollection, nil
}

func (n *NewUserServiceImpl) CreateUser(ctx context.Context, i *gen.CreateUserReq) (*gen.CreateUserRes, error) {

	user := i.GetUser()

	client, collection, err := mongoNewClient()
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)

	res, reserr := collection.InsertOne(ctx, bson.D{
		{Key: "Name", Value: user.Name},
	})

	if reserr != nil {
		log.Fatal(reserr)
	}

	fmt.Println(res.InsertedID)

	for i, s := range user.Fields {
		_, err = collection.UpdateOne(
			ctx,
			bson.D{{"Name", user.Name}},
			bson.D{
				{"$set", bson.D{{i, s}}},
			})
		if err != nil {
			log.Fatal(err)
		}
	}

	return &gen.CreateUserRes{User: user}, nil
}

type UserItem struct {
	ID     string            `bson:"_id,omitempty"`
	Name   string            `bson:"name"`
	Fields map[string]string `bson:"fields"`
}

func (n *NewUserServiceImpl) ListUsers(ctx *gen.ListUsersReq, stream gen.UserService_ListUsersServer) error {

	client, clientErr := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://dbUser:dbUserPassword@cluster0.kjucuqb.mongodb.net/?retryWrites=true&w=majority"))
	if clientErr != nil {
		return status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Client Error: %v", clientErr),
		)
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("ResourceManagement").Collection("People_Service")

	connectErr := client.Connect(context.Background())
	if connectErr != nil {
		return status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Connect Error: %v", connectErr),
		)
	}

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cursor Error: %v", err),
		)
	}
	defer cursor.Close(context.Background())

	data := &UserItem{}

	for cursor.Next(context.Background()) {

		err := cursor.Decode(data)
		if err != nil {
			return status.Errorf(
				codes.NotFound,
				fmt.Sprintf("Cursor Decode Error: %v", err),
			)
		}

		stream.Send(&gen.ListUsersRes{
			Users: &gen.User{
				Id:     data.ID,
				Name:   data.Name,
				Fields: data.Fields,
			},
		})
	}

	return nil
}

func (n *NewUserServiceImpl) AddNewField(ctx context.Context, i *gen.ColumnReq) (*gen.ColumnRes, error) {

	client, collection, err := mongoNewClient()
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)

	_, updateErr := collection.UpdateMany(ctx, bson.D{}, bson.D{{"$set", bson.D{{i.Name, 1}}}})
	if updateErr != nil {
		return nil, err
	}

	return &gen.ColumnRes{
		Success: true,
	}, nil
}

func (n *NewUserServiceImpl) DeleteNewField(ctx context.Context, i *gen.ColumnReq) (*gen.ColumnRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteNewField not implemented")
}
