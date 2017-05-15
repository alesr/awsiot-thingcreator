package things

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/google/uuid"
)

// CreateThing creates AWS IOT thing
func CreateThing(ctx context.Context, svc *iot.IoT, thingType string) string {
	name := uuid.Must(uuid.NewRandom()).String()
	// Create thing
	params := &iot.CreateThingInput{
		ThingName: &name,
		AttributePayload: &iot.AttributePayload{
			Attributes: map[string]*string{
				"Key": aws.String("AttributeValue"), // Required
				// More values...
			},
			Merge: aws.Bool(true),
		},
		ThingTypeName: &thingType,
	}

	resp, err := svc.CreateThingWithContext(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("thing created: ", resp)
	return name
}

// CreateKeysAndCertificate creates public and private keys, certificate,
// and write everything to file at certificates/thingname
func CreateKeysAndCertificate(ctx context.Context, svc *iot.IoT, thingName string) (*iot.CreateKeysAndCertificateOutput, error) {
	// Create Keys and Certificate
	params := &iot.CreateKeysAndCertificateInput{SetAsActive: aws.Bool(true)}
	resp, err := svc.CreateKeysAndCertificateWithContext(ctx, params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return nil, fmt.Errorf("failed to create keys and certificate w/ context: %s - %s", aerr.Code(), aerr.Message())
		}
		return nil, err
	}

	// Create folder named w/ thing name to store keys and certificates
	dir, err := createCertificateFolder(thingName)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate folder %s: %s", thingName, err.Error())
	}

	// Write keys and certificates to file
	if err := writeCertificates(dir, *resp.CertificatePem, *resp.KeyPair.PublicKey, *resp.KeyPair.PrivateKey); err != nil {
		return nil, fmt.Errorf("failed to write certificate for thing %s: %s", thingName, err.Error())
	}

	log.Println("keys and certificated created: ID ", *resp.CertificateId)
	return resp, nil
}

// Create folder under certificates directory to story thing certificate
func createCertificateFolder(name string) (string, error) {
	certificatesBasePath := path.Dir("certificates/")
	filepath := strings.Join([]string{certificatesBasePath, name}, "/")

	if err := os.Mkdir(filepath, os.FileMode(0770)); err != nil {
		return "", fmt.Errorf("failed to create folder for thing %s certificate: %s", name, err.Error())
	}
	return filepath, nil
}

// Write certificates to files in thingName directory
func writeCertificates(dir, crt, publ, priv string) error {
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("failed to change to %s dir in writeCertificates: %s", dir, err.Error())
	}

	files := map[string]string{
		"certificate.pem.crt": crt,
		"public.pem.key":      publ,
		"private.pem.key":     priv,
	}

	for k, v := range files {
		fo, err := os.Create(k)
		if err != nil {
			return err
		}

		_, err = io.Copy(fo, strings.NewReader(v))
		if err != nil {
			return err
		}
		fo.Close()
	}
	return nil
}

// AttachThingToCertificate attach things to certificates
func AttachThingToCertificate(ctx context.Context, svc *iot.IoT, thing string, crt *iot.CreateKeysAndCertificateOutput) error {
	// Attach certificate to thing
	params := &iot.AttachThingPrincipalInput{
		Principal: crt.CertificateArn,
		ThingName: &thing,
	}

	resp, err := svc.AttachThingPrincipalWithContext(ctx, params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return fmt.Errorf("failed to attach thing principal w/ context: %s - %s", aerr.Code(), aerr.Message())
		}
		return fmt.Errorf("failed to attach thing principal w/ context: %s", err.Error())
	}
	log.Println("thing and certificate attached: ", resp.GoString())
	return nil
}

// AttachCertificateToPolicy attach certificate to policy
func AttachCertificateToPolicy(ctx context.Context, svc *iot.IoT, thingPolicy string, crt *string) error {
	params := &iot.AttachPrincipalPolicyInput{
		PolicyName: aws.String(thingPolicy),
		Principal:  crt,
	}
	resp, err := svc.AttachPrincipalPolicyWithContext(ctx, params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return fmt.Errorf("failed to attach certificate to policy w/ context: %s - %s - %s", aerr.Code(), aerr.Message(), aerr.OrigErr())
		}
		return err
	}
	log.Println(resp.String())
	return nil
}
