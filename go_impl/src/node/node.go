package node

import (
	"fmt"
	"math/rand"
	"time"
)

type Msg interface{}

type Node struct {
	id                  int
	msgQueue            chan Msg // Incoming channels for distributed message passing
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
}

func NewNode(x, y, energyLevel int) *Node {
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
		msgTimeout:          1000,
	}
}

func (node *Node) SendMessage(msg Msg) {
	if node.isAlive {
		node.msgQueue <- msg
	}
}

func (node *Node) SetID(id int) {
	node.id = id
}

func (node *Node) SetAsLeader() {
	node.SetLeader(node)
	node.isLeader = true
}

func (node *Node) SetLeader(leader *Node) {
	node.leader = leader
	node.isElectionOngoing = false
	node.isElectionCandidate = false
	node.heartBeatSent = false
}

func (node *Node) SetGroup(group []*Node, groupID int) {
	node.group = group
	node.groupID = groupID
	node.msgQueue = make(chan Msg, len(group)*2)
	for _, n := range group {
		node.doElectionResponses[n.id] = true
	}
}

func (node *Node) GetX() int {
	return node.x
}

func (node *Node) GetY() int {
	return node.y
}

func (node *Node) GetEnergyLevel() int {
	return node.energyLevel
}

func (node *Node) ConsumeEnergyToLive() bool {
	node.energyLevel--
	return node.energyLevel > 0
}

func (node *Node) ParseMsg(msg Msg) bool {
	switch msg := msg.(type) {
	case *HeartBeat:
		if node.isLeader {
			heartBeat := msg
			if !node.ConsumeMsgEnergy() {
				return false
			}
			fmt.Printf("Timestamp: %d, Node ID: %d, Group ID: %d, Leader ID: %d, Action: HeartBeatAck, Energy Level: %d, (x,y): (%d,%d)\n",
				time.Now().UnixNano(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
			heartBeat.SendAck(node)
		} else {
			// Non-leader received HeartBeat, do nothing or handle accordingly
		}

	case *HeartBeatAck:
		node.heartBeatSent = false

	case *DoElection:
		doElection := msg
		if node.energyLevel >= doElection.SenderEnergyLevel {
			node.doElectionResponses[doElection.SenderID] = false
			if !node.ConsumeMsgEnergy() {
				return false
			}
			fmt.Printf("Timestamp: %d, Node ID: %d, Group ID: %d, Leader ID: %d, Action: DoElectionAck(standDown), Energy Level: %d, (x,y): (%d,%d)\n",
				time.Now().UnixNano(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
			doElection.Sender.SendMessage(NewDoElectionAck(node, true))
			node.StartElection()
		} else {
			node.isElectionOngoing = true
			node.isElectionCandidate = false
			node.doElectionSent = false
			node.doElectionResponses[doElection.SenderID] = true
			if !node.ConsumeMsgEnergy() {
				return false
			}
			fmt.Printf("Timestamp: %d, Node ID: %d, Group ID: %d, Leader ID: %d, Action: DoElectionAck(!standDown), Energy Level: %d, (x,y): (%d,%d)\n",
				time.Now().UnixNano(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
			doElection.Sender.SendMessage(NewDoElectionAck(node, false))
		}

	case *DoElectionAck:
		doElectionAck := msg
		node.doElectionResponses[doElectionAck.SenderID] = doElectionAck.StandDown

	case *ElectionResult:
		// Election result received, set the sender as the new leader
		electionResult := msg
		node.SetLeader(electionResult.Sender)

	default:
		// Unknown message type, do nothing or handle accordingly
	}

	return true
}

func (node *Node) CheckHeartBeat() bool {
	if !node.heartBeatSent {
		if !node.ConsumeMsgEnergy() {
			return false
		}
		fmt.Printf("Timestamp: %d, Node ID: %d, Group ID: %d, Leader ID: %d, Action: HeartBeat, Energy Level: %d, (x,y): (%d,%d)\n",
			time.Now().UnixNano(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
		node.leader.SendMessage(NewHeartBeat(node))
		node.heartBeatSent = true
		node.lastHeartBeatTime = time.Now().UnixNano()
	} else {
		if time.Now().UnixNano()-node.lastHeartBeatTime > node.msgTimeout {
			// Leader down, start a new election
			node.heartBeatSent = false
			node.StartElection()
		}
	}

	return true
}

func (node *Node) CheckElection() bool {
	if node.doElectionSent {
		if time.Now().UnixNano()-node.lastDoElectionTime > node.msgTimeout {
			makeLeader := true
			for _, standDown := range node.doElectionResponses {
				if standDown {
					makeLeader = false
				}
			}
			if makeLeader {
				for _, n := range node.group {
					node.SetAsLeader()
					if !node.ConsumeMsgEnergy() {
						return false
					}
					fmt.Printf("Timestamp: %d, Node ID: %d, Group ID: %d, Leader ID: %d, Action: ElectionResult, Energy Level: %d, (x,y): (%d,%d)\n",
						time.Now().UnixNano(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
					n.SendMessage(NewElectionResult(node))
				}
			}
			node.doElectionSent = false
		}
	}
	return true
}

func (node *Node) ConsumeMsgEnergy() bool {
	node.energyLevel -= 2
	return node.energyLevel > 0
}

func (node *Node) StartElection() {
	node.isElectionOngoing = true
	node.isElectionCandidate = true
	for _, n := range node.group {
		node.doElectionResponses[n.id] = false
		if n.id > node.id {
			if !node.ConsumeMsgEnergy() {
				return
			}
			fmt.Printf("Timestamp: %d, Node ID: %d, Group ID: %d, Leader ID: %d, Action: DoElection, Energy Level: %d, (x,y): (%d,%d)\n",
				time.Now().UnixNano(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
			n.SendMessage(NewDoElection(node, node.energyLevel))
		}
	}
	node.doElectionSent = true
	node.lastDoElectionTime = time.Now().UnixNano()
}

func (node *Node) Run() {
	rand.Seed(time.Now().UnixNano())
	fmt.Printf("Timestamp: %d, Node ID: %d, Group ID: %d, Leader ID: %d, Action: Start, Energy Level: %d, (x,y): (%d,%d)\n",
		time.Now().UnixNano(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)

	for node.energyLevel > 0 {
		msg := <-node.msgQueue
		if !node.ParseMsg(msg) {
			break
		}

		if node.isElectionOngoing {
			if node.isElectionCandidate {
				if !node.CheckElection() {
					break
				}
			}
		} else {
			if !node.isLeader {
				if !node.CheckHeartBeat() {
					break
				}
			}
		}

		if !node.ConsumeEnergyToLive() {
			break
		}

		time.Sleep(50 * time.Millisecond) // Adding some delay between simulation steps
	}

	node.isAlive = false
	node.energyLevel = 0 // Set correct energy level in case of a break
	fmt.Printf("Timestamp: %d, Node ID: %d, Group ID: %d, Leader ID: %d, Action: NodeRemoved, Energy Level: %d, (x,y): (%d,%d)\n",
		time.Now().UnixNano(), node.id, node.groupID, node.leader.id, node.energyLevel, node.x, node.y)
}

func (node *Node) String() string {
	return fmt.Sprintf("Node{x=%d, y=%d, isLeader=%t, groupId=%d}", node.x, node.y, node.isLeader, node.groupID)
}
