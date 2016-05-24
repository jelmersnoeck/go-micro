package blacklist

import (
	"math/rand"
	"sync"
	"time"

	"github.com/micro/go-micro/registry"
)

type blackListNode struct {
	age     time.Time
	id      string
	service string
	count   int
}

type BlackList struct {
	ttl  int
	exit chan bool

	sync.RWMutex
	bl map[string]blackListNode
}

var (
	// number of times we see an error before blacklisting
	count = 3

	// the ttl to blacklist for
	ttl = 30
)

func init() {
	rand.Seed(time.Now().Unix())
}

func (r *BlackList) purge() {
}

func (r *BlackList) run() {
}

func (r *BlackList) Filter(services []*registry.Service) ([]*registry.Service, error) {
	return services, nil
}

func (r *BlackList) Mark(service string, node *registry.Node, err error) {
	return
}

func (r *BlackList) Reset(service string) {
	return
}

func (r *BlackList) Close() error {
	return nil
}

func New() *BlackList {
	return &BlackList{}
}
