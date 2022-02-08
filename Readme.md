## Step 1:
### Setup rabbitMQ locally (instructions for mac)
```
brew install rabbitmq 
export PATH=$PATH:/usr/local/sbin 
rabbitmq-server start 
```

## Step 2:
### Setup project
Run `go mod download` from project root </br>
From project root run ` cd emailService`

## Step 3:
### Start local server to add email ids to RabbitMQ
Run `go run addEmailServer/server.go`

## Step 4:
### Send async request from add email client to add email ids (100 emails)
Run `go run addEmailClient/client.go`

## Step 5:
### Generate Send Grid Api Key and store it locally
Create an account with sendgrid and generate your API key reger []() </br>
Create an environment variable with your API Key and registered email id

```
echo "export SENDGRID_API_KEY='YOUR-API-KEY'" > sendgrid.env
echo "export SENDER_EMAIL_ID='YOUR-REGISTRED-EMAIL-ID" > sendgrid.env
echo "sendgrid.env" >> .gitignore
source ./sendgrid.env
```

## Step 6:
### Start send grid server to consume email IDs from RabbitMQ and send emails using sengrid API
Run `source ./sendgrid.env && go run sendgridServer/sendgridServer.go`

## Step 7:
### Send async request from send email client to send emails to email ids consumed from RabbitMQ
Run `source ./sendgrid.env && go run sendEmailClient/sendgridClient.go`
