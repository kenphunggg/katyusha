package fukabunsan // 負荷分散 - ふかぶんさん - Load Balancing

import (
	"strings"

	"math/rand"

	"github.com/bonavadeur/katyusha/pkg/bonalib"
)

func (lb *LoadBalancer) LBAlgorithm(lbRequest *LBRequest) *LBResponse {
	bonalib.Info("LBAlgorithm", "lbRequest", lbRequest)

	// GET NODE SOURCE
	node_source := RequestfromNode(strings.Split(lbRequest.SourceIP, ":")[0])
	node_sourceidx := NODEIDX[node_source]

	// FIND RANDOM NODE
	var node_des_idx int
	randomNum := rand.Intn(100) + 1
	tempSum := 0
	for des := range MIPORIN_MATRIX[node_sourceidx] {
		tempSum += MIPORIN_MATRIX[node_sourceidx][des]
		if randomNum <= tempSum {
			node_des_idx = des
			break
		}
	}

	// FIND ALL PODS IP ADDRESS ON THE NODE SELECTED
	var selected_targets []string
	for _, target := range lbRequest.Targets {
		if IsPodInPodCIDR(target, PODCIDRS[node_des_idx]) {
			selected_targets = append(selected_targets, target)
		}
	}

	// SELECT A POD TO DIRECT TRAFFIC TO
	var (
		final_target_idx int
		final_target     string
	)
	if len(selected_targets) == 0 {
		final_target_idx = rand.Intn((len(lbRequest.Targets)))
		final_target = lbRequest.Targets[final_target_idx]
	} else {
		final_target_idx = rand.Intn(len(selected_targets))
		final_target = selected_targets[final_target_idx]
	}
	bonalib.Log("final target", final_target)

	ret := &LBResponse{
		Target:  final_target,
		Headers: make([]*LBResponse_HeaderSchema, 0),
	}
	ret.Headers = append(ret.Headers, &LBResponse_HeaderSchema{
		Field: "Katyusha-F-Field",
		Value: "Katyusha-F-Field",
	})

	return ret
}
