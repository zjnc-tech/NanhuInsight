package rpc

import pb "infrahi/backend/inspection-manager/pkg/proto"

type JobNodesInfo struct {
	ProcessNodes map[string]string
	BaseNode     map[string]string
}

var CardTypeToString = map[pb.CardType]string{
	pb.CardType_UNIDENTIFIED: "Unidentified",
	pb.CardType_V100:         "Z1120",
	pb.CardType_A100_40:      "Z3200",
	pb.CardType_A100_80:      "Z2120",
	pb.CardType_XPU:          "V5000",
	pb.CardType_IX:           "U2000",
	pb.CardType_MX:           "W64",
	pb.CardType_MT:           "X10000",
}
