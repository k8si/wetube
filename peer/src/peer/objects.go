package main

import (
	"fmt"
	"sync"
)

type Message struct {
	ID      string
	Sender  string
	Subject string
	Body    string
}

//used for voting
type NodeIDRegistry struct {
	id int
	mu sync.RWMutex
}

func (n *NodeIDRegistry) getNewID() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.id += 1
	v := n.id
	return v
}

type Hub struct {
	peers map[string]chan<- Message
	mu    sync.RWMutex
}

func (h *Hub) Add(addr string) <-chan Message {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.peers[addr]; ok {
		return nil
	}
	ch := make(chan Message)
	h.peers[addr] = ch
	// fmt.Printf("hub.Add: added addr %s\n", addr)
	return ch
}

func (h *Hub) Remove(addr string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.peers[addr] != nil {
		close(h.peers[addr])
	}
	delete(h.peers, addr)
}

func (h *Hub) List() []chan<- Message {
	h.mu.RLock()
	defer h.mu.RUnlock()
	l := make([]chan<- Message, 0, len(h.peers))
	for _, ch := range h.peers {
		l = append(l, ch)
	}
	return l
}

func (h *Hub) ListAddrs() []string {
	l := make([]string, 0)
	for a, ch := range h.peers {
		if ch != nil {
			l = append(l, a)
		}
	}
	return l
}

func (h *Hub) PrintAll() {
	// fmt.Println("hub connected to:")
	for k, _ := range h.peers {
		fmt.Printf("\t\t%s\n", k)
	}
}
