package main

import (
	"testing"
)

var (
	pattern string = "/(?P<queue>[a-z]+)"
)

func TestRouterCreation(t *testing.T) {
	r := &Router{}
	r.AddRoute("routeName1", "GET", "/route1")
	r.AddRoute("routeName2", "GET", "/route2")
	r.AddRoute("routeName3", "GET", "/route3")

	if len(r.Routes) != 3 {
		t.Error("Routes were not attached correctly.")
	}
}

func TestRouterMatching(t *testing.T) {
	r := &Router{}
	r.AddRoute("routeName1", "GET", pattern)
	r.AddRoute("routeName2", "PUT", pattern)

	match := r.Match("PUT", "/myqueue")

	if match.Name != "routeName2" {
		t.Error("Did not match name.")
	}

	if match.Path != "/myqueue" {
		t.Error("Did not match path.")
	}

	if match.Verb != "PUT" {
		t.Error("Did not match verb.")
	}

	if match.Expression.String() != pattern {
		t.Error("Did not match expression.")
	}
}

func BenchmarkPositiveMatch(b *testing.B) {
	r := &Router{}
	r.AddRoute("routeName1", "GET", pattern)

	for i := 0; i < b.N; i++ {
		r.Match("GET", "/myqueue")
	}
}

func BenchmarkNegativeMatch(b *testing.B) {
	r := &Router{}
	r.AddRoute("routeName1", "GET", pattern)

	for i := 0; i < b.N; i++ {
		r.Match("GET", "12345")
	}
}
