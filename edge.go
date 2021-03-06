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
	"errors"
	"fmt"
)

var alreadyConnected = errors.New("graph: edge already fully connected")

type Edge interface {
	ID() int
	Weight() float64
	SetWeight(float64)
	Nodes() (u, v Node)
	Head() Node
	Tail() Node
	Flags() EdgeFlags
	SetFlags(EdgeFlags)
	String() string

	index() int
	setIndex(int)
	setID(int)
	join(u, v Node)
	disconnect(Node)
	reconnect(dst, src Node)
}

var _ Edge = &edge{}

// EdgeFilter is a function type used for assessment of edges during graph traversal. 
type EdgeFilter func(Edge) bool

// EdgeFlags is a type that can be used to arbitrarily alter the behavior of edges.
type EdgeFlags uint32

const (
	EdgeCut EdgeFlags = 1 << iota // Set and use this flag to prevent traversal of temporarily cut edges.
)

// An edge is an edge in a graph.
type edge struct {
	id     int
	i      int
	u, v   Node
	weight float64
	flags  EdgeFlags
}

// NewEdge returns a new Edge.
func NewEdge() Edge {
	return &edge{}
}

// newEdge returns a new edge.
func newEdge(id, i int, u, v Node, w float64, f EdgeFlags) Edge {
	return &edge{id: id, i: i, u: u, v: v, weight: w, flags: f}
}

// ID returns the id of the edge.
func (e *edge) ID() int {
	return e.id
}

func (e *edge) setID(id int) {
	e.id = id
}

// Index returns the index of the edge in the compact edge list of the graph. The value returned
// cannot be reliably used after an edge deletion.
func (e *edge) index() int {
	return e.i
}

func (e *edge) setIndex(i int) {
	e.i = i
}

// Nodes returns the two nodes, u and v, that are joined by the edge.
func (e *edge) Nodes() (u, v Node) {
	return e.u, e.v
}

// Head returns the first node of an edge's node pair.
func (e *edge) Head() Node {
	return e.v
}

// Tail returns the second node of an edge's node pair.
func (e *edge) Tail() Node {
	return e.u
}

// Weight returns the weight of the edge.
func (e *edge) Weight() float64 {
	return e.weight
}

// SetWeight sets the weight of the edge to w.
func (e *edge) SetWeight(w float64) {
	e.weight = w
}

// Flags returns the flags value for the edge. One flag is currently defined, EdgeCut.
func (e *edge) Flags() EdgeFlags {
	return e.flags
}

// SetFlags sets the flags of the edge. One flag is currently defined, EdgeCut.
func (e *edge) SetFlags(f EdgeFlags) {
	e.flags = f
}

func (e *edge) reconnect(u, v Node) {
	switch u {
	case e.u:
		e.u = v
	case e.v:
		e.v = v
	}
}

func (e *edge) disconnect(n Node) {
	switch n {
	case e.u:
		e.u.drop(e)
		e.u = nil
	case e.v:
		e.v.drop(e)
		e.v = nil
	}
}

func (e *edge) connect(n Node) (err error) {
	switch Node(nil) {
	case e.u:
		e.u = n
		e.u.add(e)
	case e.v:
		e.v = n
		e.v.add(e)
	default:
		err = alreadyConnected
	}

	return

}
func (e *edge) join(u, v Node) {
	e.u, e.v = u, v
}

func (e *edge) String() string {
	return fmt.Sprintf("%d--%d", e.u.ID(), e.v.ID())
}

// Edges is a collection of edges used for internal representation of edge lists in a graph. 
type Edges []Edge

func (e Edges) delFromGraph(i int) Edges {
	e[i], e[len(e)-1] = e[len(e)-1], e[i]
	e[i].setIndex(i)
	e[len(e)-1].setIndex(-1)
	return e[:len(e)-1]
}

func (e Edges) delFromNode(i int) Edges {
	e[i], e[len(e)-1] = e[len(e)-1], e[i]
	return e[:len(e)-1]
}
