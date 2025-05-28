package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/alexflint/go-arg"
	"github.com/valkey-io/valkey-go"
)

type Rconf struct {
	Host string
	Port string
	Cert string
}

var args struct {
	Project  string `help:"GCP ProjectID" default:"" arg:"--project, -p, env:GCP_PROJECT"`
	Instance string `help:"Memorystore Instance name" default:"" arg:"--instance, -i, env:MEMORYSTORE_INSTANCE"`
}

func retrieveTokenFunc(yo valkey.AuthCredentialsContext) (valkey.AuthCredentials, error) {
	username := "default"
	password := os.Getenv("TOKEN")
	return valkey.AuthCredentials{Username: username, Password: password}, nil
}

func getSecret(projectID string, secretID string) (Rconf, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	cfg := Rconf{}
	if err != nil {
		return cfg, err
	}

	// Fetch CERT
	secret, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s-cert/versions/latest", projectID, secretID),
	})
	if err != nil {
		return cfg, err
	}
	cfg.Cert = string(secret.Payload.Data)

	// Fetch HOST
	secret, err = client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s-ip/versions/latest", projectID, secretID),
	})
	if err != nil {
		return cfg, err
	}
	cfg.Host = string(secret.Payload.Data)

	// Fetch PORT
	secret, err = client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s-port/versions/latest", projectID, secretID),
	})
	if err != nil {
		return cfg, err
	}
	cfg.Port = string(secret.Payload.Data)

	return cfg, nil
}

func valkeyConfig(cfg Rconf) valkey.ClientOption {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(cfg.Cert))
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	return valkey.ClientOption{
		InitAddress:       []string{addr, addr},
		AuthCredentialsFn: retrieveTokenFunc,
		TLSConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}
}

func main() {
	arg.MustParse(&args)
	if args.Project == "" || args.Instance == "" {
		fmt.Println("Must specify --project and --instance")
		return
	}
	cfg, err := getSecret(args.Project, args.Instance)
	if err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}

	conf := valkeyConfig(cfg)

	ctx := context.Background()
	rdb, err := valkey.NewClient(conf)
	if err != nil {
		log.Fatalf("configuration failed: %v", err)
	}
	defer rdb.Close()
	err = rdb.Do(ctx, rdb.B().Set().Key("key").Value("val").Build()).Error()
	if err != nil {
		log.Fatalf("Failed to set key: %v", err)
	}
	val, err := rdb.Do(ctx, rdb.B().Get().Key("key").Build()).ToString()
	if err != nil {
		log.Fatalf("Failed to get key: %v", err)
	}
	fmt.Printf("Got value: %s\n", val)
}
