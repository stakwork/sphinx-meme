package lightning

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
)

var (
	macLocation = ""
	tlsLocation = ""
	lndIP       = ""
	lndPort     = ""
)

var lightningClient lnrpc.LightningClient

// Init env vars
func Init() {
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Println("no .env file")
	}
	macLocation = os.Getenv("MACAROON_LOCATION")
	tlsLocation = os.Getenv("TLS_LOCATION")
	lndIP = os.Getenv("LND_IP")
	lndPort = os.Getenv("LND_PORT")
	if macLocation == "" || tlsLocation == "" || lndIP == "" || lndPort == "" {
		log.Fatal("missing vars")
	}
}

// LoadLightning ...
func LoadLightning() lnrpc.LightningClient {
	if lightningClient != nil {
		return lightningClient
	}

	tlsCreds, err := credentials.NewClientTLSFromFile(tlsLocation, "")
	if err != nil {
		fmt.Printf("Cannot get node tls credentials %s", err.Error())
		return nil
	}

	macaroonBytes, err := ioutil.ReadFile(macLocation)
	if err != nil {
		fmt.Printf("Cannot read macaroon file %s", err.Error())
		return nil
	}

	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macaroonBytes); err != nil {
		fmt.Printf("Cannot unmarshal macaroon %s", err.Error())
		return nil
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(macaroons.NewMacaroonCredential(mac)),
	}

	fmt.Println("dial", lndIP+":"+lndPort)
	conn, err := grpc.Dial(lndIP+":"+lndPort, opts...)

	if err != nil {
		fmt.Printf("cannot dial to lnd %s", err.Error())
		return nil
	}
	client := lnrpc.NewLightningClient(conn)

	lightningClient = client
	return client
}

// GetInfo ...
func GetInfo() (*lnrpc.GetInfoResponse, error) {
	client := LoadLightning()
	ctx := context.Background()
	getInfoResp, err := client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		fmt.Printf("Cannot get info from node: %s", err.Error())
		return nil, err
	}
	return getInfoResp, nil
}

// SubscribeInvoices ...
func SubscribeInvoices() error {
	client := LoadLightning()
	ctx := context.Background()
	invStream, err := client.SubscribeInvoices(ctx, &lnrpc.InvoiceSubscription{})
	if err != nil {
		fmt.Printf("Cannot get info from node: %s", err.Error())
		return err
	}
	go func() {
		for {
			invoice, err := invStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("%v.SubsribeInvoices(_) = _, %v", client, err)
			}
			fmt.Printf("%+v\n", invoice)
			receiveInvoice(invoice)
		}
	}()
	return nil
}
