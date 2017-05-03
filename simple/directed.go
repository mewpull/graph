// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package simple

import (
	"fmt"

	"golang.org/x/tools/container/intsets"

	"github.com/gonum/graph"
)

// DirectedGraph implements a generalized directed graph.
type DirectedGraph struct {
	nodes map[int]graph.Node
	from  map[int]map[int]graph.Edge
	to    map[int]map[int]graph.Edge

	self, absent float64

	freeIDs intsets.Sparse
	usedIDs intsets.Sparse
}

// NewDirectedGraph returns a DirectedGraph with the specified self and absent
// edge weight values.
func NewDirectedGraph(self, absent float64) *DirectedGraph {
	return &DirectedGraph{
		nodes: make(map[int]graph.Node),
		from:  make(map[int]map[int]graph.Edge),
		to:    make(map[int]map[int]graph.Edge),

		self:   self,
		absent: absent,
	}
}

// NewNodeID returns a new unique ID for a node to be added to g. The returned ID does
// not become a valid ID in g until it is added to g.
func (g *DirectedGraph) NewNodeID() graph.NodeID {
	if len(g.nodes) == 0 {
		return 0
	}
	if len(g.nodes) == maxInt {
		panic(fmt.Sprintf("simple: cannot allocate node: no slot"))
	}

	var id int
	if g.freeIDs.Len() != 0 && g.freeIDs.TakeMin(&id) {
		return graph.NodeID(id)
	}
	if id = g.usedIDs.Max(); id < maxInt {
		return graph.NodeID(id + 1)
	}
	for id = 0; id < maxInt; id++ {
		if !g.usedIDs.Has(id) {
			return graph.NodeID(id)
		}
	}
	panic("unreachable")
}

// AddNode adds n to the graph. It panics if the added node ID matches an existing node ID.
func (g *DirectedGraph) AddNode(n graph.Node) {
	id := int(n.ID())
	if _, exists := g.nodes[id]; exists {
		panic(fmt.Sprintf("simple: node ID collision: %d", id))
	}
	g.nodes[id] = n
	g.from[id] = make(map[int]graph.Edge)
	g.to[id] = make(map[int]graph.Edge)

	g.freeIDs.Remove(id)
	g.usedIDs.Insert(id)
}

// RemoveNode removes n from the graph, as well as any edges attached to it. If the node
// is not in the graph it is a no-op.
func (g *DirectedGraph) RemoveNode(n graph.NodeID) {
	id := int(n)
	if _, ok := g.nodes[id]; !ok {
		return
	}
	delete(g.nodes, id)

	for from := range g.from[id] {
		delete(g.to[from], id)
	}
	delete(g.from, id)

	for to := range g.to[id] {
		delete(g.from[to], id)
	}
	delete(g.to, id)

	g.freeIDs.Insert(id)
	g.usedIDs.Remove(id)
}

// SetEdge adds e, an edge from one node to another. If the nodes do not exist, they are added.
// It will panic if the IDs of the e.From and e.To are equal.
func (g *DirectedGraph) SetEdge(e graph.Edge) {
	var (
		from = e.From()
		fid  = from.ID()
		to   = e.To()
		tid  = to.ID()
	)

	if fid == tid {
		panic("simple: adding self edge")
	}

	if !g.Has(fid) {
		g.AddNode(from)
	}
	if !g.Has(tid) {
		g.AddNode(to)
	}

	g.from[int(fid)][int(tid)] = e
	g.to[int(tid)][int(fid)] = e
}

// RemoveEdge removes e from the graph, leaving the terminal nodes. If the edge does not exist
// it is a no-op.
func (g *DirectedGraph) RemoveEdge(e graph.Edge) {
	from, to := e.From(), e.To()
	fid, tid := int(from.ID()), int(to.ID())
	if _, ok := g.nodes[fid]; !ok {
		return
	}
	if _, ok := g.nodes[tid]; !ok {
		return
	}

	delete(g.from[fid], tid)
	delete(g.to[tid], fid)
}

