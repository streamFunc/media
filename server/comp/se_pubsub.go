package comp

import (
	"errors"
	"fmt"
	"github.com/appcrash/media/server/event"
	"sync"
	"time"
)

// this node accepts input data (from pub, one or many), make multiple copy of it then send them to
// all subscriber (to sub). subscriber can be added or removed dynamically by input commands or api.
// the node will actively remove a subscriber once event is not successfully delivered to it.

// for input, one of :
// 1. other node send data message to pubsub node (inter-node communication)
// 2. call this node's Publish method (feed event to event graph)
//
// for output(subscriber), one of:
// 1. other node that receives event from pubsub node (inter-node communication)
// 2. provider a channel of type (chan<- *event.Event) to which pubsub deliveries (consume event from event graph),
// the channel must be buffered channel, i.e. cap(c) != 0, and in the SAME session of this node, so communication
// across sessions must take inter-node measures
//
// PubSub can have only one input and one output, then it becomes like 'tee' in linux command line
// which read from stdin and write to stdout. so PubSub can be a bridge between event graph and outside world
//
// |outside| ------> pubsub -----> event graph
//            feed
//
// event graph -----> pubsub -----> |outside|
//                          consume

const PUBSUB_DEFAULT_DELIVERY_TIMEOUT = 20 * time.Millisecond

const (
	psSubscribeTypeNode = iota
	psSubscribeTypeChannel
)

type psSubscriberInfo struct {
	subType int
	linkId  int                 // if subscriber is a node
	channel chan<- *event.Event // if subscriber is a chan
	name    string
}

type PubSubNode struct {
	SessionNode
	event.NodeProperty

	mutex       sync.Mutex
	subscribers []*psSubscriberInfo
}

func (p *PubSubNode) OnEvent(evt *event.Event) {
	obj := evt.GetObj()
	if obj == nil {
		return
	}
	switch evt.GetCmd() {
	case DATA_OUTPUT:
		if c, ok := obj.(Cloneable); ok {
			p.Publish(c)
		}
	case CTRL_CALL:
		if msg, ok := obj.(*CtrlMessage); ok {
			p.handleCall(msg)
		}
	}
}

func (p *PubSubNode) OnLinkDown(_ int, scope string, nodeName string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if index, si := p.findNodeSubscriber(scope, nodeName); si != nil {
		// a node subscriber is down, just remove it from subscribers
		p.deleteSubscriber(index)
	}
}

// OnExit close all channel subscribers
func (p *PubSubNode) OnExit() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, s := range p.subscribers {
		if s.subType == psSubscribeTypeChannel {
			if s.channel != nil {
				close(s.channel)
				s.channel = nil
			}
		}
	}
}

// SetPipeOut overrides default session node's behaviour, it allows multiple pipes simultaneously
// (that's what "pubsub" stands for) instead of only one data output pipe
func (p *PubSubNode) SetPipeOut(session, name string) error {
	if p.delegate == nil {
		return errors.New("delegate not ready when set pipe")
	}
	return p.SubscribeNode(session, name)
}

//------------------------------- api & implementation --------------------------------------

func (p *PubSubNode) Publish(obj Cloneable) {
	var subscribers []*psSubscriberInfo
	if obj == nil {
		return
	}
	p.mutex.Lock()
	subscribers = p.subscribers // copy the array, release lock
	p.mutex.Unlock()

	// publish message to all subscribers
	// for node: delivery timeout by field 'deliveryTimeout'
	// for channel: nonblock sending without timeout, so must use buffered channel to avoid losing message
	for _, s := range subscribers {
		switch s.subType {
		case psSubscribeTypeNode:
			if s.linkId < 0 {
				continue
			}
			evt := event.NewEvent(DATA_OUTPUT, obj.Clone())
			p.delegate.Deliver(s.linkId, evt)
		case psSubscribeTypeChannel:
			if s.channel == nil {
				continue
			}
			evt := event.NewEvent(DATA_OUTPUT, obj.Clone())
			select {
			case s.channel <- evt:
			default:
			}
		}
	}
}

func newPubSubNode() SessionAware {
	node := new(PubSubNode)
	node.Name = TYPE_PUBSUB
	node.SetDeliveryTimeout(PUBSUB_DEFAULT_DELIVERY_TIMEOUT)
	return node
}

