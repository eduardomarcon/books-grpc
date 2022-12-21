package server

import (
	"books/internal/pb"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

type Server struct {
	mongoCliente *mongo.Client
	pb.UnimplementedBookServiceServer
}

func NewGRPCServer(mongoCliente *mongo.Client) (*Server, error) {
	return &Server{mongoCliente: mongoCliente}, nil
}

func (s *Server) Run(serverAddress string) {
	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		panic(err)
	}

	gsrv := grpc.NewServer()
	pb.RegisterBookServiceServer(gsrv, s)
	reflection.Register(gsrv)

	log.Printf("starting server on address %s", serverAddress)

	go func() {
		gsrv.Serve(listener)
	}()
}

func (s *Server) AddBook(ctx context.Context, in *pb.AddBookRequest) (*pb.AddBookResponse, error) {
	booksCol := s.mongoCliente.Database("books").Collection("books")
	book := bson.D{{"title", in.Title}, {"author", in.Author}}
	bookInserted, err := booksCol.InsertOne(ctx, book)
	if err != nil {
		return nil, err
	}

	oid, ok := bookInserted.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("error to get book id")
	}

	return &pb.AddBookResponse{
		Id: oid.Hex(),
	}, nil
}

func (s *Server) GetBook(ctx context.Context, in *pb.GetBookRequest) (*pb.GetBookResponse, error) {
	bookId, err := primitive.ObjectIDFromHex(in.Id)
	if err != nil {
		return nil, err
	}

	booksCol := s.mongoCliente.Database("books").Collection("books")

	bookFound := new(pb.Book)
	if err = booksCol.FindOne(ctx, bson.M{"_id": bson.M{"$eq": bookId}}).Decode(&bookFound); err != nil {
		return nil, err
	}

	return &pb.GetBookResponse{
		Book: bookFound,
	}, nil
}
