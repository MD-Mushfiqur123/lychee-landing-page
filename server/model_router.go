package server

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/lychee/lychee/api"
)

type endpointState struct {
	endpoint api.ModelEndpoint
	active   int32
}

type routeState struct {
	route     api.ModelRoute
	endpoints []*endpointState
	counter   uint64
}

// ModelRouter load-balances requests to virtual models across defined backend endpoints.
type ModelRouter struct {
	mu        sync.RWMutex
	routes    map[string]*routeState
	storePath string
}

// NewModelRouter creates and initializes a ModelRouter.
func NewModelRouter(storePath string) *ModelRouter {
	mr := &ModelRouter{
		routes:    make(map[string]*routeState),
		storePath: storePath,
	}
	_ = mr.load()
	return mr
}

func (mr *ModelRouter) load() error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	data, err := os.ReadFile(mr.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var routesList []api.ModelRoute
	if err := json.Unmarshal(data, &routesList); err != nil {
		return err
	}

	for _, r := range routesList {
		endpoints := make([]*endpointState, len(r.Endpoints))
		for i, ep := range r.Endpoints {
			endpoints[i] = &endpointState{endpoint: ep}
		}
		mr.routes[r.Name] = &routeState{
			route:     r,
			endpoints: endpoints,
		}
	}
	return nil
}

func (mr *ModelRouter) persist() error {
	var routesList []api.ModelRoute
	for _, state := range mr.routes {
		routesList = append(routesList, state.route)
	}

	data, err := json.MarshalIndent(routesList, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(mr.storePath), 0755); err != nil {
		return err
	}

	return os.WriteFile(mr.storePath, data, 0644)
}

// AddRoute registers a new virtual model route.
func (mr *ModelRouter) AddRoute(route api.ModelRoute) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if route.Name == "" {
		return errors.New("route name is required")
	}
	if len(route.Endpoints) == 0 {
		return errors.New("at least one endpoint is required")
	}
	if route.Strategy == "" {
		route.Strategy = "round_robin"
	}

	endpoints := make([]*endpointState, len(route.Endpoints))
	for i, ep := range route.Endpoints {
		endpoints[i] = &endpointState{endpoint: ep}
	}

	mr.routes[route.Name] = &routeState{
		route:     route,
		endpoints: endpoints,
	}

	return mr.persist()
}

// RemoveRoute deletes a virtual model route.
func (mr *ModelRouter) RemoveRoute(name string) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if _, ok := mr.routes[name]; !ok {
		return errors.New("route not found")
	}

	delete(mr.routes, name)
	return mr.persist()
}

// ListRoutes lists all registered virtual model routes.
func (mr *ModelRouter) ListRoutes() []api.ModelRoute {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	routesList := make([]api.ModelRoute, 0, len(mr.routes))
	for _, state := range mr.routes {
		routesList = append(routesList, state.route)
	}
	return routesList
}

// Resolve matches a virtual name to a target endpoint and returns a callback to release the load tracking lock.
func (mr *ModelRouter) Resolve(name string) (*api.ModelEndpoint, func(), error) {
	mr.mu.RLock()
	state, ok := mr.routes[name]
	mr.mu.RUnlock()

	if !ok {
		return nil, func() {}, nil
	}

	if len(state.endpoints) == 0 {
		return nil, func() {}, errors.New("no endpoints for route")
	}

	var chosen *endpointState

	switch state.route.Strategy {
	case "random":
		idx := rand.Intn(len(state.endpoints))
		chosen = state.endpoints[idx]
	case "least_loaded":
		chosen = state.endpoints[0]
		minActive := atomic.LoadInt32(&chosen.active)
		for _, ep := range state.endpoints[1:] {
			act := atomic.LoadInt32(&ep.active)
			if act < minActive {
				minActive = act
				chosen = ep
			}
		}
	case "round_robin":
		fallthrough
	default:
		val := atomic.AddUint64(&state.counter, 1)
		idx := (val - 1) % uint64(len(state.endpoints))
		chosen = state.endpoints[idx]
	}

	atomic.AddInt32(&chosen.active, 1)

	release := func() {
		atomic.AddInt32(&chosen.active, -1)
	}

	return &chosen.endpoint, release, nil
}
