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
 sendgrid server is responsible to get requests from clients having sender email and email body,
 consume message from RabbitMQ and send emails using send grid API
*/

package main

import (
	"context"
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os"
	"project/emailService/proto"
)

const (
	address = "0.0.0.0:50052"
)

type sendGridServer struct {
	proto.UnimplementedSendGridServiceServer
}

// consume RabbitMQ message
func consumeSingleRabbitMQ() (string, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"RMQ", // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")
	if err != nil {
		return "", err
	}

	emailID := (<-msgs).Body

	return string(emailID), nil
}

// send email using sendgrid API
func sendEmail(emailID string, senderEmailID string, emailBody string) error {
	from := mail.NewEmail("Shivanshu", senderEmailID)
	subject := "Sending with SendGrid"
	plainTextContent := emailBody
	htmlContent := "<strong>sending mails with send grid</strong>"
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	to := mail.NewEmail("User", emailID)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	response, err := client.Send(message)

	log.Printf("Succesfully sent email and recevied the response %v\n", response)
	if err != nil {
		return status.Error(codes.Aborted, err.Error())
	}
	return nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// SendGridService takes request from clients, extract sender email ID, and email body
func (*sendGridServer) SendGridService(ctx context.Context, req *proto.SendgridRequest) (*proto.SendGridResponse, error) {

	senderEmailID := req.GetEmailMetadata().GetEmailID()
	emailBody := req.GetEmailMetadata().GetBody()

	// get receiver email id and from connecting to RabbitMQ
	emailID, err := consumeSingleRabbitMQ()

	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	// send email using send grid API
	err = sendEmail(emailID, senderEmailID, emailBody)
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	return &proto.SendGridResponse{Res: "Successfully sent email to " + emailID}, nil
}

func main() {
	fmt.Println("Send Email from extracted ids of RabbitMQ via sendgrid Service initiating....")
	// server listening to "0.0.0.0:50052"
	lis, err := net.Listen("tcp", address)

	if err != nil {
		log.Fatalf("retry")
	}

	// creating new gRPC server
	s := grpc.NewServer()

	// Registering sendEmailServer server
	proto.RegisterSendGridServiceServer(s, &sendGridServer{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v\n", err)
	}
}
