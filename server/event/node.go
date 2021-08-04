package event

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// NodeDelegate is delegated to interact with event graph directly
// it is the only way for Node to send events by NodeDelegate
type NodeDelegate struct {
	nodeImpl        Node
	id              string
	ctrlC           chan *Event
	dataC           chan *Event
	graph           *EventGraph
	inExit          atomic.Value
	deliveryTimeout time.Duration // in milliseconds

	invokeMutex sync.Mutex
	// lock-free dlink array with fixed size(maxLink)
	// even though links field and its pointers(*atomic.Value) is not
	// volatile(that means they can be cpu-local cached), it doesn't matter.
	// in fact, all of them are read-only once created,
	// accessing element by atomic.Load always gets the latest value
	// the limitation of max links are the trade-off between performance and
	// flexibility. if links can dynamic grow and event delivery has to atomic
	// read it everytime(instead of cached) because we have to rewrite this variable
	// when enlarging array. in real world, output link number is fixed in most case,
	// so put this limitation doesn't hurt much
	links []*atomic.Value
	// recycled dlink slots are saved here
	freeLinkSlot []int
}

// cope with atomic(Store/Load) that doesn't permit nil value
var nullLink = &dlink{}

const defaultMaxLink = 5
const defaultDataChannelSize = 100
const defaultDeliveryTimeout = 100 * time.Millisecond

func newNodeDelegate(graph *EventGraph, node Node, maxLink int) *NodeDelegate {
	delegate := &NodeDelegate{
		nodeImpl:        node,
		ctrlC:           make(chan *Event),
		graph:           graph,
	}
	delegate.id = node.GetNodeScope() + ":" + node.GetNodeName()
	delegate.inExit.Store(false)
	delegate.links = make([]*atomic.Value, maxLink)
	delegate.freeLinkSlot = make([]int, maxLink)
	for i := 0; i < maxLink; i++ {
		val := new(atomic.Value)
		val.Store(nullLink)
		delegate.links[i] = val
		delegate.freeLinkSlot[i] = maxLink - 1 - i
	}

	dataSize := defaultDataChannelSize
	deliveryTimeout := defaultDeliveryTimeout
	ps := reflect.ValueOf(node)
	elem := ps.Elem()
	if elem.Kind() == reflect.Struct {
		field := elem.FieldByName("dataChannelSize")
		if field.IsValid() {
			switch field.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if ds := int(field.Int()); ds > 0 {
					dataSize = ds
				}
			}
		}
		field = elem.FieldByName("deliveryTimeout")
		if field.IsValid() && field.Type().String() == "time.Duration" {
			to := (*time.Duration)(unsafe.Pointer(field.UnsafeAddr()))
			if *to > 0 {
				deliveryTimeout = *to
			}
		}
	}

	// only buffered channel can satisfy nonblock sending in most case
	delegate.dataC = make(chan *Event,dataSize)
	delegate.deliveryTimeout = deliveryTimeout
	return delegate
}

func (nd *NodeDelegate) getNodeName() string {
	return nd.nodeImpl.GetNodeName()
}

func (nd *NodeDelegate) getNodeScope() string {
	return nd.nodeImpl.GetNodeScope()
}

// unique id in event graph
func (nd *NodeDelegate) getId() string {
	return nd.id
}

func (nd *NodeDelegate) isExiting() bool {
	return nd.inExit.Load().(bool)
}

func (nd *NodeDelegate) setExiting() {
	nd.inExit.Store(true)
}

func (nd *NodeDelegate) receiveCtrl(evt *Event) {
	// sync receive, assume event graph is bug free
	nd.ctrlC <- evt
}

func (nd *NodeDelegate) receiveData(evt *Event, timeoutMs time.Duration) (ok bool) {
	// always try nonblock delivery first, take chance to avoid creating timer
	select {
	case nd.dataC <- evt:
		ok = true
		return
	default:
	}

	// receiver doesn't catch up with me, wait until timeout
	select {
	case nd.dataC <- evt:
		ok = true
	case <-time.After(timeoutMs):
	}
	return
}

