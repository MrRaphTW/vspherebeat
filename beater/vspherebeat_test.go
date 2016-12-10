package beater

import (
	"testing"

	"github.com/elastic/beats/libbeat/common"
)

type mockBeatInterface struct {
	Name string
}

func testClusterfuncs(t *testing.T) {
	// This tests evenRender for cluster, and also jsonRenderOnScreen
	testCluster := cluster{dc: "a", name: "b", totalCPU: 3, totalMemory: 4, nbHosts: 5, path: "f", cpuOverallocPreco: 7, ramOverallocPreco: 8}
	myBeat := mockBeatInterface{Name: "i"}
	realResult := testCluster.eventRender(myBeat)
	testCluster.jsonRenderOnScreen()
	//we have no change that @timestamp time.Now() is the same, so we take the
	// one created, this is not tested then.
	expectedResult := common.MapStr{
		"type":              "i",
		"dc":                "a",
		"name":              "b",
		"totalCPU":          3,
		"totalMemory":       4,
		"nbHosts":           5,
		"path":              'f',
		"cpuOverallocPreco": 7,
		"ramOverallocPreco": 8,
		"vsphereType":       "Cluster",
	}

	for key, value := range realResult {
		//we have no chance that @timestamp time.Now() is the same, so we take the
		// one created, this is not tested then.
		if key != "@timestamp" {
			if expectedResult[key] != value {
				t.Error("For Cluster, event rendered is not what is expected.")
			}
		}
	}

}

func testVMfuncs(t *testing.T) {
	// This tests evenRender for cluster, and also jsonRenderOnScreen
	testVM := vm{name: "a", dc: "b", path: "c", cluster: "d", cpuLimit: 5, memoryLimit: 6, diskLimit: 7}
	myBeat := mockBeatInterface{Name: "h"}
	realResult := testVM.eventRender(myBeat)
	testVM.jsonRenderOnScreen()
	//we have no change that @timestamp time.Now() is the same, so we take the
	// one created, this is not tested then.
	expectedResult := common.MapStr{
		"type":        "h",
		"name":        "a",
		"dc":          "b",
		"path":        "c",
		"cluster":     "d",
		"cpuLimit":    5,
		"memoryLimit": 6,
		"diskLimit":   7,
		"vsphereType": "VirtualMachine",
	}

	for key, value := range realResult {
		//we have no chance that @timestamp time.Now() is the same, so we take the
		// one created, this is not tested then.
		if key != "@timestamp" {
			if expectedResult[key] != value {
				t.Error("For VM, event rendered is not what is expected.")
			}
		}
	}

}

func testDataStorefuncs(t *testing.T) {
	// This tests evenRender for cluster, and also jsonRenderOnScreen
	testDS := datastore{dc: "a", name: "b", capacity: 3, freeSpace: 4, path: "e", diskOverallocPreco: 6}
	myBeat := mockBeatInterface{Name: "g"}
	realResult := testDS.eventRender(myBeat)
	testDS.jsonRenderOnScreen()
	//we have no change that @timestamp time.Now() is the same, so we take the
	// one created, this is not tested then.
	expectedResult := common.MapStr{
		"type":               "g",
		"dc":                 "a",
		"name":               "b",
		"capacity":           3,
		"freeSpace":          "eee",
		"path":               "qdsqds",
		"diskOverallocPreco": 6,
		"vsphereType":        "DataStore",
	}

	for key, value := range realResult {
		//we have no chance that @timestamp time.Now() is the same, so we take the
		// one created, this is not tested then.
		if key != "@timestamp" {
			if expectedResult[key] != value {
				t.Error("For VM, event rendered is not what is expected.")
			}
		}
	}

}
