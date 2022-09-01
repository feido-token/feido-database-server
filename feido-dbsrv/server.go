package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/pion/dtls/v2/pkg/crypto/selfsign"
	"net"
	"os"
	"sync"
)

type ServerConfig struct {
	address  string
	dbPath   string
	certPath string
	keyPath  string
}

type serverSocket struct {
	tcpAddress  *net.TCPAddr
	tlsConfig   *tls.Config
	tlsListener net.Listener
}

func (s *serverSocket) Open() (err error) {
	if s.tcpAddress == nil {
		return errors.New("Server Socket address not initialized")
	}
	if s.tlsConfig == nil {
		return errors.New("Server TLS Config not initialized")
	}

	s.tlsListener, err = net.ListenTCP(s.tcpAddress.Network(), s.tcpAddress)
	if err != nil {
		return
	}

	s.tlsListener = tls.NewListener(s.tlsListener, s.tlsConfig)
	return
}

func (s *serverSocket) Close() (err error) {
	if s.tcpAddress == nil || s.tlsConfig == nil {
		return
	}
	err = s.tlsListener.Close()
	return
}

// TODO: Still used/needed?
type serverContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type RevDbServer struct {
	srvConn serverSocket
	srvCert tls.Certificate
	dbMng   *databaseManager
	wg      sync.WaitGroup
	srvCtx  *serverContext
}

func (s *RevDbServer) Init(config ServerConfig) (err error) {
	/* Load/Generate server ECDSA keys and certificate */
	if config.certPath != "" && config.keyPath != "" {
		s.srvCert, err = tls.LoadX509KeyPair(config.certPath, config.keyPath)
	} else {
		// Generate a certificate and private key to secure the connection
		fmt.Println("Info: using auto-generated, self-signed server certificate")
		s.srvCert, err = selfsign.GenerateSelfSigned() // TODO: seems to create elliptic curve key(s)
	}
	// certificate error
	if err != nil {
		return
	}

	// TODO: Still used/needed?
	// Create parent context to cleanup handshaking connections on exit.
	s.srvCtx = new(serverContext)
	s.srvCtx.ctx, s.srvCtx.cancel = context.WithCancel(context.Background())

	/* Server TCP/TLS Socket */
	s.srvConn.tcpAddress, err = net.ResolveTCPAddr("tcp4", config.address)
	// IPv4-only atm
	if err != nil {
		return
	}
	s.srvConn.tlsConfig = s.newTlsSrvConfig()

	/* Shared database manager */
	s.dbMng = &databaseManager{}
	if err = s.dbMng.Init(config.dbPath); err != nil {
		return
	}

	return
}

func (s *RevDbServer) newTlsSrvConfig() *tls.Config {
	// Prepare the configuration of the DTLS connection
	config := &tls.Config{
		Certificates: []tls.Certificate{s.srvCert},

		// TODO: pion does not yet support server RSA keys
		CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},

		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,

		ClientAuth: tls.NoClientCert,
	}
	return config
}

func (s *RevDbServer) Open() (err error) {
	/* Create Server's welcome socket */
	if err = s.srvConn.Open(); err != nil {
		return
	}
	return
}

func (s *RevDbServer) Close() (err error) {
	if err = s.srvConn.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "Error on closing server's welcome socket:", err)
	}

	if s.srvCtx != nil {
		s.srvCtx.cancel()
		s.srvCtx = nil
	}
	return
}

func (s *RevDbServer) ServerLoop(wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("Hello from serverLoop")
	s.wg.Add(1)
	go s.srvWelcomeLoop()
	s.wg.Wait()
}

func (s *RevDbServer) srvWelcomeLoop() {
	defer s.wg.Done()
	for {
		fmt.Println("Wait for Credential Service connection")
		conn, err := s.srvConn.tlsListener.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		tlsConn := conn.(*tls.Conn) // tls.Conn
		fmt.Println("New client has connected")

		err = s.setupNewCliSession(tlsConn)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to start client session", err)
		}
	}
}

func (s *RevDbServer) setupNewCliSession(conn *tls.Conn) (err error) {
	hostAddr := conn.RemoteAddr().(*net.TCPAddr)
	fmt.Println("New RemoteAddr: ", hostAddr)

	cliSess := ClientSession{
		cliConn: conn,
		dbMng:   s.dbMng,
		wg:      s.wg,
	}

	s.wg.Add(1)

	// spawn goroutine for appTun
	go cliSess.HandleClient()
	return
}
