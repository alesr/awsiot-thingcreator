package main

import (
	"context"
	"flag"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/nesto/awsiot-thingcreator/things"
)

func main() {
	thingType := flag.String("type", "FooType", "thing type")
	thingPolicy := flag.String("policy", "FooPolicy", "thing policy")
	flag.Parse()

	// Create session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	}))

	// Create the service's client with the session
	svc := iot.New(sess)

	// Create context. Helpful to handle connections timeout
	// and cancellation of pending requests
	ctx := context.Background()

	thingName := things.CreateThing(ctx, svc, *thingType)

	crtOut, err := things.CreateKeysAndCertificate(ctx, svc, thingName)
	if err != nil {
		log.Fatal("failed to create keys and certificate: ", err.Error())
	}

	if err := things.AttachThingToCertificate(ctx, svc, thingName, crtOut); err != nil {
		log.Fatal("failed to attach thing to certificate: ", err.Error())
	}

	if err := things.AttachCertificateToPolicy(ctx, svc, *thingPolicy, crtOut.CertificateArn); err != nil {
		log.Fatal("failed to attach certificate to policy: ", err.Error())
	}
}
