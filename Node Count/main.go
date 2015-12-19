package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
)

// NodeData contains the datatypes corresponding to the json fields in the raw
// node data pulled from the bitnodes.com api.
type NodeData struct {
	Timestamp    int                          `json:"timestamp"`
	TotalNodes   int                          `json:"total_nodes"`
	LatestHeight int                          `json:"latest_height"`
	Nodes        map[string][]json.RawMessage `json:"nodes"` // the array has several different types
}

// VersionAggregate contains a version and the number of nodes running that
// version.
type VersionAggregate struct {
	Version     string
	Appearances int
}

// VASlice is a sortable collection of version aggregates.
type VASlice []VersionAggregate

// Len implements the sort.Sort inferface.
func (vas VASlice) Len() int {
	return len(vas)
}

// Less implements the sort.Sort interface.
func (vas VASlice) Less(i, j int) bool {
	return vas[i].Appearances < vas[j].Appearances
}

// Swap implements the sort.Sort interface.
func (vas VASlice) Swap(i, j int) {
	vas[i], vas[j] = vas[j], vas[i]
}

// Main parses a nodes.json file containing a list of nodes and some stats
// about each node, particularly the version number.
func main() {
	// Read the nodes data into memeory and parse them into a struct.
	rawNodeData, err := ioutil.ReadFile("nodes.json")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	var nd NodeData
	err = json.Unmarshal(rawNodeData, &nd)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Go through each node and extract the version number, tallying up the
	// number of nodes running each version.
	nodeVersions := make(map[string]int)
	for _, dataArray := range nd.Nodes {
		var nodeVersion string
		err = json.Unmarshal(dataArray[1], &nodeVersion)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		nodeSiblings := nodeVersions[nodeVersion]
		nodeVersions[nodeVersion] = nodeSiblings + 1
	}

	// Create a slice containing all of the versions and the number of nodes
	// running each version, and then sort it and print the results.
	var vass VASlice
	for version, count := range nodeVersions {
		vass = append(vass, VersionAggregate{Version: version, Appearances: count})
	}
	sort.Sort(vass)
	for _, vas := range []VersionAggregate(vass) {
		fmt.Println(vas.Appearances, "::", vas.Version)
	}
}