func (nd *NodeDelegate) startEventLoop() {
	// buffered channel keep system loop from spinning after node exits
	// 2 = 1(system event loop) + 1(finalize)
	doneC := make(chan int, 2)
	go func(n *NodeDelegate, done chan int) {
		err := nd.systemEventLoop()
		if err != nil {
			// come here means event graph has a serious bug
			fmt.Errorf("[graph]: fatal error")
		}
		// notify user event loop stop
		done <- 0
	}(nd, doneC)
	go func(n *NodeDelegate, done chan int) {
		err := n.userEventLoop(done)
		if err != nil {
			// normal execution should not come here, we are out of loop
			// because either normal exit(err == nil) or exception
			// thrown(err != nil), it means this node is in abnormal state,
			// report to graph and finalize me
			n.finalize(err, done)
		}
	}(nd, doneC)
}

func (nd *NodeDelegate) systemEventLoop() (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			fmt.Errorf("[graph]: bug %v\n", r)
			if err, ok = r.(error); ok {
				return
			}
			err = errors.New("unknown error type in event loop")
		}
	}()

	for {
		evt := <-nd.ctrlC
		nd.handleSystemEvent(evt)
		if evt.cmd == resp_node_exit {
			// node-exit is the last ctrl message in node's lifecycle
			// end this goroutine now
			return
		}
	}
}

func (nd *NodeDelegate) handleSystemEvent(evt *Event) {
	switch evt.cmd {
	case resp_node_add:
		if resp, ok := evt.obj.(*nodeAddResponse); !ok {
			panic(errors.New("[graph]:node add response with wrong event object"))
		} else {
			go func(n *NodeDelegate) {
				n.invokeMutex.Lock()
				defer n.invokeMutex.Unlock()
				n.nodeImpl.OnEnter(n)
			}(resp.delegate)
		}
	case resp_node_exit:
		if _, ok := evt.obj.(*nodeExitResponse); !ok {
			panic(errors.New("[graph]:node exit response with wrong event object"))
		} else {
			go func(n *NodeDelegate) {
				n.invokeMutex.Lock()
				defer n.invokeMutex.Unlock()
				n.nodeImpl.OnExit()
			}(nd)
		}
	case resp_link_up:
		if resp, ok := evt.obj.(*linkUpResponse); !ok {
			panic(errors.New("[graph]:dlink up response with wrong event object"))
		} else {
			scope := resp.scope
			nodeName := resp.nodeName
			if resp.state != 0 {
				go func(n *NodeDelegate, s string, name string) {
					n.invokeMutex.Lock()
					defer n.invokeMutex.Unlock()
					n.nodeImpl.OnLinkUp(-1, s, name)
				}(nd, scope, nodeName)
				return
			}
			// if free list is not empty, recycle the slot
			var newLinkId int
			link := resp.link
			freeLen := len(nd.freeLinkSlot)
			if freeLen > 0 {
				var val *atomic.Value
				newLinkId = nd.freeLinkSlot[freeLen-1]
				nd.freeLinkSlot = nd.freeLinkSlot[:freeLen-1]
				val = nd.links[newLinkId]
				if val == nil {
					val = &atomic.Value{}
					nd.links[newLinkId] = val
				}
				// atomic rewrite old dlink info
				val.Store(link)
			}
			//} else {
			//	var v atomic.Value
			//	v.Store(link)
			//	newLinkId = len(nd.links)
			//	nd.links = append(nd.links, &v)
			//}

			go func(n *NodeDelegate, id int, s string, name string) {
				n.invokeMutex.Lock()
				defer n.invokeMutex.Unlock()
				n.nodeImpl.OnLinkUp(id, s, name)
			}(nd, newLinkId, scope, nodeName)
		}
	case resp_link_down:
		// we receive link_down either toNode exits the graph or we actively requested previously
		if resp, ok := evt.obj.(*linkDownResponse); !ok {
			panic(errors.New("[graph]:dlink down response with wrong event object"))
		} else {
			link := resp.link
			linkId := -1
			for i, v := range nd.links {
				l := v.Load().(*dlink)
				if l == link {
					linkId = i
					break
				}
			}
			if linkId < 0 {
				// wrong link passed to node, TODO: log error
				return
			}
			// put index to free list, set the slot to null link so Delivery would fail
			nd.freeLinkSlot = append(nd.freeLinkSlot, linkId)
			nd.links[linkId].Store(nullLink)
			scope := link.toNode.getNodeScope()
			nodeName := link.toNode.getNodeName()
			go func(n *NodeDelegate, id int, s string, name string) {
				n.invokeMutex.Lock()
				defer n.invokeMutex.Unlock()
				n.nodeImpl.OnLinkDown(id, s, name)
			}(nd, linkId, scope, nodeName)
		}
	default:
		panic(errors.New("[graph]: received unknown command"))
	}
}

