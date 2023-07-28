package main

import (
	"math"
	"sort"
)

func sortNodes(nodes []*Node) {
	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].getEnergyLevel() > nodes[j].getEnergyLevel()
	})
	i := len(nodes) - 1

	// Set node ids. Higher energy nodes get larger ids.
	for _, node := range nodes {
		node.setId(i)
		i--
	}
}

func calculateDistance(node1, node2 *Node) float64 {
	x1, y1 := node1.getX(), node1.getY()
	x2, y2 := node2.getX(), node2.getY()
	return math.Sqrt(math.Pow(float64(x2-x1), 2) + math.Pow(float64(y2-y1), 2))
}

func formGroups(nodes []*Node) [][]*Node {
	sortNodes(nodes) // Sort by energy level
	groups := make([][]*Node, 0)
	ungroupedNodes := append([]*Node(nil), nodes...)
	groupID := 0

	for len(ungroupedNodes) > 0 {
		leader := ungroupedNodes[0] // Pick the highest energy node
		group := []*Node{leader}
		leader.setAsLeader()
		ungroupedNodes = ungroupedNodes[1:]

		// Pick nodes for the group
		for i := len(ungroupedNodes) - 1; i >= 0; i-- {
			node := ungroupedNodes[i]
			distance := calculateDistance(leader, node)
			if distance <= 20 {
				node.setLeader(leader)
				group = append(group, node)
				ungroupedNodes = append(ungroupedNodes[:i], ungroupedNodes[i+1:]...)
			}
		}

		for _, node := range group {
			node.setGroup(group, groupID)
		}
		groupID++
		groups = append(groups, group)
	}

	return groups
}

// Other utility functions...
