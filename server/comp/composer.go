package comp

import (
	"errors"
	"fmt"
	"github.com/appcrash/media/server/event"
	"strings"
)

type Composer struct {
	sessionId       string
	gt              *GraphTopology
	messageProvider []MessageProvider
	dispatch        *Dispatch

	namedChannel map[string]chan<- *event.Event
}

func NewSessionComposer(sessionId string) *Composer {
	sc := &Composer{
		sessionId:    sessionId,
		namedChannel: make(map[string]chan<- *event.Event),
	}
	return sc
}

func (c *Composer) ParseGraphDescription(desc string) (err error) {
	gt := newGraphTopology()
	lines := strings.Split(desc, "\n")

	for _, l := range lines {
		if l == "" {
			continue
		}
		gt.parseLine(l)
	}
	if gt.nbParseError > 0 {
		errStr := fmt.Sprintf("there are total %v error in graph description:\n%v", gt.nbParseError, desc)
		err = errors.New(errStr)
		return
	}
	err = gt.topographicalSort()
	c.gt = gt
	return
}

// PrepareNodes create node instances by type, add them to graph, and link them
func (c *Composer) PrepareNodes(graph *event.Graph) (err error) {
	var nodeList []SessionAware
	var nodeIds []*Id
	var dispatch *Dispatch
	nbNode := len(c.gt.sortedNodeList)

	defer func() {
		if err != nil {
			// undo AddNode
			for _, n := range nodeList {
				n.ExitGraph()
			}
			if dispatch != nil {
				dispatch.ExitGraph()
			}
		}
	}()

	// create node instances, collect message providers if any
	for _, n := range c.gt.sortedNodeList {
		n.Props.Set("Name", n.Name)
		sn := MakeSessionNode(n.Type, c.sessionId, n.Props)
		if sn == nil {
			logger.Errorf("unknown node type: %v\n", n.Name)
			err = errors.New("can not make unknown node")
			return
		}
		nodeList = append(nodeList, sn)
		if provider, ok := sn.(MessageProvider); ok {
			c.messageProvider = append(c.messageProvider, provider)
		}

		id := NewId(sn.GetNodeScope(), sn.GetNodeName())
		nodeIds = append(nodeIds, id)
	}

	// add all nodes to graph, create links between them, as nodes are already topographical sorted,
	// for each node, its dependent nodes are in graph when adding it to graph
	for i, n := range nodeList {
		if !graph.AddNode(n) {
			err = errors.New(fmt.Sprintf("failed to add node %v to graph", n.GetNodeName()))
			return
		}
		deps := c.gt.sortedNodeList[i].Deps
		for _, ni := range deps {
			// set pipe end to local session nodes
			if n.SetPipeOut(c.sessionId, ni.Name) != nil {
				err = errors.New(fmt.Sprintf("failed to link %v => %v", n.GetNodeName(), ni.Name))
				return
			}
		}
	}

	// now every node is added to graph and linked
	// create dispatch node which links to all nodes in the session
	ci := make(ConfigItems)
	dispatch = MakeSessionNode(TYPE_DISPATCH, c.sessionId, ci).(*Dispatch)
	dispatch.SetMaxLink(nbNode * 2) // reserved nbNode for dynamical link requests
	if !graph.AddNode(dispatch) {
		err = errors.New("fail to add send-node to graph")
		return
	}
	if err = dispatch.connectTo(nodeIds); err != nil {
		return
	}

	// again, let all nodes reference this dispatch
	for _, n := range nodeList {
		n.SetController(dispatch)
	}
	c.dispatch = dispatch

	// subscribe channels, for all nodes of type pubsub, find the registered channel with same name
	// as specified in pubsub's "channel" property
	if len(c.namedChannel) > 0 {
		for i, n := range c.gt.sortedNodeList {
			if n.Type != TYPE_PUBSUB {
				continue
			}
			if name, ok := n.Props["channel"]; ok {
				chNameList, ok1 := name.(string)
				if !ok1 {
					break
				}
				// pubsub property, for example: channel=a,b,c ...
				for _, chName := range strings.Split(chNameList, ",") {
					if ch, exist := c.namedChannel[chName]; exist {
						nodeList[i].(*PubSubNode).SubscribeChannel(chName, ch)
					}
				}

			}
		}
	}

	return
}

func (c *Composer) GetSortedNodes() (ni []*NodeInfo) {
	return c.gt.sortedNodeList
}

// GetMessageProvider get entry by its name
func (c *Composer) GetMessageProvider(name string) MessageProvider {
	for _, provider := range c.messageProvider {
		if provider.GetName() == name {
			return provider
		}
	}
	return nil
}

func (c *Composer) GetController() Controller {
	return c.dispatch
}

func (c *Composer) RegisterChannel(name string, ch chan<- *event.Event) {
	c.namedChannel[name] = ch
}