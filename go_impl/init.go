package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Node represents a single node in the distributed system.
type Node struct {
	msgQueue            chan Msg // Incoming channels for distributed message passing
	id                  int      // Unique node id. Higher id means higher starting energy level. id doesn't change throughout the time
	x, y, energyLevel   int      // Input
	isAlive             bool
	isLeader            bool
	isElectionOngoing   bool
	isElectionCandidate bool
	doElectionSent      bool
	lastDoElectionTime  int64
	leader              *Node
	group               []*Node
	groupID             int
	heartBeatSent       bool
	lastHeartBeatTime   int64
	msgTimeout          int64
	doElectionResponses map[int]bool
	lock                sync.Mutex
}

// Msg is an interface for messages passed among distributed nodes.
type Msg interface{}

// HeartBeat is a message to get a response from a node to check whether it is alive.
type HeartBeat struct {
	sender *Node
}

// HeartBeatAck is a response for the HeartBeat message.
type HeartBeatAck struct {
	sender *Node
}

// DoElection is a message sent by a leader election candidate.
type DoElection struct {
	sender            *Node
	senderEnergyLevel int
}

// DoElectionAck is a response for the DoElection message.
type DoElectionAck struct {
	sender    *Node
	standDown bool // Informs the receiver to whether stand down or continue in an election
}

// ElectionResult is a leader election result sent by the newly chosen leader.
type ElectionResult struct {
	sender *Node
}

func newNode(x, y, energyLevel int) *Node {
	return &Node{
		msgQueue:            make(chan Msg, 2),
		x:                   x,
		y:                   y,
		energyLevel:         energyLevel,
		isAlive:             true,
		isLeader:            false,
		isElectionOngoing:   false,
		isElectionCandidate: false,
		doElectionSent:      false,
		doElectionResponses: make(map[int]bool),
		heartBeatSent:       false,
		msgTimeout:          500,
	}
}

// sendMessage sends a message to the node's msgQueue.
func (node *Node) sendMessage(msg Msg) {
	if node.isAlive {
		node.msgQueue <- msg
	}
}

// setId sets the node's id.
func (node *Node) setId(id int) {
	node.id = id
}

// setAsLeader sets the node as the leader.
func (node *Node) setAsLeader() {
	node.setLeader(node)
	node.isLeader = true
}

// setLeader sets the node's leader and resets relevant flags.
func (node *Node) setLeader(leader *Node) {
	node.leader = leader
	node.isElectionOngoing = false
	node.isElectionCandidate = false
	node.heartBeatSent = false
}

// setGroup sets the node's group and initializes the message queue.
func (node *Node) setGroup(group []*Node, groupID int) {
	node.group = group
	node.groupID = groupID
	node.msgQueue = make(chan Msg, len(group)*2)
	for _, n := range group {
		node.doElectionResponses[n.id] = true
	}
}

// getX returns the node's x coordinate.
func (node *Node) getX() int {
	return node.x
}

// getY returns the node's y coordinate.
func (node *Node) getY() int {
	return node.y
}

// getEnergyLevel returns the node's energy level.
func (node *Node) getEnergyLevel() int {
	return node.energyLevel
}

func (node *Node) consumeEnergyToLive() bool {
	node.energyLevel--
	return node.energyLevel > 0
}

func (node *Node) parseMsg(msg Msg) bool {
	switch m := msg.(type) {
	case HeartBeat:
		if node.isLeader {
			if !node.consumeMsgEnergy() {
				return false
			}
			fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Action: HeartBeatAck, Energy Level: %v, (x,y): (%v,%v)\n",
				node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
			m.sender.sendMessage(HeartBeatAck{sender: node})
		} else {
			// Heartbeat received by non-leader!
		}
	case HeartBeatAck:
		node.heartBeatSent = false
	case DoElection:
		if node.energyLevel >= m.senderEnergyLevel {
			node.doElectionResponses[m.sender.id] = false
			if !node.consumeMsgEnergy() {
				return false
			}
			fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Action: DoElectionAck(stanDown), Energy Level: %v, (x,y): (%v,%v)\n",
				node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
			m.sender.sendMessage(DoElectionAck{sender: node, standDown: true})
			node.startElection()
		} else {
			node.isElectionOngoing = true
			node.isElectionCandidate = false
			node.doElectionSent = false
			node.doElectionResponses[m.sender.id] = true
			if !node.consumeMsgEnergy() {
				return false
			}
			fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Action: DoElectionAck(!stanDown), Energy Level: %v, (x,y): (%v,%v)\n",
				node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
			m.sender.sendMessage(DoElectionAck{sender: node, standDown: false})
		}
	case DoElectionAck:
		node.doElectionResponses[m.sender.id] = m.standDown
	case ElectionResult:
		node.setLeader(m.sender)
	}
	return true
}

func (node *Node) checkHeartBeat() bool {
	if !node.heartBeatSent {
		if !node.consumeMsgEnergy() {
			return false
		}
		fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Action: HeartBeat, Energy Level: %v, (x,y): (%v,%v)\n",
			node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
		node.leader.sendMessage(HeartBeat{sender: node})
		node.heartBeatSent = true
		node.lastHeartBeatTime = node.getCurrentTimestamp()
	} else {
		diff := node.getCurrentTimestamp() - node.lastHeartBeatTime
		fmt.Printf("Timeout Diff: %v, Node ID: %v, Group ID: %v, Leader ID: %v\n", diff, node.id, node.groupID, node.leader.id)
		if diff > node.msgTimeout {
			node.heartBeatSent = false
			node.startElection()
		}
	}
	return true
}