func (nd *NodeDelegate) userEventLoop(done chan int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(error); ok {
				return
			}
			err = errors.New("unknown error type in event loop")
		}
	}()

	for {
		select {
		case evt := <-nd.dataC:
			nd.handleUserEvent(evt)
		case <-done:
			// if uncomment this, golang race detector would complain...
			//nd.dataC = nil

			// breakout leaving dataC no receiver, so events can not be delivered now
			// as the delivery method would return false and senders should stop sending
			// more events. however, events buffered in dataC are lost
			return
		}
	}
}

func (nd *NodeDelegate) handleUserEvent(evt *Event) {
	nd.nodeImpl.OnEvent(evt)
}

func (nd *NodeDelegate) finalize(err error, done chan int) {
	// exception handling
	// user code error happened and caught by node delegate, we will
	// 1. notify user event loop to exit, all senders' delivery would fail;
	// 2. request node exit to graph, the node would be notified
	// output links are down and moved out of the graph, the node's
	// system event loop would terminate when node-exit response received
	done <- 0
	nd.RequestNodeExit()
}

/*****************************************************
 ****************** APIs for end user ****************
 *****************************************************/

// RequestLinkUp node request dlink to other node of @param scope and name @param nodeName
// the request is passed to graph, then graph would create the dlink and notify the node
// delegate. the node's OnLinkUp would be invoked with linkId as parameter
// if the requested node doesn't exist, node would be notified with linkId == -1
func (nd *NodeDelegate) RequestLinkUp(scope string, nodeName string) (err error) {
	// graph will give error response if duplication happened
	// i.e. (fromScope,fromName,toScope,toName) quaternion is unique across graph
	if scope == "" || nodeName == "" {
		err = errors.New("wrong dlink-up parameters")
		return
	}
	evt := newLinkUpRequest(nd, scope, nodeName)
	nd.graph.deliveryEvent(evt)
	return
}

// RequestLinkDown node request tearing down an output dlink, and node's
// OnLinkDown would be invoked once successfully tearing down
func (nd *NodeDelegate) RequestLinkDown(linkId int) (err error) {
	// what if linkId >= len(nd.links) when nd.links itself not protected by
	// mutex and not being atomic? nd.links can only increase, so linkId assigned
	// before is always safe to access. however, the value may have been changed
	// due to slot recycling, sanity checking is necessary. a normal-behavior node
	// should never see such case (i.e. pass wrong linkId in)
	if linkId >= len(nd.links) {
		err = errors.New("linkId out of range")
		return
	}
	link := nd.links[linkId].Load().(*dlink)
	if link.fromNode != nd {
		err = errors.New("wrong linkId with different fromNode(not you)")
		return
	}
	evt := newLinkDownRequest(link)
	nd.graph.deliveryEvent(evt)
	return
}

func (nd *NodeDelegate) RequestNodeExit() (err error) {
	if nd.isExiting() {
		err = errors.New("node is already in exiting state")
		return
	}

	// ensure all input/output links are torn down before the node gets
	// out of the graph, send exit request to graph, and graph will handle
	// all of this
	nd.setExiting()
	nd.graph.deliveryEvent(newNodeExitRequest(nd))
	return
}

// DeliveryWithTimeout return true if successfully delivered
func (nd *NodeDelegate) DeliveryWithTimeout(linkId int, evt *Event, timeout time.Duration) bool {
	link := nd.links[linkId].Load().(*dlink)
	if link == nullLink {
		return false
	}
	if link.fromNode != nd {
		// if it is an input dlink
		return false
	}
	//nodeFile.WriteString(link.fromNode.getNodeName() + " ==> " + link.toNode.getNodeName() + "\n")
	return link.toNode.receiveData(evt, timeout)
}

func (nd *NodeDelegate) Delivery(linkId int, evt *Event) bool {
	return nd.DeliveryWithTimeout(linkId, evt, nd.deliveryTimeout)
}
