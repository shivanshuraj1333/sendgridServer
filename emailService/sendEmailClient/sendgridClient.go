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
This client is used to make concurrent request to send grid server responsible
for sending emails using given email id and email body.
*/

package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"os"
	"project/emailService/proto"
	"strconv"
	"sync"
)

const (
	address = "0.0.0.0:50052"
)

var (
	concurrentClients = 5
)

// sending request to send grid server to send extract receiver emails and send emails using sender email id and email body
func sendEmailRequest(c proto.SendGridServiceClient, senderEmailID string, emailBody string, wg *sync.WaitGroup) {
	defer wg.Done()
	req := &proto.SendgridRequest{
		EmailMetadata: &proto.Sendgrid{
			EmailID: senderEmailID,
			Body:    emailBody,
		},
	}
	// making gRPC request to send grid service
	res, err := c.SendGridService(context.Background(), req)
	if err != nil {
		log.Fatalf("Error occured, %v\n", err)
	}
	log.Printf("Response %v", res)
}

func main() {
	// create network connection using grpc
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Error occured %v", err)
	}
	defer conn.Close()

	//  create send grid service client
	c := proto.NewSendGridServiceClient(conn)

	wg := new(sync.WaitGroup)

	for i := 0; i < concurrentClients; i++ {
		senderEmail := os.Getenv("SENDER_EMAIL_ID")
		body := "Hi from Shivanshu, email number is " + strconv.Itoa(i)
		wg.Add(1)
		// sending concurrent request to send emails
		go sendEmailRequest(c, senderEmail, body, wg)
	}
	wg.Wait()
}
