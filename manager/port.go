package manager

import (
    "sync"
    "sort"
)

type Ports struct {
    sync.RWMutex
    m map[int32]bool
}

func New(items... int32) *Ports {
    p := &Ports{
        m: make(map[int32]bool, len(items)),
    }
    p.Add(items...)
    return p
}

func (p *Ports) Duplicate() *Ports {
    p.Lock()
    defer p.Unlock()
    r := &Ports{
        m: make(map[int32]bool, len(p.m)),
    }
    for e := range p.m {
        r.m[e] = true
    }
    return r
}

func (p *Ports) Add(items... int32) {
    p.Lock()
    defer p.Unlock()
    for _, v := range items {
        p.m[v] = true
    }
}

func (p *Ports) Remove(items... int32) {
    p.Lock()
    defer p.Unlock()
    for _, v := range items {
        delete(p.m, v)
    }
}

func (p *Ports) Has(items... int32) {
    p.RLock()
    defer p.RUnlock()
    for _, v := range items {
        if _, ok := p.m[v]; !ok {
            return false
        }
    }
    return true
}

func (p *Ports) Empty() bool {
    p.Lock()
    defer p.Unlock()
    return len(p.m) == 0
}

func (p *Ports) Clear() {
    p.Lock()
    defer p.Unlock()
    p.m = map[int32]bool{}
}

func (p *Ports) List() []int32 {
    p.RLock()
    defer p.RUnlock()
    list := make([]int32, 0, len(p.m))
    for item := range p.m {
        list = append(list, item)
    }
    sort.Ints(list)
    return list
}

func Minus(sets ...*Ports) *Ports {
    if len(sets) == 0 {
        return New()
    } else if len(sets) == 1 {
        return sets[0]
    }
    r := sets[0].Duplicate()
    for _, set := range sets[1:] {
        for e := range set.m {
            delete(r.m, e)
        }
    }
    return r
}