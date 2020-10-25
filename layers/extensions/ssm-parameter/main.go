package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"ssm-parameter/extension"
	"ssm-parameter/ipc"
	"syscall"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// :shrugs:
var extClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
var sess = session.Must(session.NewSession())

func main() {
	log.Println("--- Extension Starts ---")
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigs
		log.Println("Extension stops")
		cancel()
	}()

	_, err := extClient.Register(ctx)
	if err != nil {
		panic(err)
	}
	log.Println("Extension registered")

	port := os.Getenv("EXTENSION_HTTP_PORT")
	if len(port) == 0 {
		port = "2772"
	}

	ssmClient := ssm.New(sess)
	ipcServer := ipc.New(port, ssmClient)
	go ipcServer.Start(ctx)

	processEvents(ctx)
}

func processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := extClient.NextEvent(ctx)
			if err != nil {
				log.Println("Error while getting next event", err)
			}

			if res.EventType == extension.Shutdown {
				log.Println("shutdown!")
				return
			}
		}
	}
}