// Node returns the node in the graph with the given ID.
func (g *DirectedGraph) Node(id graph.NodeID) graph.Node {
	return g.nodes[int(id)]
}

// Has reports whether the node exists within the graph.
func (g *DirectedGraph) Has(n graph.NodeID) bool {
	_, ok := g.nodes[int(n)]

	return ok
}

// Nodes returns all the nodes in the graph.
func (g *DirectedGraph) Nodes() []graph.Node {
	nodes := make([]graph.Node, len(g.from))
	i := 0
	for _, n := range g.nodes {
		nodes[i] = n
		i++
	}

	return nodes
}

// Edges returns all the edges in the graph.
func (g *DirectedGraph) Edges() []graph.Edge {
	var edges []graph.Edge
	for _, u := range g.nodes {
		for _, e := range g.from[int(u.ID())] {
			edges = append(edges, e)
		}
	}
	return edges
}

// From returns all nodes in g that can be reached directly from n.
func (g *DirectedGraph) From(n graph.NodeID) []graph.Node {
	fid := int(n)
	if _, ok := g.from[fid]; !ok {
		return nil
	}

	from := make([]graph.Node, len(g.from[fid]))
	i := 0
	for tid := range g.from[fid] {
		from[i] = g.nodes[tid]
		i++
	}

	return from
}

// To returns all nodes in g that can reach directly to n.
func (g *DirectedGraph) To(n graph.NodeID) []graph.Node {
	tid := int(n)
	if _, ok := g.from[tid]; !ok {
		return nil
	}

	to := make([]graph.Node, len(g.to[tid]))
	i := 0
	for fid := range g.to[tid] {
		to[i] = g.nodes[fid]
		i++
	}

	return to
}

// HasEdgeBetween reports whether an edge exists between nodes x and y without
// considering direction.
func (g *DirectedGraph) HasEdgeBetween(x, y graph.NodeID) bool {
	xid := int(x)
	yid := int(y)
	if _, ok := g.nodes[xid]; !ok {
		return false
	}
	if _, ok := g.nodes[yid]; !ok {
		return false
	}
	if _, ok := g.from[xid][yid]; ok {
		return true
	}
	_, ok := g.from[yid][xid]
	return ok
}

// Edge returns the edge from u to v if such an edge exists and nil otherwise.
// The node v must be directly reachable from u as defined by the From method.
func (g *DirectedGraph) Edge(u, v graph.NodeID) graph.Edge {
	uid, vid := int(u), int(v)
	if _, ok := g.nodes[uid]; !ok {
		return nil
	}
	if _, ok := g.nodes[vid]; !ok {
		return nil
	}
	edge, ok := g.from[uid][vid]
	if !ok {
		return nil
	}
	return edge
}

// HasEdgeFromTo reports whether an edge exists in the graph from u to v.
func (g *DirectedGraph) HasEdgeFromTo(u, v graph.NodeID) bool {
	uid, vid := int(u), int(v)
	if _, ok := g.nodes[uid]; !ok {
		return false
	}
	if _, ok := g.nodes[vid]; !ok {
		return false
	}
	if _, ok := g.from[uid][vid]; !ok {
		return false
	}
	return true
}

// Weight returns the weight for the edge between x and y if Edge(x, y) returns a non-nil Edge.
// If x and y are the same node or there is no joining edge between the two nodes the weight
// value returned is either the graph's absent or self value. Weight returns true if an edge
// exists between x and y or if x and y have the same ID, false otherwise.
func (g *DirectedGraph) Weight(x, y graph.NodeID) (w float64, ok bool) {
	xid := int(x)
	yid := int(y)
	if xid == yid {
		return g.self, true
	}
	if to, ok := g.from[xid]; ok {
		if e, ok := to[yid]; ok {
			return e.Weight(), true
		}
	}
	return g.absent, false
}

// Degree returns the in+out degree of n in g.
func (g *DirectedGraph) Degree(n graph.NodeID) int {
	id := int(n)
	if _, ok := g.nodes[id]; !ok {
		return 0
	}

	return len(g.from[id]) + len(g.to[id])
}
