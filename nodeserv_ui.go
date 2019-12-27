package main

import (
	"context"
	"flag"
	"bufio"
	"os"
	"log"
	"fmt"
	"strings"
	"strconv"
	"google.golang.org/grpc"
	nodecapi "github.com/synerex/synerex_nodeserv_controlapi"
)

var (
	nodesrv         = flag.String("nodesrv", "127.0.0.1:9990", "Node ID Server")
	client nodecapi.NodeControlClient
	conn   *grpc.ClientConn
)


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
			Provider, _ := strconv.Atoi(strings.TrimSpace(PrvSvr[0]))
			Server, _ := strconv.Atoi(strings.TrimSpace(PrvSvr[1]))
			SwitchServer(int32(Provider),int32(Server))
		} else {
			fmt.Printf("Invalid Input\n")
		}

	}
}

func SwitchServer(prvId, srvId int32) {
	var order nodecapi.Order
	var prvInfo, srvInfo nodecapi.NodeInfo
	var switchInfo nodecapi.SwitchInfo
	var oswitchInfo nodecapi.Order_SwitchInfo

	var filter nodecapi.NodeInfoFilter

	filter.NodeType = nodecapi.NodeType_PROVIDER
	nodeinfos , err := client.QueryNodeInfos( context.Background(), &filter )
	if err != nil {
		log.Printf("Error on QueryNodeInfos\n", err)
		return
	}

	prvName := ""
	for _, ni := range nodeinfos.Infos {
		if ni.NodeId == prvId {
			prvName = ni.NodeName
			break
		}
	}
	if prvName == "" {
		fmt.Printf("  ProviderID is invalid\n")
		return
	}

	filter.NodeType = nodecapi.NodeType_SERVER
	nodeinfos , err = client.QueryNodeInfos( context.Background(), &filter )
	if err != nil {
		log.Printf("Error on QueryNodeInfos\n", err)
		return
	}

	srvName := ""
	for _, ni := range nodeinfos.Infos {
		if ni.NodeId == srvId {
			srvName = ni.NodeName
			break
		}
	}
	if srvName == "" {
		fmt.Printf("  ServerID is invalid\n")
		return
	}

	fmt.Printf("  %d %s Switch Server to %d %s\n", prvId, prvName, srvId, srvName)

	prvInfo.NodeId = prvId
	prvInfo.NodeType = nodecapi.NodeType_PROVIDER

	srvInfo.NodeId = srvId
	srvInfo.NodeType = nodecapi.NodeType_SERVER


	order.OrderType = nodecapi.OrderType_SWITCH_SERVER
	order.TargetNode = &prvInfo

	switchInfo.SxServer = &srvInfo
	oswitchInfo.SwitchInfo = &switchInfo
	order.OrderInfo = &oswitchInfo

	_ , err = client.ControlNodes( context.Background(), &order )
	if err != nil {
		log.Printf("Error on ControlNodes\n", err)
		return
	}
}

// Output Node Information
func OutputCurrentSP() {
	var filter nodecapi.NodeInfoFilter

	filter.NodeType = nodecapi.NodeType_GATEWAY
	nodeinfos , err := client.QueryNodeInfos( context.Background(), &filter )
	if err != nil {
		log.Printf("Error on QueryNodeInfos\n", err)
		return
	}

	fmt.Printf("  GATEWAY\n")
	fmt.Printf("  ID Name         GateWayInfo        NodePBVer With Cluster Area       ChannelTypes\n")
	for _, ni := range nodeinfos.Infos {
		fmt.Printf("  %2d %-12.12s %-18.18s %-10.10s %3d %7d %-10.10s %d\n",
				ni.NodeId,
				ni.NodeName,
				ni.GwInfo,
				ni.NodePbaseVersion,
				ni.WithNodeId,
				ni.ClusterId,
				ni.AreaId,
				ni.ChannelTypes)
	}

	filter.NodeType = nodecapi.NodeType_SERVER
	nodeinfos , err = client.QueryNodeInfos( context.Background(), &filter )
	if err != nil {
		log.Printf("Error on QueryNodeInfos\n", err)
		return
	}
	srvinfos := nodeinfos

	fmt.Printf("\n  SERVER\n")
	fmt.Printf("  ID Name         ServerInfo         NodePBVer With Cluster Area       ChannelTypes\n")
	for _, ni := range nodeinfos.Infos {
		fmt.Printf("  %2d %-12.12s %-18.18s %-10.10s %3d %7d %-10.10s %d\n",
				ni.NodeId,
				ni.NodeName,
				ni.ServerInfo,
				ni.NodePbaseVersion,
				ni.WithNodeId,
				ni.ClusterId,
				ni.AreaId,
				ni.ChannelTypes)
	}

	filter.NodeType = nodecapi.NodeType_PROVIDER
	nodeinfos , err = client.QueryNodeInfos( context.Background(), &filter )
	if err != nil {
		log.Printf("Error on QueryNodeInfos\n", err)
		return
	}

	fmt.Printf("\n  PROVIDER\n")
	fmt.Printf("  ID Name         ConnectServer      NodePBVer With Cluster Area       ChannelTypes\n")
	for _, ni := range nodeinfos.Infos {
		srvName := ""
		for _, si := range srvinfos.Infos {
			if si.NodeId == ni.ServerId {
				srvName = si.NodeName
				break
			}
		}
		fmt.Printf("  %2d %-12.12s%2d %-16.16s %-10.10s %3d %7d %-10.10s %d\n",
				ni.NodeId,
				ni.NodeName,
				ni.ServerId,
				srvName,
				ni.NodePbaseVersion,
				ni.WithNodeId,
				ni.ClusterId,
				ni.AreaId,
				ni.ChannelTypes)
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
	fmt.Printf("  d : to display Node information\n")
	fmt.Printf("  q : to quit\n")
	fmt.Printf("  ProviderID -> ServerID : to change server\n")

	GetInput()
}
