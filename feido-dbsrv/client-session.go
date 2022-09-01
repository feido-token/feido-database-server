package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"sync"

	feidoProto "github.com/feido-token/feido-database-server/feido-proto"
	proto "google.golang.org/protobuf/proto"
)

type ClientSession struct {
	cliConn *tls.Conn
	dbMng   *databaseManager
	wg      sync.WaitGroup
}

func (s *ClientSession) HandleClient() (err error) {
	defer s.wg.Done()
	defer s.cliConn.Close()

	// receive message
	buffer := make([]byte, 1500)
	var nread int
	nread, err = s.cliConn.Read(buffer)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to receive client message:", err)
		return
	}

	// parse query
	var queryMsg = &feidoProto.ICheckitQuery{}
	if err = proto.Unmarshal(buffer[:nread], queryMsg); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to unmarshal query message:", err)
		return
	}

	// use content from query message for database lookup
	fmt.Println("Received the following client query:\n",
		"----------\n",
		"Travel Document Number:", queryMsg.TravelDocumentNumber, "\n",
		"Country of Issuance:", queryMsg.CountryOfIssuance, "\n",
		"Document Type:", queryMsg.DocumentType, "\n",
		"----------")
	var isHit bool
	isHit, err = s.dbMng.LookupEID(queryMsg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to look up eID in database:", err)
		return
	}
	fmt.Println("eID revocation status:", isHit)

	// craft response
	var respMsg = &feidoProto.ICheckitResponse{
		IsDbHit: &isHit,
	}

	out, err := proto.Marshal(respMsg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to marshal response message:", err)
		return
	}

	// send response
	var nsent int
	nsent, err = s.cliConn.Write(out)
	if nsent != len(out) {
		fmt.Fprintln(os.Stderr, "Failed to send full response message:", err)
		return
	}

	// done
	return
}
