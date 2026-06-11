package server

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/lychee/lychee/api"
)

func TestModelRouter(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "routes.json")
	router := NewModelRouter(storePath)

	route := api.ModelRoute{
		Name: "fast",
		Endpoints: []api.ModelEndpoint{
			{Host: "http://192.168.1.10:11434", Model: "gemma3"},
			{Host: "http://192.168.1.20:11434", Model: "gemma3"},
			{Host: "http://192.168.1.30:11434", Model: "gemma3"},
		},
		Strategy: "round_robin",
	}

	t.Run("add and list routes", func(t *testing.T) {
		err := router.AddRoute(route)
		if err != nil {
			t.Fatalf("failed to add route: %v", err)
		}

		routes := router.ListRoutes()
		if len(routes) != 1 {
			t.Fatalf("expected 1 route, got %d", len(routes))
		}
		if routes[0].Name != "fast" {
			t.Errorf("expected route name 'fast', got %q", routes[0].Name)
		}
	})

	t.Run("persist and load routes", func(t *testing.T) {
		router2 := NewModelRouter(storePath)
		routes := router2.ListRoutes()
		if len(routes) != 1 {
			t.Fatalf("expected 1 route on load, got %d", len(routes))
		}
		if routes[0].Name != "fast" {
			t.Errorf("expected loaded route name 'fast', got %q", routes[0].Name)
		}
	})

	t.Run("round robin resolution", func(t *testing.T) {
		ep1, rel1, _ := router.Resolve("fast")
		ep2, rel2, _ := router.Resolve("fast")
		ep3, rel3, _ := router.Resolve("fast")
		ep4, rel4, _ := router.Resolve("fast")

		defer rel1()
		defer rel2()
		defer rel3()
		defer rel4()

		if ep1.Host != "http://192.168.1.10:11434" {
			t.Errorf("expected first call to resolve host 10, got %q", ep1.Host)
		}
		if ep2.Host != "http://192.168.1.20:11434" {
			t.Errorf("expected second call to resolve host 20, got %q", ep2.Host)
		}
		if ep3.Host != "http://192.168.1.30:11434" {
			t.Errorf("expected third call to resolve host 30, got %q", ep3.Host)
		}
		if ep4.Host != "http://192.168.1.10:11434" {
			t.Errorf("expected fourth call to wrap around to host 10, got %q", ep4.Host)
		}
	})

	t.Run("random strategy resolution", func(t *testing.T) {
		randRoute := api.ModelRoute{
			Name: "random-route",
			Endpoints: []api.ModelEndpoint{
				{Host: "host1", Model: "m"},
				{Host: "host2", Model: "m"},
			},
			Strategy: "random",
		}
		_ = router.AddRoute(randRoute)

		for i := 0; i < 20; i++ {
			ep, rel, err := router.Resolve("random-route")
			if err != nil {
				t.Fatalf("resolve failed: %v", err)
			}
			rel()
			if ep.Host != "host1" && ep.Host != "host2" {
				t.Errorf("unexpected resolved host: %q", ep.Host)
			}
		}
	})

	t.Run("least loaded strategy resolution", func(t *testing.T) {
		llRoute := api.ModelRoute{
			Name: "least-loaded-route",
			Endpoints: []api.ModelEndpoint{
				{Host: "host1", Model: "m"},
				{Host: "host2", Model: "m"},
			},
			Strategy: "least_loaded",
		}
		_ = router.AddRoute(llRoute)

		// Both are active 0, host1 (first) is returned.
		epA, relA, _ := router.Resolve("least-loaded-route")
		
		// host1 is active = 1, host2 is active = 0. Resolves to host2.
		epB, relB, _ := router.Resolve("least-loaded-route")
		
		if epA.Host != "host1" {
			t.Errorf("expected epA to be host1, got %q", epA.Host)
		}
		if epB.Host != "host2" {
			t.Errorf("expected epB to be host2, got %q", epB.Host)
		}

		relA()
		relB()
	})

	t.Run("resolve unknown name", func(t *testing.T) {
		ep, _, err := router.Resolve("unknown-model")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ep != nil {
			t.Errorf("expected nil endpoint for unknown name, got %v", ep)
		}
	})

	t.Run("concurrent resolve safety", func(t *testing.T) {
		var wg sync.WaitGroup
		concurrency := 50

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ep, rel, err := router.Resolve("fast")
				if err == nil && ep != nil {
					rel()
				}
			}()
		}
		wg.Wait()
	})

	t.Run("remove route", func(t *testing.T) {
		err := router.RemoveRoute("fast")
		if err != nil {
			t.Fatalf("failed to remove route: %v", err)
		}

		routes := router.ListRoutes()
		for _, r := range routes {
			if r.Name == "fast" {
				t.Error("expected route 'fast' to be removed")
			}
		}

		ep, _, _ := router.Resolve("fast")
		if ep != nil {
			t.Error("expected resolved endpoint to be nil after deletion")
		}
	})
}
