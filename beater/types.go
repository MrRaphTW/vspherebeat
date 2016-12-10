package beater

import (
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
)

type cluster struct {
	dc                string
	name              string
	totalCPU          int16
	totalMemory       int64
	nbHosts           int32
	path              string
	cpuOverallocPreco int
	ramOverallocPreco int
}

func (theCluster cluster) eventRender(b *beat.Beat) common.MapStr {
	event := common.MapStr{
		"@timestamp":        common.Time(time.Now()),
		"type":              b.Name,
		"dc":                theCluster.dc,
		"name":              theCluster.name,
		"totalCPU":          theCluster.totalCPU,
		"totalMemory":       theCluster.totalMemory,
		"nbHosts":           theCluster.nbHosts,
		"path":              theCluster.path,
		"cpuOverallocPreco": theCluster.cpuOverallocPreco,
		"ramOverallocPreco": theCluster.ramOverallocPreco,
		"vsphereType":       "Cluster",
	}
	return event
}

type vm struct {
	name        string
	dc          string
	path        string
	cluster     string
	cpuLimit    int32
	memoryLimit int32
	diskLimit   int64
}

func (theVM vm) eventRender(b *beat.Beat) common.MapStr {
	event := common.MapStr{
		"@timestamp":  common.Time(time.Now()),
		"name":        theVM.name,
		"type":        b.Name,
		"dc":          theVM.dc,
		"path":        theVM.path,
		"cluster":     theVM.cluster,
		"cpuLimit":    theVM.cpuLimit,
		"memoryLimit": theVM.memoryLimit,
		"diskLimit":   theVM.diskLimit,
		"vsphereType": "VirtualMachine",
	}
	return event
}

type datastore struct {
	dc                 string
	name               string
	capacity           int64
	freeSpace          int64
	path               string
	diskOverallocPreco int
}

func (theDS datastore) eventRender(b *beat.Beat) common.MapStr {
	event := common.MapStr{
		"@timestamp":         common.Time(time.Now()),
		"type":               b.Name,
		"dc":                 theDS.dc,
		"name":               theDS.name,
		"capacity":           theDS.capacity,
		"freeSpace":          theDS.freeSpace,
		"path":               theDS.path,
		"diskOverallocPreco": theDS.diskOverallocPreco,
		"vsphereType":        "DataStore",
	}
	return event
}
