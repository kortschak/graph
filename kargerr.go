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
	"math"
	"runtime"
	"sync"
)

// FIXME Use Index() instead of ID() on edges and nodes - this requires a change to node.go

const sqrt2 = 1.4142135623730950488016887242096980785696718753769480

var MaxProcs = runtime.GOMAXPROCS(0)

func FastRandMinCut(g *Undirected, iter int) (c []Edge, w float64) {
	ka := newKargerR(g)
	ka.init()
	w = math.Inf(1)
	for i := 0; i < iter; i++ {
		ka.fastRandMinCut()
		if ka.w < w {
			w = ka.w
			c = ka.c
		}
	}

	return
}

// parallelised outside the recursion tree

func FastRandMinCutPar(g *Undirected, iter, thread int) (c []Edge, w float64) {
	if thread > MaxProcs {
		thread = MaxProcs
	}
	if thread > iter {
		thread = iter
	}
	iter, rem := iter/thread+1, iter%thread

	type r struct {
		c []Edge
		w float64
	}
	rs := make([]*r, thread)

	wg := &sync.WaitGroup{}
	for j := 0; j < thread; j++ {
		if rem == 0 {
			iter--
		}
		if rem >= 0 {
			rem--
		}
		wg.Add(1)
		go func(j, iter int) {
			defer wg.Done()
			ka := newKargerR(g)
			ka.init()
			var (
				w = math.Inf(1)
				c []Edge
			)
			for i := 0; i < iter; i++ {
				ka.fastRandMinCut()
				if ka.w < w {
					w = ka.w
					c = ka.c
				}
			}

			rs[j] = &r{c, w}
		}(j, iter)
	}

	w = math.Inf(1)
	wg.Wait()
	for _, subr := range rs {
		if subr.w < w {
			w = subr.w
			c = subr.c
		}
	}

	return
}

type kargerR struct {
	g     *Undirected
	order int
	ind   []super
	sel   Selector
	c     []Edge
	w     float64
}

func newKargerR(g *Undirected) *kargerR {
	return &kargerR{
		g:   g,
		ind: make([]super, g.NextNodeID()),
		sel: make(Selector, g.Size()),
	}
}

func (ka *kargerR) init() {
	ka.order = ka.g.Order()
	for i := range ka.ind {
		ka.ind[i].label = -1
		ka.ind[i].nodes = nil
	}
	for _, n := range ka.g.Nodes() {
		id := n.ID()
		ka.ind[id].label = id
	}
	for i, e := range ka.g.Edges() {
		ka.sel[i] = WeightedItem{Index: e.ID(), Weight: e.Weight()}
	}
	ka.sel.Init()
}

func (ka *kargerR) clone() (c *kargerR) {
	c = &kargerR{
		g:     ka.g,
		ind:   make([]super, ka.g.NextNodeID()),
		sel:   make(Selector, ka.g.Size()),
		order: ka.order,
	}

	copy(c.sel, ka.sel)
	for i, n := range ka.ind {
		s := &c.ind[i]
		s.label = n.label
		if n.nodes != nil {
			s.nodes = make([]int, len(n.nodes))
			copy(s.nodes, n.nodes)
		}
	}

	return
}

func (ka *kargerR) fastRandMinCut() {
	if ka.order <= 6 {
		ka.randCompact(2)
		return
	}

	t := int(math.Ceil(float64(ka.order)/sqrt2 + 1))

	sub := []*kargerR{ka, ka.clone()}
	for i := range sub {
		sub[i].randContract(t)
		sub[i].fastRandMinCut()
	}

	if sub[0].w < sub[1].w {
		*ka = *sub[0]
		return
	}
	*ka = *sub[1]
}

func (ka *kargerR) randContract(k int) {
	for ka.order > k {
		id, err := ka.sel.Select()
		if err != nil {
			break
		}

		e := ka.g.Edge(id)
		if ka.loop(e) {
			continue
		}

		hid, tid := e.Head().ID(), e.Tail().ID()
		hl, tl := ka.ind[hid].label, ka.ind[tid].label
		if len(ka.ind[hl].nodes) < len(ka.ind[tl].nodes) {
			hid, tid = tid, hid
			hl, tl = tl, hl
		}

		if ka.ind[hl].nodes == nil {
			ka.ind[hl].nodes = []int{hid}
		}
		if ka.ind[tl].nodes == nil {
			ka.ind[hl].nodes = append(ka.ind[hl].nodes, tid)
		} else {
			ka.ind[hl].nodes = append(ka.ind[hl].nodes, ka.ind[tl].nodes...)
			ka.ind[tl].nodes = nil
		}
		for _, i := range ka.ind[hl].nodes {
			ka.ind[i].label = ka.ind[hid].label
		}

		ka.order--
	}
}

func (ka *kargerR) randCompact(k int) {
	for ka.order > k {
		id, err := ka.sel.Select()
		if err != nil {
			break
		}

		e := ka.g.Edge(id)
		if ka.loop(e) {
			continue
		}

		hid, tid := e.Head().ID(), e.Tail().ID()
		hl, tl := ka.ind[hid].label, ka.ind[tid].label
		if len(ka.ind[hl].nodes) < len(ka.ind[tl].nodes) {
			hid, tid = tid, hid
			hl, tl = tl, hl
		}

		if ka.ind[hl].nodes == nil {
			ka.ind[hl].nodes = []int{hid}
		}
		if ka.ind[tl].nodes == nil {
			ka.ind[hl].nodes = append(ka.ind[hl].nodes, tid)
		} else {
			ka.ind[hl].nodes = append(ka.ind[hl].nodes, ka.ind[tl].nodes...)
			ka.ind[tl].nodes = nil
		}
		for _, i := range ka.ind[hl].nodes {
			ka.ind[i].label = ka.ind[hid].label
		}

		ka.order--
	}

	ka.c, ka.w = []Edge{}, 0
	for _, e := range ka.g.Edges() {
		if ka.loop(e) {
			continue
		}
		ka.c = append(ka.c, e)
		ka.w += e.Weight()
	}
}

