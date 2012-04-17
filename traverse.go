package graph

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

import (
	"errors"
)

var notFound = errors.New("graph: target not found")

// BreadthFirst is a type that can perform a breadth-first search on a graph.
type BreadthFirst struct {
	q      *queue
	visits []bool
}

// NewBreadthFirst creates a new BreadthFirst searcher.
func NewBreadthFirst() *BreadthFirst {
	return &BreadthFirst{q: &queue{}}
}

// Search searches a graph starting from node s until the NodeFilter function nf returns a value of
// true, traversing edges in the graph that allow the Edgefilter function ef to return true. On success
// the terminating node, t is returned. If no node is found an error is returned.
func (self *BreadthFirst) Search(s *Node, ef EdgeFilter, nf NodeFilter) (t *Node, err error) {
	self.q.Enqueue(s)
	self.visits = mark(s, self.visits)
	for self.q.Len() > 0 {
		t, err = self.q.Dequeue()
		if err != nil {
			return nil, err
		}
		if nf(t) {
			return
		}
		for _, n := range t.Neighbors(ef) {
			if !self.Visited(n) {
				self.visits = mark(n, self.visits)
				self.q.Enqueue(n)
			}
		}
	}

	return nil, notFound
}

// Visited marks the node n as having been visited by the sercher.
func (self *BreadthFirst) Visited(n *Node) bool {
	id := n.id
	if id < 0 || id >= len(self.visits) {
		return false
	}
	return self.visits[n.id]
}

// Reset clears the search queue and visited list.
func (self *BreadthFirst) Reset() {
	self.q.Clear()
	self.visits = self.visits[:0]
}

// DepthFirst is a type that can perform a depth-first search on a graph.
type DepthFirst struct {
	s      *stack
	visits []bool
}

// NewDepthFirst creates a new DepthFirst searcher.
func NewDepthFirst() *DepthFirst {
	return &DepthFirst{s: &stack{}}
}

// Search searches a graph starting from node s until the NodeFilter function nf returns a value of
// true, traversing edges in the graph that allow the Edgefilter function ef to return true. On success
// the terminating node, t is returned. If no node is found an error is returned.
func (self *DepthFirst) Search(s *Node, ef EdgeFilter, nf NodeFilter) (t *Node, err error) {
	self.s.Push(s)
	self.visits = mark(s, self.visits)
	for self.s.Len() > 0 {
		t, err = self.s.Pop()
		if err != nil {
			return nil, err
		}
		if nf(t) {
			return
		}
		for _, n := range t.Neighbors(ef) {
			if !self.Visited(n) {
				self.visits = mark(n, self.visits)
				self.s.Push(n)
			}
		}
	}

	return nil, notFound
}

// Visited marks the node n as having been visited by the sercher.
func (self *DepthFirst) Visited(n *Node) bool {
	id := n.id
	if id < 0 || id >= len(self.visits) {
		return false
	}
	return self.visits[n.id]
}

// Reset clears the search stack and visited list.
func (self *DepthFirst) Reset() {
	self.s.Clear()
	self.visits = self.visits[:0]
}

func mark(n *Node, v []bool) []bool {
	id := n.id
	if id == len(v) {
		v = append(v, true)
	} else if id > len(v) {
		t := make([]bool, id+1)
		copy(t, v)
		v = t
		v[id] = true
	} else {
		v[id] = true
	}

	return v
}