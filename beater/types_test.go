package beater

import (
	"testing"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/stretchr/testify/assert"
)

func TestClusterEventRender(t *testing.T) {

	// This tests evenRender for cluster, and also jsonRenderOnScreen
	testCluster := cluster{dc: "a", name: "b", totalCPU: 3, totalMemory: 4, nbHosts: 5, path: "f", cpuOverallocPreco: 7, ramOverallocPreco: 8}
	myBeat := &beat.Beat{Name: "i"}
	realResult := testCluster.eventRender(myBeat)
	//we have no change that @timestamp time.Now() is the same, so we take the
	// one created, this is not tested then.
	expectedResult := common.MapStr{
		"type":                "i",
		"dc":                  "a",
		"name":                "b",
		"total_cpu":           int16(3),
		"total_memory":        int64(4),
		"hosts_count":         int32(5),
		"path":                "f",
		"cpu_overalloc_preco": int(7),
		"ram_overalloc_preco": int(8),
		"vsphere_type":        "Cluster",
	}

	for key, value := range realResult {
		//we have no chance that @timestamp time.Now() is the same, so we take the
		// one created, this is not tested then.
		if key != "@timestamp" {
			assert.Equal(t, expectedResult[key], value, key)
		}
	}

}

func TestVMEventRender(t *testing.T) {
	// This tests evenRender for cluster, and also jsonRenderOnScreen
	testVM := vm{
		name:        "a",
		dc:          "b",
		path:        "c",
		cluster:     "d",
		cpuLimit:    5,
		memoryLimit: 6,
		diskLimit:   7,
	}
	myBeat := &beat.Beat{Name: "h"}
	realResult := testVM.eventRender(myBeat)
	//we have no change that @timestamp time.Now() is the same, so we take the
	// one created, this is not tested then.
	expectedResult := common.MapStr{
		"type":         "h",
		"name":         "a",
		"dc":           "b",
		"path":         "c",
		"cluster":      "d",
		"cpu_limit":    int32(5),
		"memory_limit": int32(6),
		"disk_limit":   int64(7),
		"vsphere_type": "VirtualMachine",
	}

	for key, value := range realResult {
		//we have no chance that @timestamp time.Now() is the same, so we take the
		// one created, this is not tested then.
		if key != "@timestamp" {
			assert.Equal(t, expectedResult[key], value, key)

		}
	}

}

func TestDataStoreEventRender(t *testing.T) {
	// This tests evenRender for cluster, and also jsonRenderOnScreen
	testDS := datastore{dc: "a", name: "b", capacity: 3, freeSpace: 4, path: "e", diskOverallocPreco: 6}
	myBeat := &beat.Beat{Name: "g"}
	realResult := testDS.eventRender(myBeat)
	//we have no change that @timestamp time.Now() is the same, so we take the
	// one created, this is not tested then.
	expectedResult := common.MapStr{
		"type":                 "g",
		"dc":                   "a",
		"name":                 "b",
		"capacity":             int64(3),
		"free_space":           int64(4),
		"path":                 "e",
		"disk_overalloc_preco": int(6),
		"vsphere_type":         "DataStore",
	}

	for key, value := range realResult {
		//we have no chance that @timestamp time.Now() is the same, so we take the
		// one created, this is not tested then.
		if key != "@timestamp" {
			assert.Equal(t, expectedResult[key], value, key)

		}
	}

}
