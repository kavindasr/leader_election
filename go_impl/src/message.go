package main

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
	standDown bool // Informs the receiver whether to stand down or continue in an election
}

// ElectionResult is a leader election result sent by the newly chosen leader.
type ElectionResult struct {
	sender *Node
}

// Other methods of the message types...
