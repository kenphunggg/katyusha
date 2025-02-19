package fukabunsan // 負荷分散 - ふかぶんさん - Load Balancing
import (
	"context"
	"sort"

	"github.com/bonavadeur/katyusha/pkg/bonalib"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CLUSTER INFORMATION

func GetNodenames() []string {
	nodes, err := CLIENTSET.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		bonalib.Warn("Error listing nodes: %v\n", err)
		return []string{}
	}

	ret := []string{}
	for _, node := range nodes.Items {
		ret = append(ret, node.Name)
	}
	sort.Strings(ret)
	return ret
}
