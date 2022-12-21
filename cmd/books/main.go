package main

import (
	"books/internal/config"
	"books/internal/server"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	envConfs, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	fmt.Println(envConfs.DBSource)

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelFn()

	database, err := config.OpenConnection(ctx, envConfs.DBSource)
	if err != nil {
		log.Fatalf("d.OpenDatabase failed with error: %s", err)
	}
	defer config.CloseConnection(ctx)

	grpcServer, err := server.NewGRPCServer(database)
	if err != nil {
		log.Fatal(err)
	}
	grpcServer.Run(envConfs.ServerAddress)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	userSignal := <-sigChan
	log.Printf("shutting down server with signal: %s", userSignal)
}
