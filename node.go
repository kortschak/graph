// Copyright ©2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package graph

import (
	"fmt"
)

type Node interface {
	ID() int
	Edges() []Edge
	Degree() int
	Neighbors(EdgeFilter) []Node
	Hops(EdgeFilter) []*Hop
	String() string

	add(Edge)
	drop(Edge)
	dropAll()
	index() int
	setIndex(int)
	setID(int)
}

var _ Node = &node{}

// NodeFilter is a function type used for assessment of nodes during graph traversal.
type NodeFilter func(Node) bool

// A Node is a node in a graph.
type node struct {
	id    int
	i     int
	edges Edges
}

// newNode creates a new *Nodes with ID id. Nodes should only ever exist in the context of a
// graph, so this is not a public function.
func newNode(id int) Node {
	return &node{
		id: id,
	}
}

// A Hop is an edge/node pair where the edge leads to the node from a neighbor.
type Hop struct {
	Edge Edge
	Node Node
}

// ID returns the id of a node.
func (n *node) ID() int {
	return n.id
}

// Edges returns a slice of edges that are incident on the node.
func (n *node) Edges() []Edge {
	return n.edges
}

// Degree returns the number of incident edges on a node. Looped edges are counted at both ends.
func (n *node) Degree() int {
	l := 0
	for _, e := range n.edges {
		if e.Head() == e.Tail() {
			l++
		}
	}
	return l + len(n.edges)
}

// Neighbors returns a slice of nodes that share an edge with the node. Multiply connected nodes are
// repeated in the slice. If the node is n-connected it will be included in the slice, potentially
// repeatedly if there are multiple n-connecting edges.
func (n *node) Neighbors(ef EdgeFilter) []Node {
	var nodes []Node
	for _, e := range n.edges {
		if ef(e) {
			if a := e.Tail(); a == n {
				nodes = append(nodes, e.Head())
			} else {
				nodes = append(nodes, a)
			}
		}
	}
	return nodes
}

// Hops has essentially the same functionality as Neighbors with the exception that the connecting
// edge is also returned.
func (n *node) Hops(ef EdgeFilter) []*Hop {
	var h []*Hop
	for _, e := range n.edges {
		if ef(e) {
			if a := e.Tail(); a == n {
				h = append(h, &Hop{e, e.Head()})
			} else {
				h = append(h, &Hop{e, a})
			}
		}
	}
	return h
}

func (n *node) add(e Edge) { n.edges = append(n.edges, e) }

func (n *node) dropAll() {
	for i := range n.edges {
		n.edges[i] = nil
	}
	n.edges = n.edges[:0]
}

func (n *node) drop(e Edge) {
	for i := 0; i < len(n.edges); {
		if n.edges[i] == e {
			n.edges = n.edges.delFromNode(i)
			return // assumes e has not been added more than once - this should not happen, but we don't check for it
		} else {
			i++
		}
	}
}

func (n *node) setID(id int)   { n.id = id }
func (n *node) setIndex(i int) { n.i = i }
func (n *node) index() int     { return n.i }

func (n *node) String() string {
	return fmt.Sprintf("%d:%v", n.id, n.edges)
}

// Nodes is a collection of nodes.
type Nodes []Node

// BuildUndirected creates a new Undirected graph using nodes and edges specified by the
// set of nodes in the receiver. If edges of nodes in the receiver connect to nodes that are not, these extra nodes
// will be included in the resulting graph. If compact is set to true, edge IDs are chosen to minimize
// space consumption, but breaking edge ID consistency between the new graph and the original.
func (ns Nodes) BuildUndirected(compact bool) (*Undirected, error) {
	seen := make(map[Edge]struct{})
	g := NewUndirected()
	for _, n := range ns {
		g.AddID(n.ID())
		for _, e := range n.Edges() {
			if _, ok := seen[e]; ok {
				continue
			}
			seen[e] = struct{}{}
			u, v := e.Nodes()
			uid, vid := u.ID(), v.ID()
			if uid < 0 || vid < 0 {
				return nil, NodeIDOutOfRange
			}
			g.AddID(uid)
			g.AddID(vid)
			var ne Edge
			if compact {
				ne = g.newEdge(g.nodes[uid], g.nodes[vid], e.Weight(), e.Flags())
			} else {
				ne = g.newEdgeKeepID(e.ID(), g.nodes[uid], g.nodes[vid], e.Weight(), e.Flags())
			}
			g.nodes[uid].add(ne)
			if vid != uid {
				g.nodes[vid].add(ne)
			}
		}
	}

	return g, nil
}

func (ns Nodes) delFromGraph(i int) Nodes {
	ns[i], ns[len(ns)-1] = ns[len(ns)-1], ns[i]
	ns[i].setIndex(i)
	ns[len(ns)-1].setIndex(-1)
	return ns[:len(ns)-1]
}
