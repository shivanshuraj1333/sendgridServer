/*
Copyright 2022 Shivanshu Raj Shrivastava(shivanshu1333).

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
This server is responsible to add email addresses to RabbitMQ"
*/

package main

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"project/emailService/proto"
)

const (
	address = "0.0.0.0:50051"
)

type emailServer struct {
	proto.UnimplementedEmailServiceServer
}

// EmailService takes request from clients containing email ids and add these email ids to RabbitMQ
func (*emailServer) EmailService(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	// extracting email id from request
	emailId := req.GetEmailId().GetId()

	// adding emailID to rabbitMQ
	err := addToRabbitMQ(emailId)
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}
	return &proto.Response{Res: "Email id " + emailId + " added successfully to rabbitMQ"}, nil
}

// add received email ids to RabbitMQ
func addToRabbitMQ(emailId string) error {
	// opening connection with RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// oppening a channel using the RabbitMQ connection
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declaring a RabbitMQ
	q, err := ch.QueueDeclare(
		"RMQ", // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	body := emailId

	// publishing the email id (as body) to RabbitMQ
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" Message added to RMQ %s\n", body)

	return err
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	fmt.Println("Add Email to RabbitMQ Service initiating....")

	// server listening to "0.0.0.0:50051"
	lis, err := net.Listen("tcp", address)

	if err != nil {
		log.Fatalf("retry")
	}

	// creating new gRPC server
	s := grpc.NewServer()

	// Registering email server
	proto.RegisterEmailServiceServer(s, &emailServer{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
