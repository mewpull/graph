// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package simple

import (
	"fmt"

	"golang.org/x/tools/container/intsets"

	"github.com/gonum/graph"
)

// UndirectedGraph implements a generalized undirected graph.
type UndirectedGraph struct {
	nodes map[int]graph.Node
	edges map[int]map[int]graph.Edge

	self, absent float64

	freeIDs intsets.Sparse
	usedIDs intsets.Sparse
}

// NewUndirectedGraph returns an UndirectedGraph with the specified self and absent
// edge weight values.
func NewUndirectedGraph(self, absent float64) *UndirectedGraph {
	return &UndirectedGraph{
		nodes: make(map[int]graph.Node),
		edges: make(map[int]map[int]graph.Edge),

		self:   self,
		absent: absent,
	}
}

// NewNodeID returns a new unique ID for a node to be added to g. The returned ID does
// not become a valid ID in g until it is added to g.
func (g *UndirectedGraph) NewNodeID() int64 {
	if len(g.nodes) == 0 {
		return 0
	}
	if len(g.nodes) == maxInt {
		panic(fmt.Sprintf("simple: cannot allocate node: no slot"))
	}

	var id int
	if g.freeIDs.Len() != 0 && g.freeIDs.TakeMin(&id) {
		return int64(id)
	}
	if id = g.usedIDs.Max(); id < maxInt {
		return int64(id + 1)
	}
	for id = 0; id < maxInt; id++ {
		if !g.usedIDs.Has(id) {
			return int64(id)
		}
	}
	panic("unreachable")
}

// AddNode adds n to the graph. It panics if the added node ID matches an existing node ID.
func (g *UndirectedGraph) AddNode(n graph.Node) {
	id := int(n.ID())
	if _, exists := g.nodes[id]; exists {
		panic(fmt.Sprintf("simple: node ID collision: %d", n.ID()))
	}
	g.nodes[id] = n
	g.edges[id] = make(map[int]graph.Edge)

	g.freeIDs.Remove(id)
	g.usedIDs.Insert(id)
}

// RemoveNode removes the node from the graph, as well as any edges attached to it. If the node
// is not in the graph it is a no-op.
func (g *UndirectedGraph) RemoveNode(id int64) {
	n := int(id)
	if _, ok := g.nodes[n]; !ok {
		return
	}
	delete(g.nodes, n)

	for from := range g.edges[n] {
		delete(g.edges[from], n)
	}
	delete(g.edges, n)

	g.freeIDs.Insert(n)
	g.usedIDs.Remove(n)

}

// SetEdge adds e, an edge from one node to another. If the nodes do not exist, they are added.
// It will panic if the IDs of the e.From and e.To are equal.
func (g *UndirectedGraph) SetEdge(e graph.Edge) {
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

	g.edges[int(fid)][int(tid)] = e
	g.edges[int(tid)][int(fid)] = e
}

// RemoveEdge removes e from the graph, leaving the terminal nodes. If the edge does not exist
// it is a no-op.
func (g *UndirectedGraph) RemoveEdge(e graph.Edge) {
	from, to := e.From(), e.To()
	fid, tid := int(from.ID()), int(to.ID())
	if _, ok := g.nodes[fid]; !ok {
		return
	}
	if _, ok := g.nodes[tid]; !ok {
		return
	}

	delete(g.edges[fid], tid)
	delete(g.edges[tid], fid)
}

// Node returns the node in the graph with the given ID.
func (g *UndirectedGraph) Node(id int64) graph.Node {
	return g.nodes[int(id)]
}

// Has reports whether the node exists within the graph.
func (g *UndirectedGraph) Has(n int64) bool {
	_, ok := g.nodes[int(n)]
	return ok
}

// Nodes returns all the nodes in the graph.
func (g *UndirectedGraph) Nodes() []graph.Node {
	nodes := make([]graph.Node, len(g.nodes))
	i := 0
	for _, n := range g.nodes {
		nodes[i] = n
		i++
	}

	return nodes
}

// Edges returns all the edges in the graph.
func (g *UndirectedGraph) Edges() []graph.Edge {
	var edges []graph.Edge

	seen := make(map[[2]int]struct{})
	for _, u := range g.edges {
		for _, e := range u {
			uid := int(e.From().ID())
			vid := int(e.To().ID())
			if _, ok := seen[[2]int{uid, vid}]; ok {
				continue
			}
			seen[[2]int{uid, vid}] = struct{}{}
			seen[[2]int{vid, uid}] = struct{}{}
			edges = append(edges, e)
		}
	}

	return edges
}

// From returns all nodes in g that can be reached directly from n.
func (g *UndirectedGraph) From(n int64) []graph.Node {
	if !g.Has(n) {
		return nil
	}

	nodes := make([]graph.Node, len(g.edges[int(n)]))
	i := 0
	for from := range g.edges[int(n)] {
		nodes[i] = g.nodes[from]
		i++
	}

	return nodes
}

// HasEdgeBetween reports whether an edge exists between nodes x and y.
func (g *UndirectedGraph) HasEdgeBetween(x, y int64) bool {
	_, ok := g.edges[int(x)][int(y)]
	return ok
}

// Edge returns the edge from u to v if such an edge exists and nil otherwise.
// The node v must be directly reachable from u as defined by the From method.
func (g *UndirectedGraph) Edge(u, v int64) graph.Edge {
	return g.EdgeBetween(u, v)
}

// EdgeBetween returns the edge between nodes x and y.
func (g *UndirectedGraph) EdgeBetween(x, y int64) graph.Edge {
	// We don't need to check if neigh exists because
	// it's implicit in the edges access.
	if !g.Has(x) {
		return nil
	}

	return g.edges[int(x)][int(y)]
}

// Weight returns the weight for the edge between x and y if Edge(x, y) returns a non-nil Edge.
// If x and y are the same node or there is no joining edge between the two nodes the weight
// value returned is either the graph's absent or self value. Weight returns true if an edge
// exists between x and y or if x and y have the same ID, false otherwise.
func (g *UndirectedGraph) Weight(x, y int64) (w float64, ok bool) {
	if x == y {
		return g.self, true
	}
	if n, ok := g.edges[int(x)]; ok {
		if e, ok := n[int(y)]; ok {
			return e.Weight(), true
		}
	}
	return g.absent, false
}

// Degree returns the degree of n in g.
func (g *UndirectedGraph) Degree(n int64) int {
	if _, ok := g.nodes[int(n)]; !ok {
		return 0
	}

	return len(g.edges[int(n)])
}