func psNewNodeSubscriber(scope, nodeName string, linkId int) *psSubscriberInfo {
	name := psMakeNodeName(scope, nodeName)
	si := new(psSubscriberInfo)
	si.subType = psSubscribeTypeNode
	si.name = name
	si.linkId = linkId
	return si
}

func psNewChannelSubscriber(chName string, c chan<- *event.Event) *psSubscriberInfo {
	name := psMakeChannelName(chName)
	si := new(psSubscriberInfo)
	si.subType = psSubscribeTypeChannel
	si.name = name
	si.channel = c
	si.linkId = -1
	return si
}

func psMakeNodeName(scope, name string) string {
	return fmt.Sprintf("node_%v_%v", scope, name)
}

func psMakeChannelName(chName string) string {
	return "chan_" + chName
}

func (p *PubSubNode) findSubInfo(name string) (index int, si *psSubscriberInfo) {
	if name == "" {
		return -1, nil
	}
	for i, n := range p.subscribers {
		if n.name == name {
			return i, n
		}
	}
	return -1, nil
}

func (p *PubSubNode) findNodeSubscriber(scope, name string) (index int, si *psSubscriberInfo) {
	nodeName := psMakeNodeName(scope, name)
	return p.findSubInfo(nodeName)
}

func (p *PubSubNode) findChannelSubscriber(chName string) (index int, si *psSubscriberInfo) {
	name := psMakeChannelName(chName)
	return p.findSubInfo(name)
}

func (p *PubSubNode) deleteSubscriber(index int) {
	p.mutex.Lock()
	siLen := len(p.subscribers)
	p.subscribers[index] = p.subscribers[siLen-1]
	p.subscribers = p.subscribers[:siLen-1]
	p.mutex.Unlock()
}

func (p *PubSubNode) handleCall(msg *CtrlMessage) {
	m := msg.M
	if ml := len(m); ml > 0 {
		switch m[0] {
		case "conn":
			if ml == 3 {
				toSession, toName := m[1], m[2]
				if err := p.SetPipeOut(toSession, toName); err == nil {
					msg.C <- WithOk()
				} else {
					msg.C <- WithError()
				}
			}
		}
	}
}

// SubscribeNode add a node as a subscriber of this pubsub node
func (p *PubSubNode) SubscribeNode(scope, name string) error {
	if _, s := p.findNodeSubscriber(scope, name); s != nil {
		return errors.New(fmt.Sprintf("node %v:%v is already a subscriber", scope, name))
	}
	if linkId := p.delegate.RequestLinkUp(scope, name); linkId >= 0 {
		si := psNewNodeSubscriber(scope, name, linkId)
		p.mutex.Lock()
		p.subscribers = append(p.subscribers, si)
		p.mutex.Unlock()
		return nil
	} else {
		return errors.New(fmt.Sprintf("node(%v:%v) subscribes failed", scope, name))
	}
}

// UnsubscribeNode remove a node subscriber from this pubsub node
func (p *PubSubNode) UnsubscribeNode(scope, name string) error {
	if index, si := p.findNodeSubscriber(scope, name); si == nil {
		return nil
	} else {
		// delete the subscriber now
		p.deleteSubscriber(index)
		return p.delegate.RequestLinkDown(si.linkId)
	}
}

// SubscribeChannel add a channel as a subscriber of this pubsub node
func (p *PubSubNode) SubscribeChannel(chName string, c chan<- *event.Event) error {
	if _, s := p.findChannelSubscriber(chName); s != nil {
		return errors.New(fmt.Sprintf("channel %v is already subscribed", chName))
	}
	if cap(c) == 0 {
		// must be buffered channel
		return errors.New("must use buffered channel to subscribe")
	}
	si := psNewChannelSubscriber(chName, c)
	p.mutex.Lock()
	p.subscribers = append(p.subscribers, si)
	p.mutex.Unlock()
	return nil
}

// UnsubscribeChannel remove a channel subscriber with given name from this pubsub node
func (p *PubSubNode) UnsubscribeChannel(chName string) error {
	if index, si := p.findChannelSubscriber(chName); si == nil {
		return errors.New(fmt.Sprintf("channel %v is not subscribed, so unsubscribe fails", chName))
	} else {
		p.deleteSubscriber(index)
		return nil
	}
}