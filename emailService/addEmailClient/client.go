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
 This client is used to make concurrent request to email server responsible for adding email ids to RabbitMQ
*/

package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"project/emailService/proto"
	"strconv"
	"sync"
)

const (
	address = "0.0.0.0:50051"
)

var (
	concurrentClients = 100
)

// send emailIDs to server talking to RabbitMQ
func sendEmailID(c proto.EmailServiceClient, email string, wg *sync.WaitGroup) {
	defer wg.Done()
	req := &proto.Request{
		EmailId: &proto.Email{Id: email},
	}
	// gRPC call to email service server
	res, err := c.EmailService(context.Background(), req)

	if err != nil {
		log.Fatalf("Error occured, %v\n", err)
	}
	log.Printf("Response: %v", res)
}

func main() {
	// create network connection using grpc
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Error occured %v", err)
	}
	defer conn.Close()

	// create new email service client
	c := proto.NewEmailServiceClient(conn)

	wg := new(sync.WaitGroup)

	// add emails here
	for i := 0; i < concurrentClients; i++ {
		emails := "abc" + strconv.Itoa(i) + ".com"
		wg.Add(1)
		// add emails to rabbitMQ using rabbitMQ server
		sendEmailID(c, emails, wg)
	}
	wg.Wait()
}
