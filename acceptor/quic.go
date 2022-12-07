package acceptor

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/ebar-go/ego/utils/runtime"
	"github.com/ebar-go/znet/codec"
	"github.com/lucas-clemente/quic-go"
	"log"
	"math/big"
	"net"
)

type QUICAcceptor struct {
	*Acceptor
	options Options
}

// Run runs the acceptor
func (acceptor *QUICAcceptor) Listen(onAccept func(conn net.Conn)) (err error) {

	lis, err := quic.ListenAddr(acceptor.schema.Addr, generateTLSConfig(), nil)
	if err != nil {
		return
	}

	// use multiple cpus to improve performance
	for i := 0; i < acceptor.options.Core; i++ {
		go func() {
			defer runtime.HandleCrash()
			acceptor.accept(lis, onAccept)
		}()
	}

	return
}

// accept connection
func (acceptor *QUICAcceptor) accept(lis quic.Listener, onAccept func(conn net.Conn)) {
	for {
		select {
		case <-acceptor.done:
			return
		default:
			conn, err := lis.Accept(context.Background())
			if err != nil {
				// if listener close then return
				log.Printf("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
				continue
			}

			onAccept(codec.NewQUICDecoder(conn))
		}
	}

}

// generateTLSConfig set up a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}
}
