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
		"type":              "i",
		"dc":                "a",
		"name":              "b",
		"totalCPU":          int16(3),
		"totalMemory":       int64(4),
		"nbHosts":           int32(5),
		"path":              "f",
		"cpuOverallocPreco": int(7),
		"ramOverallocPreco": int(8),
		"vsphereType":       "Cluster",
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
		"type":        "h",
		"name":        "a",
		"dc":          "b",
		"path":        "c",
		"cluster":     "d",
		"cpuLimit":    int32(5),
		"memoryLimit": int32(6),
		"diskLimit":   int64(7),
		"vsphereType": "VirtualMachine",
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
		"type":               "g",
		"dc":                 "a",
		"name":               "b",
		"capacity":           int64(3),
		"freeSpace":          int64(4),
		"path":               "e",
		"diskOverallocPreco": int(6),
		"vsphereType":        "DataStore",
	}

	for key, value := range realResult {
		//we have no chance that @timestamp time.Now() is the same, so we take the
		// one created, this is not tested then.
		if key != "@timestamp" {
			assert.Equal(t, expectedResult[key], value, key)

		}
	}

}
