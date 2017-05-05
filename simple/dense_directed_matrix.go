// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package simple

import (
	"sort"

	"github.com/gonum/graph"
	"github.com/gonum/graph/internal/ordered"
	"github.com/gonum/matrix/mat64"
)

// DirectedMatrix represents a directed graph using an adjacency
// matrix such that all IDs are in a contiguous block from 0 to n-1.
// Edges are stored implicitly as an edge weight, so edges stored in
// the graph are not recoverable.
type DirectedMatrix struct {
	mat   *mat64.Dense
	nodes []graph.Node

	self   float64
	absent float64
}

// NewDirectedMatrix creates a directed dense graph with n nodes.
// All edges are initialized with the weight given by init. The self parameter
// specifies the cost of self connection, and absent specifies the weight
// returned for absent edges.
func NewDirectedMatrix(n int, init, self, absent float64) *DirectedMatrix {
	mat := make([]float64, n*n)
	if init != 0 {
		for i := range mat {
			mat[i] = init
		}
	}
	for i := 0; i < len(mat); i += n + 1 {
		mat[i] = self
	}
	return &DirectedMatrix{
		mat:    mat64.NewDense(n, n, mat),
		self:   self,
		absent: absent,
	}
}

// NewDirectedMatrixFrom creates a directed dense graph with the given nodes.
// The IDs of the nodes must be contiguous from 0 to len(nodes)-1, but may
// be in any order. If IDs are not contiguous NewDirectedMatrixFrom will panic.
// All edges are initialized with the weight given by init. The self parameter
// specifies the cost of self connection, and absent specifies the weight
// returned for absent edges.
func NewDirectedMatrixFrom(nodes []graph.Node, init, self, absent float64) *DirectedMatrix {
	sort.Sort(ordered.ByID(nodes))
	for i, n := range nodes {
		if i != int(n.ID()) {
			panic("simple: non-contiguous node IDs")
		}
	}
	g := NewDirectedMatrix(len(nodes), init, self, absent)
	g.nodes = nodes
	return g
}

// Node returns the node in the graph with the given ID.
func (g *DirectedMatrix) Node(id int64) graph.Node {
	if !g.has(id) {
		return nil
	}
	if g.nodes == nil {
		return Node(id)
	}
	return g.nodes[id]
}

// Has reports whether the node exists within the graph.
func (g *DirectedMatrix) Has(n int64) bool {
	return g.has(n)
}

// has reports whether the node exists within the graph.
func (g *DirectedMatrix) has(id int64) bool {
	r, _ := g.mat.Dims()
	return 0 <= int(id) && int(id) < r
}

// Nodes returns all the nodes in the graph.
func (g *DirectedMatrix) Nodes() []graph.Node {
	if g.nodes != nil {
		nodes := make([]graph.Node, len(g.nodes))
		copy(nodes, g.nodes)
		return nodes
	}
	r, _ := g.mat.Dims()
	nodes := make([]graph.Node, r)
	for i := 0; i < r; i++ {
		nodes[i] = Node(i)
	}
	return nodes
}

// Edges returns all the edges in the graph.
func (g *DirectedMatrix) Edges() []graph.Edge {
	var edges []graph.Edge
	r, _ := g.mat.Dims()
	for i := 0; i < r; i++ {
		for j := 0; j < r; j++ {
			if i == j {
				continue
			}
			if w := g.mat.At(i, j); !isSame(w, g.absent) {
				edges = append(edges, Edge{F: g.Node(int64(i)), T: g.Node(int64(j)), W: w})
			}
		}
	}
	return edges
}

// From returns all nodes in g that can be reached directly from the node.
func (g *DirectedMatrix) From(id int64) []graph.Node {
	if !g.has(id) {
		return nil
	}
	var neighbors []graph.Node
	i := int(id)
	_, c := g.mat.Dims()
	for j := 0; j < c; j++ {
		if j == i {
			continue
		}
		if !isSame(g.mat.At(i, j), g.absent) {
			neighbors = append(neighbors, g.Node(int64(j)))
		}
	}
	return neighbors
}