func (node *Node) checkElection() bool {
	if node.doElectionSent {
		if node.getCurrentTimestamp()-node.lastDoElectionTime > node.msgTimeout {
			makeLeader := true
			for _, v := range node.doElectionResponses {
				if v {
					makeLeader = false
					break
				}
			}
			if makeLeader {
				for _, n := range node.group {
					node.setAsLeader()
					if !node.consumeMsgEnergy() {
						return false
					}
					fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Action: ElectionResult, Energy Level: %v, (x,y): (%v,%v)\n",
						node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
					n.sendMessage(ElectionResult{sender: node})
				}
			}
			node.doElectionSent = false
		}
	}
	return true
}

func (node *Node) consumeMsgEnergy() bool {
	node.energyLevel -= 2
	return node.energyLevel > 0
}

func (node *Node) startElection() {
	node.isElectionOngoing = true
	node.isElectionCandidate = true
	for _, n := range node.group {
		node.doElectionResponses[n.id] = false
		if n.id > node.id {
			if !node.consumeMsgEnergy() {
				return
			}
			fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Action: DoElection, Energy Level: %v, (x,y): (%v,%v)\n",
				node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
			n.sendMessage(DoElection{sender: node, senderEnergyLevel: node.energyLevel})
		}
	}
	node.doElectionSent = true
	node.lastDoElectionTime = node.getCurrentTimestamp()
}

func (node *Node) getCurrentTimestamp() int64 {
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	return currentTime
}

func (node *Node) toString() string {
	return fmt.Sprintf("Node{Node ID: %v, x=%d, y=%d, isLeader=%t, groupId=%d}", node.id, node.x, node.y, node.isLeader, node.groupID)
}

func (node *Node) run(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Action: Start, Energy Level: %v, (x,y): (%v,%v)\n",
		node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
	nodeRemoved := false
	for node.energyLevel > 0 {
		// fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Action: Inside Run , Energy Level: %v\n",
		// 	node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel)
		select {
		case msg := <-node.msgQueue:
			// fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Msg: %#v , Energy Level: %v\n",
			// 	node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, msg, node.energyLevel)
			if !node.parseMsg(msg) {
				nodeRemoved = true
				// fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Return: Retuned from parseMsg , Energy Level: %v\n",
				// 	node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel)

			}
		default:
			if node.isElectionOngoing {
				if node.isElectionCandidate {
					if !node.checkElection() {
						nodeRemoved = true
						// fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Return: Retuned from checkElection , Energy Level: %v\n",
						// 	node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel)
					}
				}
			} else {
				if !node.isLeader {
					if !node.checkHeartBeat() {
						nodeRemoved = true
						// fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Return: Retuned from checkHeartBeat , Energy Level: %v\n",
						// 	node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel)
					}
				}
			}

			if !node.consumeEnergyToLive() {
				nodeRemoved = true
				// fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Return: Retuned from consumeEnergy , Energy Level: %v\n",
				// 	node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel)
			}

			if nodeRemoved {
				node.isAlive = false
				node.energyLevel = 0 // Set correct energy level in case of a break

				fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Action: NodeRemoved, Energy Level: %v, (x,y): (%v,%v)\n",
					node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)

				return
			}

			// Adding some delay between simulation steps
			// Simulating sleep for 50 milliseconds.
			time.Sleep(50 * time.Millisecond)
		}
	}

	node.isAlive = false
	node.energyLevel = 0 // Set correct energy level in case of a break

	fmt.Printf("Timestamp: %v, Node ID: %v, Group ID: %v, Leader ID: %v, Action: NodeRemoved, Energy Level: %v, (x,y): (%v,%v)\n",
		node.getCurrentTimestamp(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
}

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

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Usage: go run DistributedSystemSimulation.go <input_file>")
		return
	}

	inputFileName := os.Args[1]

	// Initialize nodes from input
	nodes := make([]*Node, 0)
	file, err := os.Open(inputFileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Text()
		tupleData := strings.Split(data[1:len(data)-1], "),(")
		for _, nodeDataStr := range tupleData {
			nodeData := strings.Split(nodeDataStr, ",")
			x, _ := strconv.Atoi(strings.TrimSpace(nodeData[0]))
			y, _ := strconv.Atoi(strings.TrimSpace(nodeData[1]))
			energyLevel, _ := strconv.Atoi(strings.TrimSpace(nodeData[2]))
			nodes = append(nodes, newNode(x, y, energyLevel))
		}
	}

	// Initial leader selection and group formation
	groups := formGroups(nodes)
	fmt.Println("\nInitial list of nodes:\n")
	for _, group := range groups {
		for _, node := range group {
			fmt.Println(node.toString())
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(nodes))

	// Start the simulation by running each node as a separate goroutine
	fmt.Println("\nStarting simulation...\n")
	for _, group := range groups {
		for _, node := range group {
			go node.run(&wg)
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("\nSimulation finished!")
}
