package election

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"dfs-backend/dfs/common"

	"github.com/google/uuid"
)

type NodeCommunicator interface {
	GetNodeInfo() *common.NodeInfo
	GetID() uuid.UUID
	GetPriority() int
	GetRole() common.NodeRole
	SetRole(role common.NodeRole)
	GetPeers() []common.NodeInfo
	UpdatePeerRole(peerID uuid.UUID, role common.NodeRole)
	RemovePeer(peerID uuid.UUID)
	SendMessage(ip string, port int, data common.Message) error
	BroadcastMessage(data common.Message) error
}

const (
	coordinatorWaitTimeout = 500 * time.Millisecond
	electionTimeout        = 3 * time.Second
)

type BullyElector struct {
	node NodeCommunicator

	electionInProgress bool
	electionMutex      sync.Mutex

	okChan  chan bool
	okMutex sync.Mutex

	leaderCancelChan  chan struct{}
	leaderCancelMutex sync.Mutex

	stopChan chan struct{}

	peerMutex sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	logger *log.Logger

	coordinatorDeadChan chan uuid.UUID
}

func NewBullyElector(node NodeCommunicator) *BullyElector {
	return &BullyElector{
		node:                node,
		stopChan:            make(chan struct{}),
		coordinatorDeadChan: make(chan uuid.UUID, 1),
		logger:              log.New(log.Writer(), "[BullyElector] ", log.LstdFlags),
	}
}

func (b *BullyElector) Start(ctx context.Context) error {
	b.ctx, b.cancel = context.WithCancel(ctx)
	go b.monitorLeader()
	return nil
}

func (b *BullyElector) Stop() error {
	close(b.stopChan)
	if b.cancel != nil {
		b.cancel()
	}
	return nil
}

func (b *BullyElector) StartElection(ctx context.Context) error {
	b.electionMutex.Lock()
	if b.electionInProgress {
		b.logger.Printf("Election already in progress, skipping")
		b.electionMutex.Unlock()
		return nil
	}
	b.logger.Printf("Starting election...")
	b.electionInProgress = true
	defer func() {
		b.electionMutex.Lock()
		b.electionInProgress = false
		b.electionMutex.Unlock()
	}()
	b.electionMutex.Unlock()

	nodeInfo := b.node.GetNodeInfo()
	payload, _ := json.Marshal(nodeInfo)

	b.peerMutex.RLock()
	peers := append([]common.NodeInfo{}, b.node.GetPeers()...)
	b.peerMutex.RUnlock()

	okChan := make(chan bool, 1)
	b.okMutex.Lock()
	if b.okChan != nil {
		close(b.okChan)
	}
	b.okChan = okChan
	b.okMutex.Unlock()

	higherPriorityExists := false
	for _, peer := range peers {
		if peer.Priority > nodeInfo.Priority {
			higherPriorityExists = true
			electionMsg := common.Message{
				Type:    common.MessageElection,
				From:    nodeInfo.ID,
				Payload: payload,
			}
			go b.node.SendMessage(peer.IP, peer.Port, electionMsg)
		}
	}

	if !higherPriorityExists {
		b.logger.Printf("No higher priority nodes found, declaring self as coordinator")
		b.becomeLeader()
		return nil
	}

	select {
	case <-okChan:
		b.logger.Printf("Received OK from higher priority node, stopping election")
	case <-time.After(electionTimeout):
		b.logger.Printf("No OK messages received, declaring self as coordinator")
		b.becomeLeader()
	case <-ctx.Done():
		return nil
	}

	b.okMutex.Lock()
	b.okChan = nil
	b.okMutex.Unlock()

	return nil
}

func (b *BullyElector) becomeLeader() {
	b.logger.Printf("Becoming leader...")

	cancelChan := make(chan struct{}, 1)
	b.leaderCancelMutex.Lock()
	b.leaderCancelChan = cancelChan
	b.leaderCancelMutex.Unlock()

	defer func() {
		b.leaderCancelMutex.Lock()
		b.leaderCancelChan = nil
		b.leaderCancelMutex.Unlock()
	}()

	msg := common.Message{
		Type: common.MessageCoordinator,
		From: b.node.GetID(),
	}
	b.node.BroadcastMessage(msg)

	select {
	case <-cancelChan:
		b.logger.Printf("Received higher priority coordinator, aborting")
		return
	case <-time.After(coordinatorWaitTimeout):
		b.logger.Printf("No higher priority coordinator, becoming leader")
	}

	b.node.SetRole(common.RoleMaster)
	b.logger.Printf("Now acting as leader")
}