// To returns all nodes in g that can reach directly to the node.
func (g *DirectedMatrix) To(id int64) []graph.Node {
	if !g.has(id) {
		return nil
	}
	var neighbors []graph.Node
	j := int(id)
	r, _ := g.mat.Dims()
	for i := 0; i < r; i++ {
		if i == j {
			continue
		}
		if !isSame(g.mat.At(i, j), g.absent) {
			neighbors = append(neighbors, g.Node(int64(i)))
		}
	}
	return neighbors
}

// HasEdgeBetween reports whether an edge exists between nodes x and y without
// considering direction.
func (g *DirectedMatrix) HasEdgeBetween(x, y int64) bool {
	i := int(x)
	if !g.has(x) {
		return false
	}
	j := int(y)
	if !g.has(y) {
		return false
	}
	return x != y && (!isSame(g.mat.At(i, j), g.absent) || !isSame(g.mat.At(j, i), g.absent))
}

// Edge returns the edge from u to v if such an edge exists and nil otherwise.
// The node v must be directly reachable from u as defined by the From method.
func (g *DirectedMatrix) Edge(u, v int64) graph.Edge {
	if g.HasEdgeFromTo(u, v) {
		i, j := int(u), int(v)
		return Edge{F: g.Node(u), T: g.Node(v), W: g.mat.At(i, j)}
	}
	return nil
}

// HasEdgeFromTo reports whether an edge exists in the graph from u to v.
func (g *DirectedMatrix) HasEdgeFromTo(u, v int64) bool {
	if !g.has(u) {
		return false
	}
	if !g.has(v) {
		return false
	}
	i, j := int(u), int(v)
	return u != v && !isSame(g.mat.At(i, j), g.absent)
}

// Weight returns the weight for the edge between x and y if Edge(x, y) returns a non-nil Edge.
// If x and y are the same node or there is no joining edge between the two nodes the weight
// value returned is either the graph's absent or self value. Weight returns true if an edge
// exists between x and y or if x and y have the same ID, false otherwise.
func (g *DirectedMatrix) Weight(x, y int64) (w float64, ok bool) {
	if x == y {
		return g.self, true
	}
	if g.has(x) && g.has(y) {
		i, j := int(x), int(y)
		return g.mat.At(i, j), true
	}
	return g.absent, false
}

// SetEdge sets e, an edge from one node to another. If the ends of the edge are not in g
// or the edge is a self loop, SetEdge panics.
func (g *DirectedMatrix) SetEdge(e graph.Edge) {
	fid := e.From().ID()
	tid := e.To().ID()
	if fid == tid {
		panic("simple: set illegal edge")
	}
	i, j := int(fid), int(tid)
	g.mat.Set(i, j, e.Weight())
}

// RemoveEdge removes e from the graph, leaving the terminal nodes. If the edge does not exist
// it is a no-op.
func (g *DirectedMatrix) RemoveEdge(e graph.Edge) {
	fid := e.From().ID()
	if !g.has(fid) {
		return
	}
	tid := e.To().ID()
	if !g.has(tid) {
		return
	}
	i, j := int(fid), int(tid)
	g.mat.Set(i, j, g.absent)
}

// Degree returns the in+out degree of the node in g.
func (g *DirectedMatrix) Degree(id int64) int {
	var deg int
	j := int(id)
	r, c := g.mat.Dims()
	for i := 0; i < r; i++ {
		if i == j {
			continue
		}
		if !isSame(g.mat.At(j, i), g.absent) {
			deg++
		}
	}
	for i := 0; i < c; i++ {
		if i == j {
			continue
		}
		if !isSame(g.mat.At(i, j), g.absent) {
			deg++
		}
	}
	return deg
}

// Matrix returns the mat64.Matrix representation of the graph. The orientation
// of the matrix is such that the matrix entry at G_{ij} is the weight of the edge
// from node i to node j.
func (g *DirectedMatrix) Matrix() mat64.Matrix {
	// Prevent alteration of dimensions of the returned matrix.
	m := *g.mat
	return &m
}
