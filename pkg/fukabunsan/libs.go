package fukabunsan // 負荷分散 - ふかぶんさん - Load Balancing

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/bonavadeur/katyusha/pkg/bonalib"
) // VARIABLES
var (
	KUBECONFIG        = KubeConfig()
	CLIENTSET         = GetClientSet()
	NODENAMES         = GetNodenames()
	PODCIDRS          = GetPodsCIDRs()
	NODEIDX           = NodeNamesToIndex(NODENAMES)
	MIPORIN_MATRIX, _ = GetMiporinMatrix()
)

const (
	MIPORIN_URL = "http://miporin.knative-serving.svc.cluster.local/api/weight/okasan/okaasan/kodomo/hello"
)

// Convert IP address for string to binary-32bit
// Example:
// Input: 192.168.1.1
// Output: 11000000 10101000 00000001 00000001
func IP2Int32(ip string) string {
	var binaryIP string
	// Convert string to octets
	octets := strings.Split(ip, ".")

	for _, octet := range octets {
		num, err := strconv.Atoi(octet)
		if err != nil {
			bonalib.Warn("Error while converting IP address:", err)
			return ""
		}
		// Convert to binary and ensure it have 32 bit
		binary := fmt.Sprintf("%08b", num)
		binaryIP += binary // Final answer
	}

	return binaryIP
}

// Check where the given IP address from
func IsPodInPodCIDR(ip string, cidr PodCIDR) bool {
	// Convert ip address to binary
	binaryIP := IP2Int32(ip)
	// Convert CIDR to binary
	binaryCIDR := IP2Int32(cidr.PodIPRange)
	return binaryIP[:cidr.PodPrefix] == binaryCIDR[:cidr.PodPrefix]
}

// Check where request from
func RequestfromNode(ip string) string {
	for _, cidr := range PODCIDRS {
		if IsPodInPodCIDR(ip, cidr) {
			return cidr.Nodename
		}
	}
	return "Request from Unknown Node"
}

// Convert NODENAMES to nodeidx
func NodeNamesToIndex(nodeNames []string) map[string]int {
	nodeIndexMap := make(map[string]int)
	for i, nodeName := range nodeNames {
		nodeIndexMap[nodeName] = i
	}
	return nodeIndexMap
}

// Get Miporin Matrix
func GetMiporinMatrix() ([][]int, error) {
	resp, err := http.Get(MIPORIN_URL)
	if err != nil {
		return nil, fmt.Errorf("fail to fetch Miporin Matrix: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fail to read Miporin Matrix: %v", err)
	}

	var matrix [][]int
	err = json.Unmarshal(body, &matrix)
	if err != nil {
		return nil, fmt.Errorf("fail to pass Miporin JSON: %v", err)
	}
	return matrix, nil
}
