package main

import (
	"context"
	"flag"
	"log"
	"os"
	"sync"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/envoyproxy/go-control-plane/pkg/test"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

var (
	debug  bool
	port   uint
	nodeID string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Use debug logging")
	flag.UintVar(&port, "port", 18000, "Management server port")
	flag.StringVar(&nodeID, "nodeID", "cluster2", "Node z")
}

// ClusterHash uses Cluster field as the node hash.
type ClusterHash struct{}

// This is an override method to use Cluster instead of a Node
func (ClusterHash) ID(node *core.Node) string {
	if node == nil {
		return ""
	}
	return node.Cluster
}

func main() {
	flag.Parse()
	ctx := context.Background()

	signal := make(chan struct{})
	cb := &callbacks{signal: signal}

	config := cache.NewSnapshotCache(false, ClusterHash{}, logger{})
	srv := server.NewServer(config, cb)

	go test.RunManagementServer(ctx, srv, port)

	log.Println("waiting for the first request...")

	select {
	case <-signal:
	case <-time.After(1 * time.Minute):
		log.Println("timeout waiting for the first request")
		os.Exit(1)
	}

	log.Println("received first request, sending abort fault setting...")

	var runtimes = []cache.Resource{
		&discovery.Runtime{
			Name: "rtds",
			Layer: &pstruct.Struct{
				Fields: map[string]*pstruct.Value{
					"fault.http.abort.abort_percent": &pstruct.Value{
						Kind: &pstruct.Value_NumberValue{
							NumberValue: 100,
						},
					},
					"fault.http.abort.http_status": &pstruct.Value{
						Kind: &pstruct.Value_NumberValue{
							NumberValue: 404,
						},
					},
				},
			},
		},
	}
	snapshot := cache.NewSnapshot("1", nil, nil, nil, nil, runtimes)
	config.SetSnapshot(nodeID, snapshot)

	log.Println("Abort fault setting sent")
	time.Sleep(10 * time.Second)

	var runtimes1 = []cache.Resource{
		&discovery.Runtime{
			Name: "rtds",
			Layer: &pstruct.Struct{
				Fields: map[string]*pstruct.Value{
				},
			},
		},
	}

	snapshot1 := cache.NewSnapshot("2", nil, nil, nil, nil, runtimes1)
	config.SetSnapshot(nodeID, snapshot1)
	log.Println("Clearing the fault")
}

type logger struct{}

func (logger logger) Infof(format string, args ...interface{}) {
	if debug {
		log.Printf(format+"\n", args...)
	}
}
func (logger logger) Errorf(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}

type callbacks struct {
	signal   chan struct{}
	fetches  int
	requests int
	mu       sync.Mutex
}

func (cb *callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	log.Printf("server callbacks fetches=%d requests=%d\n", cb.fetches, cb.requests)
}
func (cb *callbacks) OnStreamOpen(_ context.Context, id int64, typ string) error {
	if debug {
		log.Printf("stream %d open for %s\n", id, typ)
	}
	return nil
}
func (cb *callbacks) OnStreamClosed(id int64) {
	if debug {
		log.Printf("stream %d closed\n", id)
	}
}
func (cb *callbacks) OnStreamRequest(int64, *v2.DiscoveryRequest) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.requests++
	if cb.signal != nil {
		close(cb.signal)
		cb.signal = nil
	}
	return nil
}
func (cb *callbacks) OnStreamResponse(int64, *v2.DiscoveryRequest, *v2.DiscoveryResponse) {}
func (cb *callbacks) OnFetchRequest(_ context.Context, req *v2.DiscoveryRequest) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.fetches++
	if cb.signal != nil {
		close(cb.signal)
		cb.signal = nil
	}
	return nil
}
func (cb *callbacks) OnFetchResponse(*v2.DiscoveryRequest, *v2.DiscoveryResponse) {}
