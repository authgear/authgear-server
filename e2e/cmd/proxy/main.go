package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/google/martian"
	"github.com/google/martian/httpspec"
	"github.com/google/martian/mitm"

	"github.com/authgear/authgear-server/e2e/cmd/proxy/mockoidc"
	"github.com/authgear/authgear-server/e2e/cmd/proxy/modifier"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Setup OIDC manager server
	oidcmanager, err := mockoidc.NewMockOIDCManager()
	if err != nil {
		log.Fatal(err)
	}
	defer oidcmanager.Shutdown()

	ln, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatal(err)
	}

	oidcmanager.Start(ln)
	log.Println("Mock OIDC manager listening on", oidcmanager.Server.Addr)

	// Setup proxy to override OIDC endpoints
	proxy := martian.NewProxy()
	defer proxy.Close()

	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
		DisableCompression:    true,
	}
	proxy.SetRoundTripper(tr)

	tlsc, err := tls.LoadX509KeyPair("./ssl/ca.crt", "./ssl/ca.key")
	if err != nil {
		log.Fatal(err)
	}
	priv := tlsc.PrivateKey

	x509c, err := x509.ParseCertificate(tlsc.Certificate[0])
	if err != nil {
		log.Fatal(err)
	}

	// Configure the proxy to intercept HTTPS traffic
	config, err := mitm.NewConfig(x509c, priv)
	if err != nil {
		log.Fatal(err)
	}
	proxy.SetMITM(config)

	stack, _ := httpspec.NewStack("proxy")
	proxy.SetRequestModifier(stack)
	proxy.SetResponseModifier(stack)

	stack.AddRequestModifier(&modifier.OIDCModifier{
		Manager: oidcmanager,
	})

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	go proxy.Serve(l)
	log.Println("Proxy listening on", l.Addr().String())

	<-ctx.Done()
	log.Println("Shutting down")
}
