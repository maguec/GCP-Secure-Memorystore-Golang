package main

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/valkey-io/valkey-go"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
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

func valkeyDialLogger(ctx context.Context, s1 string, dialer *net.Dialer, tlsconfig *tls.Config) (net.Conn, error) {
	now := time.Now()
	conn, err := net.Dial("tcp", s1)
	log.Printf("Dialing %s took %v", s1, time.Since(now))
	return conn, err
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
		DialCtxFn: valkeyDialLogger,
	}
}

func worker(conf Rconf, i int) {
	if i > 1024 {
		time.Sleep(time.Duration(i/3) * time.Millisecond)
	}
	log.Printf("Worker %d started", i)
	valkeyConf := valkeyConfig(conf)
	rdb, err := valkey.NewClient(valkeyConf)
	if err != nil {
		log.Fatalf("configuration failed: %v", err)
	}
	defer rdb.Close()
	for {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		err := rdb.Do(context.Background(), rdb.B().Ping().Build()).Error()
		if err != nil {
			log.Println(err)
		}
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

	wg := sync.WaitGroup{}
	wg.Add(1)
	for i := 0; i < 10; i++ {
		go func() {
			worker(cfg, i)
			wg.Done()
		}()
	}
}