func (ka *kargerR) loop(e Edge) bool {
	return ka.ind[e.Head().ID()].label == ka.ind[e.Tail().ID()].label
}

// parallelised within the recursion tree

func ParFastRandMinCut(g *Undirected, iter, threads int) (c []Edge, w float64) {
	k := newKargerRP(g)
	k.split = threads
	if k.split == 0 {
		k.split = -1
	}
	k.init()
	w = math.Inf(1)
	for i := 0; i < iter; i++ {
		k.fastRandMinCut()
		if k.w < w {
			w = k.w
			c = k.c
		}
	}

	return
}

type kargerRP struct {
	g     *Undirected
	order int
	ind   []super
	sel   Selector
	c     []Edge
	w     float64
	count int
	split int
}

func newKargerRP(g *Undirected) *kargerRP {
	return &kargerRP{
		g:   g,
		ind: make([]super, g.NextNodeID()),
		sel: make(Selector, g.Size()),
	}
}

func (ka *kargerRP) init() {
	ka.order = ka.g.Order()
	for i := range ka.ind {
		ka.ind[i].label = -1
		ka.ind[i].nodes = nil
	}
	for _, n := range ka.g.Nodes() {
		id := n.ID()
		ka.ind[id].label = id
	}
	for i, e := range ka.g.Edges() {
		ka.sel[i] = WeightedItem{Index: e.ID(), Weight: e.Weight()}
	}
	ka.sel.Init()
}

func (ka *kargerRP) clone() (c *kargerRP) {
	c = &kargerRP{
		g:     ka.g,
		ind:   make([]super, ka.g.NextNodeID()),
		sel:   make(Selector, ka.g.Size()),
		order: ka.order,
		count: ka.count,
	}

	copy(c.sel, ka.sel)
	for i, n := range ka.ind {
		s := &c.ind[i]
		s.label = n.label
		if n.nodes != nil {
			s.nodes = make([]int, len(n.nodes))
			copy(s.nodes, n.nodes)
		}
	}

	return
}

func (ka *kargerRP) fastRandMinCut() {
	if ka.order <= 6 {
		ka.randCompact(2)
		return
	}

	t := int(math.Ceil(float64(ka.order)/sqrt2 + 1))

	var wg *sync.WaitGroup
	if ka.count < ka.split {
		wg = &sync.WaitGroup{}
	}
	ka.count++

	sub := []*kargerRP{ka, ka.clone()}
	for i := range sub {
		if wg != nil {
			wg.Add(1)
			go func(i int) {
				runtime.LockOSThread()
				defer wg.Done()
				sub[i].randContract(t)
				sub[i].fastRandMinCut()
			}(i)
		} else {
			sub[i].randContract(t)
			sub[i].fastRandMinCut()
		}
	}

	if wg != nil {
		wg.Wait()
	}

	if sub[0].w < sub[1].w {
		*ka = *sub[0]
		return
	}
	*ka = *sub[1]
}

func (ka *kargerRP) randContract(k int) {
	for ka.order > k {
		id, err := ka.sel.Select()
		if err != nil {
			break
		}

		e := ka.g.Edge(id)
		if ka.loop(e) {
			continue
		}

		hid, tid := e.Head().ID(), e.Tail().ID()
		hl, tl := ka.ind[hid].label, ka.ind[tid].label
		if len(ka.ind[hl].nodes) < len(ka.ind[tl].nodes) {
			hid, tid = tid, hid
			hl, tl = tl, hl
		}

		if ka.ind[hl].nodes == nil {
			ka.ind[hl].nodes = []int{hid}
		}
		if ka.ind[tl].nodes == nil {
			ka.ind[hl].nodes = append(ka.ind[hl].nodes, tid)
		} else {
			ka.ind[hl].nodes = append(ka.ind[hl].nodes, ka.ind[tl].nodes...)
			ka.ind[tl].nodes = nil
		}
		for _, i := range ka.ind[hl].nodes {
			ka.ind[i].label = ka.ind[hid].label
		}

		ka.order--
	}
}

func (ka *kargerRP) randCompact(k int) {
	for ka.order > k {
		id, err := ka.sel.Select()
		if err != nil {
			break
		}

		e := ka.g.Edge(id)
		if ka.loop(e) {
			continue
		}

		hid, tid := e.Head().ID(), e.Tail().ID()
		hl, tl := ka.ind[hid].label, ka.ind[tid].label
		if len(ka.ind[hl].nodes) < len(ka.ind[tl].nodes) {
			hid, tid = tid, hid
			hl, tl = tl, hl
		}

		if ka.ind[hl].nodes == nil {
			ka.ind[hl].nodes = []int{hid}
		}
		if ka.ind[tl].nodes == nil {
			ka.ind[hl].nodes = append(ka.ind[hl].nodes, tid)
		} else {
			ka.ind[hl].nodes = append(ka.ind[hl].nodes, ka.ind[tl].nodes...)
			ka.ind[tl].nodes = nil
		}
		for _, i := range ka.ind[hl].nodes {
			ka.ind[i].label = ka.ind[hid].label
		}

		ka.order--
	}

	ka.c, ka.w = []Edge{}, 0
	for _, e := range ka.g.Edges() {
		if ka.loop(e) {
			continue
		}
		ka.c = append(ka.c, e)
		ka.w += e.Weight()
	}
}

func (ka *kargerRP) loop(e Edge) bool {
	return ka.ind[e.Head().ID()].label == ka.ind[e.Tail().ID()].label
}
