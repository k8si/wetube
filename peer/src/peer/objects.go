package main

import (
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	// "encoding/hex"
	"fmt"
	// "io"
	"log"
	"sync"
)

var keystore = struct {
	m map[string]*rsa.PublicKey
	sync.Mutex
}{m: make(map[string]*rsa.PublicKey)}

func addkey(addr string, k *rsa.PublicKey) {
	keystore.Lock()
	defer keystore.Unlock()
	if _, ok := keystore.m[addr]; !ok {
		keystore.m[addr] = k
	}
}

func getkey(addr string) *rsa.PublicKey {
	if _, ok := keystore.m[addr]; ok {
		return keystore.m[addr]
	}
	return nil
}

type Message struct {
	ID        string
	Sender    string
	Subject   string
	Body      string
	Signature []byte
}

func (m *Message) String() string {
	s := m.ID + m.Sender + m.Subject + m.Body
	return s
}

func (m *Message) Hash() []byte {
	hasher := md5.New()
	hasher.Write([]byte(m.String()))
	return hasher.Sum(nil)
}

func (m *Message) sign() {
	if privkey != nil {
		hashed := m.Hash()
		s, err := rsa.SignPKCS1v15(rand.Reader, privkey, crypto.MD5, hashed)
		if err != nil {
			log.Panic(err)
		}
		m.Signature = s
	} else {
		log.Fatal("no private key found")
	}
}

func (m *Message) verify() bool {
	pubkey := getkey(m.Sender)
	if pubkey == nil {
		log.Println("no public key found for addr ", m.Sender)
		return false
	}
	hashed := m.Hash()
	err := rsa.VerifyPKCS1v15(pubkey, crypto.MD5, hashed, m.Signature)
	if err != nil {
		log.Println("message verification failure: ", err)
		return false
	}
	return true
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

func (h *Hub) Size() int {
	return len(h.peers)
}
