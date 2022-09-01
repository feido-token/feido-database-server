package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"sync"
)

const cliArgs = `Arguments:
    srv_ipv4     = IPv4 address on which the server will listen
    srv_port     = TCP port on which the server will listen`

func main() {
	/* CLI option and arguments parsing */
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] <srv_ipv4> <srv_port>\n", path.Base(os.Args[0]))
		fmt.Fprintln(flag.CommandLine.Output(), "\n"+cliArgs)
		fmt.Fprintln(flag.CommandLine.Output(), "\n"+"Options:")
		flag.PrintDefaults()
	}

	dbPathPtr := flag.String("db", "", "optional path to SQLite3 database")

	certPathPtr := flag.String("cert", "", "server certificate path (only ECDSA)")
	keyPathPtr := flag.String("key", "", "server private key path (only ECDSA)")

	flag.Parse()

	argc := len(flag.Args())
	if argc != 2 {
		fmt.Fprintln(flag.CommandLine.Output(), "invalid number of arguments:", argc)
		flag.Usage()
		os.Exit(2)
	}

	srvAddr := flag.Args()[0] + ":" + flag.Args()[1]

	/* MAIN */
	fmt.Println("Welcome to the Demo eID Revocation Database Server (mimicking I-Checkit)")

	certPath, keyPath := "", ""
	if certPathPtr != nil {
		certPath = *certPathPtr
	}
	if keyPathPtr != nil {
		keyPath = *keyPathPtr
	}

	/* Server Configuration */
	var srvConfig = ServerConfig{
		address:  srvAddr,
		dbPath:   *dbPathPtr,
		certPath: certPath,
		keyPath:  keyPath,
	}

	server := new(RevDbServer)
	err := server.Init(srvConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to init SENG Server:\n", err)
		os.Exit(3)
	}

	err = server.Open()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer server.Close()

	// Start Server loop and wait for its termination
	var wg sync.WaitGroup
	wg.Add(1)
	go server.ServerLoop(&wg)

	wg.Wait()

	if server.dbMng != nil {
		server.dbMng.Fini()
	}

	fmt.Println("The eID Database Server has stopped")
}
