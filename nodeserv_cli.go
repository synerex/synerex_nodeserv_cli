package main

import (
	"context"
	"flag"
	"os"
	"log"
	"fmt"
	"strconv"
	"strings"	
	"google.golang.org/grpc"
	nodeapi "github.com/synerex/synerex_nodeapi"
	nodecapi "github.com/synerex/synerex_nodeserv_controlapi"
)

var (
	nodesrv         = flag.String("nodesrv", "127.0.0.1:9990", "Node Server adderess and port")
	show         = flag.Bool("show",true, "Show Nodeserv information")
	sxmove          = flag.String("sxmove","", "Move provider to different synerex sever [provier id],[synerex id]")
	client nodecapi.NodeControlClient
	conn   *grpc.ClientConn
)

// for git versions
var (
	sha1ver   string // sha1 version used to build the program
	buildTime string // when the executable was built
	gitver    string // git release tag
)

func SwitchServer(prvId, srvId int32) {
	var order nodecapi.Order
	var prvInfo, srvInfo nodecapi.NodeControlInfo
	var switchInfo nodecapi.SwitchInfo
	var oswitchInfo nodecapi.Order_SwitchInfo

	var filter nodecapi.NodeControlFilter

	filter.NodeType = nodeapi.NodeType_PROVIDER
	nodeinfos , err := client.QueryNodeInfos( context.Background(), &filter )
	if err != nil {
		log.Printf("Error on QueryNodeInfos\n", err)
		return
	}

	prvName := ""
	for _, ni := range nodeinfos.Infos {
		if ni.NodeId == prvId {
			prvName = ni.NodeInfo.NodeName
			break
		}
	}
	if prvName == "" {
		fmt.Printf("  ProviderID is invalid\n")
		return
	}

	filter.NodeType = nodeapi.NodeType_SERVER
	nodeinfos , err = client.QueryNodeInfos( context.Background(), &filter )
	if err != nil {
		log.Printf("Error on QueryNodeInfos\n", err)
		return
	}

	srvName := ""
	for _, ni := range nodeinfos.Infos {
		if ni.NodeId == srvId {
			srvName = ni.NodeInfo.NodeName
			break
		}
	}
	if srvName == "" {
		fmt.Printf("  ServerID is invalid\n")
		return
	}

	fmt.Printf("  %d %s Switch Server to %d %s\n", prvId, prvName, srvId, srvName)

	prvInfo.NodeId = prvId
	prvInfo.NodeInfo = &nodeapi.NodeInfo{}
	prvInfo.NodeInfo.NodeType = nodeapi.NodeType_PROVIDER

	srvInfo.NodeId = srvId
	srvInfo.NodeInfo = &nodeapi.NodeInfo{}
	srvInfo.NodeInfo.NodeType = nodeapi.NodeType_SERVER


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
	var filter nodecapi.NodeControlFilter

	filter.NodeType = nodeapi.NodeType_GATEWAY
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
				ni.NodeInfo.NodeName,
				ni.NodeInfo.GwInfo,
				ni.NodeInfo.NodePbaseVersion,
				ni.NodeInfo.WithNodeId,
				ni.NodeInfo.ClusterId,
				ni.NodeInfo.AreaId,
				ni.NodeInfo.ChannelTypes)
	}

	filter.NodeType = nodeapi.NodeType_SERVER
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
				ni.NodeInfo.NodeName,
				ni.NodeInfo.ServerInfo,
				ni.NodeInfo.NodePbaseVersion,
				ni.NodeInfo.WithNodeId,
				ni.NodeInfo.ClusterId,
				ni.NodeInfo.AreaId,
				ni.NodeInfo.ChannelTypes)
	}

	filter.NodeType = nodeapi.NodeType_PROVIDER
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
				srvName = si.NodeInfo.NodeName
				break
			}
		}
		fmt.Printf("  %2d %-12.12s%2d %-16.16s %-10.10s %3d %7d %-10.10s %d\n",
				ni.NodeId,
				ni.NodeInfo.NodeName,
				ni.ServerId,
				srvName,
				ni.NodeInfo.NodePbaseVersion,
				ni.NodeInfo.WithNodeId,
				ni.NodeInfo.ClusterId,
				ni.NodeInfo.AreaId,
				ni.NodeInfo.ChannelTypes)
	}

}

func main() {
	var err error
	var Provider,Server int

	flag.Parse()

	log.Printf("nodeserv_cli(%s) built %s sha1 %s",  gitver, buildTime, sha1ver)


	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure()) // insecure
	conn, err = grpc.Dial(*nodesrv, opts...)
	if err != nil {
		log.Printf("fail to dial: %v", err)
		os.Exit(0)
	}

	client = nodecapi.NewNodeControlClient(conn)

	if *show {
		OutputCurrentSP()	
	} else if *sxmove != "" {
	  //
		ids := strings.Split(*sxmove,",")
		if len(ids) != 2 {
			log.Printf("Please specify [provider id],[synerex id]")
			os.Exit(1)
		}

		Provider, err = strconv.Atoi(ids[0])
		if err != nil {
			fmt.Printf("  ProviderID is invalid\n")
			os.Exit(0)
		}
		Server, err = strconv.Atoi(ids[1])
		if err != nil {
			fmt.Printf("  ServerID is invalid\n")
			os.Exit(0)
		}
		SwitchServer(int32(Provider),int32(Server))
	} 
	
}
