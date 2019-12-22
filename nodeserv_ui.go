package main

import (
	"context"
	"flag"
	"bufio"
	"os"
	"log"
	"fmt"
	"strings"
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




func GetInput() {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		Input := scanner.Text()

		if Input == "d" {
			OutputCurrentSP()
		} else if Input == "q" {
			break
		} else if strings.Contains(Input,"->") {
			PrvSvr := strings.Split(Input, "->")
			Provider := strings.TrimSpace(PrvSvr[0])
			Server := strings.TrimSpace(PrvSvr[1])
			SwitchServer(Provider,Server)
		} else {
			fmt.Printf("Wrong Input\n")
		}

	}
}

func SwitchServer(prvName, srvName string) {
	var order nodecapi.Order
	var prvInfo, srvInfo nodecapi.NodeInfo
	var switchInfo nodecapi.SwitchInfo
	var oswitchInfo nodecapi.Order_SwitchInfo

	log.Printf("Switch Order %s -> %s\n", prvName, srvName)

	prvInfo.NodeName = prvName
	prvInfo.NodeType = nodecapi.NodeType_PROVIDER

	srvInfo.NodeName = srvName
	srvInfo.NodeType = nodecapi.NodeType_SERVER


	order.OrderType = nodecapi.OrderType_SWITCH_SERVER
	order.TargetNode = &prvInfo

	switchInfo.SxServer = &srvInfo
	oswitchInfo.SwitchInfo = &switchInfo
	order.OrderInfo = &oswitchInfo

	_ , err := client.ControlNodes( context.Background(), &order )
	if err != nil {
		log.Printf("Error on ControlNodes\n", err)
		return
	}
}

// Output Current Server Provider Maps
func OutputCurrentSP() {
	var filter nodecapi.NodeInfoFilter

	filter.NodeType = nodecapi.NodeType_PROVIDER
	prvinfos , err := client.QueryNodeInfos( context.Background(), &filter )
	if err != nil {
		log.Printf("Error on QueryNodeInfos\n", err)
		return
	}

	filter.NodeType = nodecapi.NodeType_SERVER
	srvinfos , err := client.QueryNodeInfos( context.Background(), &filter )
	if err != nil {
		log.Printf("Error on QueryNodeInfos\n", err)
		return
	}

	fmt.Printf("  Current Server Provider Connections\n")
	for _, pi := range prvinfos.Infos {
		srvName := ""
		for _, si := range srvinfos.Infos {
			if si.NodeId == pi.ServerId {
				srvName = si.NodeName
				break
			}
		}
		fmt.Printf("  %s connected to %s\n", pi.NodeName, srvName)
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

	fmt.Printf("  Please input\n")
	fmt.Printf("  d : to display current Provider - Server connection\n")
	fmt.Printf("  q : to quit\n")
	fmt.Printf("  ProviderName -> ServerName : to change server\n")

	GetInput()
}
