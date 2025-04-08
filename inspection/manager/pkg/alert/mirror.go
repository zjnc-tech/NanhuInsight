package alert

import (
	"fmt"
	"strings"
)

// MirrorNameIndexNone when Mirror index is unknown, fill this by default
const MirrorNameIndexNone = "none"

const delimiter = "%2E"

// MirrorName
type MirrorName string

var (
	GPUNameTpl     = strings.Join([]string{"%s", "gpu", "device", "%s", "cluster"}, "%"+delimiter)
	PodNameTpl     = strings.Join([]string{"%s", "%s", "pod", "device", "%s", "cluster"}, "%"+delimiter)
	NodeNameTpl    = strings.Join([]string{"%s", "node", "device", "%s", "cluster"}, "%"+delimiter)
	JobNameTpl     = strings.Join([]string{"%s", "%s", "job", "device", "%s", "cluster"}, "%"+delimiter)
	ClusterNameTpl = strings.Join([]string{"%s", "cluster"}, delimiter)
)

func GenerateNodeMirrorName(node, cluster string) MirrorName {
	if node == "" {
		node = MirrorNameIndexNone
	}
	if cluster == "" {
		cluster = MirrorNameIndexNone
	}
	return MirrorName(fmt.Sprintf(NodeNameTpl,
		node,
		cluster))
}