func (b *BullyElector) HandleMessage(ctx context.Context, msg common.Message) error {
	switch msg.Type {
	case common.MessageElection:
		return b.handleElection(ctx, msg)
	case common.MessageOK:
		return b.handleOK(msg)
	case common.MessageCoordinator:
		return b.handleCoordinator(msg)
	}
	return nil
}

func (b *BullyElector) NotifyCoordinatorDead(coordinatorID uuid.UUID) {
	select {
	case b.coordinatorDeadChan <- coordinatorID:
	default:
		// Channel full, election already in progress
	}
}

func (b *BullyElector) handleElection(ctx context.Context, msg common.Message) error {
	b.logger.Printf("Election message received from %s", msg.From)

	var nodeInfo common.NodeInfo
	if err := json.Unmarshal(msg.Payload, &nodeInfo); err != nil {
		b.logger.Printf("Failed to unmarshal heartbeat payload: %v", err)
		return err
	}

	if b.node.GetPriority() > nodeInfo.Priority {
		msgOk := common.Message{
			Type: common.MessageOK,
			From: b.node.GetID(),
		}
		go b.node.SendMessage(nodeInfo.IP, nodeInfo.Port, msgOk)

		b.electionMutex.Lock()
		inProgress := b.electionInProgress
		b.electionMutex.Unlock()
		if !inProgress && b.node.GetRole() != common.RoleMaster {
			go b.StartElection(ctx)
		}
	}

	return nil
}

func (b *BullyElector) handleOK(msg common.Message) error {
	if msg.From == b.node.GetID() {
		return nil
	}
	b.okMutex.Lock()
	okChan := b.okChan
	b.okMutex.Unlock()
	if okChan != nil {
		select {
		case okChan <- true:
			b.logger.Printf("OK message received from %s, stopping election", msg.From)
		default:
		}
	}
	return nil
}

func (b *BullyElector) handleCoordinator(msg common.Message) error {
	b.logger.Printf("Coordinator message received from %s", msg.From)

	senderPriority := -1
	for _, peer := range b.node.GetPeers() {
		if peer.ID == msg.From {
			senderPriority = peer.Priority
			break
		}
	}

	myPriority := b.node.GetPriority()

	if senderPriority != -1 && senderPriority < myPriority {
		b.logger.Printf("Ignoring coordinator from lower priority node %s (their: %d, mine: %d)",
			msg.From, senderPriority, myPriority)
		if b.node.GetRole() == common.RoleMaster {
			reMsg := common.Message{
				Type: common.MessageCoordinator,
				From: b.node.GetID(),
			}
			go b.node.BroadcastMessage(reMsg)
		}
		return nil
	}

	b.electionMutex.Lock()
	b.electionInProgress = false
	b.electionMutex.Unlock()

	b.leaderCancelMutex.Lock()
	if b.leaderCancelChan != nil {
		select {
		case b.leaderCancelChan <- struct{}{}:
		default:
		}
	}
	b.leaderCancelMutex.Unlock()

	b.okMutex.Lock()
	if b.okChan != nil {
		select {
		case b.okChan <- true:
		default:
		}
	}
	b.okMutex.Unlock()

	b.node.UpdatePeerRole(msg.From, common.RoleMaster)

	for _, peer := range b.node.GetPeers() {
		if peer.ID != msg.From && peer.Role == common.RoleMaster {
			b.node.UpdatePeerRole(peer.ID, common.RoleStorage)
		}
	}

	if b.node.GetRole() == common.RoleMaster {
		b.logger.Printf("Demoting self - coordinator %s has priority %d (mine: %d)",
			msg.From, senderPriority, myPriority)
		b.node.SetRole(common.RoleStorage)
	}

	return nil
}

func (b *BullyElector) monitorLeader() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-b.stopChan:
			return

		case deadCoordinatorID := <-b.coordinatorDeadChan:
			b.logger.Printf("Notified of dead coordinator %s, starting election", deadCoordinatorID)
			b.node.RemovePeer(deadCoordinatorID)
			b.StartElection(b.ctx)
		case <-ticker.C:
			now := time.Now()
			needsElection := false
			deadCoordinatorID := uuid.Nil
			b.peerMutex.RLock()
			for _, peer := range b.node.GetPeers() {
				if peer.Role == common.RoleMaster {
					if now.Sub(peer.LastHeartbeat) > 15*time.Second {
						needsElection = true
						deadCoordinatorID = peer.ID
						b.logger.Printf("Coordinator %s is unresponsive, starting election", peer.ID)
						break
					}
				}
			}
			b.peerMutex.RUnlock()

			if needsElection {
				b.node.RemovePeer(deadCoordinatorID)
				b.StartElection(b.ctx)
			}
		}
	}
}
