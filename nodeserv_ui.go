package main

import (
	"context"
	"flag"
	"bufio"
	"os"
	"log"
	"fmt"
	"google.golang.org/grpc"
	nodecapi "github.com/synerex/synerex_nodeserv_controlapi"
)

var (
	nodesrv         = flag.String("nodesrv", "127.0.0.1:9990", "Node ID Server")
	client nodecapi.NodeControlClient
	conn   *grpc.ClientConn
)


	// Returning SERVER_CHANGE command if threre is server change request for the provider
	/*
		for k := range ChangeSvrList {
			if ni.NodeName == ChangeSvrList[k].Provider {

				log.Printf("Returning SERVER_CHANGE command for %s connected to %s\n ",
					ChangeSvrList[k].Provider, ChangeSvrList[k].Server)

				return &nodepb.Response{Ok: false, Command: nodepb.KeepAliveCommand_SERVER_CHANGE, Err: ""}, nil
			}
		}
	*/




// read server change requests like ProviderA -> ServerB
func GetInput() {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		Input := scanner.Text()

		if Input == "d" {
			OutputCurrentSP()
		} else if Input == "q" {
			break
		}

		/*
		PrvSvr := strings.Split(Input, "->")
		Provider := strings.TrimSpace(PrvSvr[0])
		Server := strings.TrimSpace(PrvSvr[1])

		ChangeSvrList = append(ChangeSvrList, ProviderConn{
			Provider: Provider,
			Server:   Server,
		})
		*/
	}
}

// Output Current Server Provider Maps
func OutputCurrentSP() {
	var filter nodecapi.NodeInfoFilter
	filter.NodeType = nodecapi.NodeType_PROVIDER
	nodeinfos , err := client.QueryNodeInfos( context.Background(), &filter )
	if err != nil {
		log.Printf("Error on QueryNodeInfos\n", err)
		return
	}
	
	log.Printf("Current Server Provider Connections\n")
	for _, ni := range nodeinfos.Infos {
		log.Printf("Provider info %d %s connected to %d\n", ni.NodeId, ni.NodeName, ni.ServerId)
	}
}

func main() {
	flag.Parse()

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure()) // insecure
	var err error
	conn, err = grpc.Dial(*nodesrv, opts...)
	if err != nil {
		log.Printf("fail to dial: %v", err)
		os.Exit(0)
	}

	client = nodecapi.NewNodeControlClient(conn)

	fmt.Printf("Please input\n")
	fmt.Printf("d : to display current Provider - Server connection\n")
	fmt.Printf("q : to quit\n")
	fmt.Printf("ProviderName -> ServerName : to change server\n")

	GetInput()
}
