package main

import (
	"regexp"
)

type Route struct {
	Name       string
	Verb       string
	Expression *regexp.Regexp
}

type RouteMatch struct {
	Name       string
	Verb       string
	Path       string
	Variables  map[string]string
	Expression *regexp.Regexp
}

type Router struct {
	Routes []*Route
}

func (r *Router) AddRoute(name string, verb string, expression string) {
	route := &Route{}
	route.Name = name
	route.Verb = verb
	route.Expression = regexp.MustCompile(expression)

	r.Routes = append(r.Routes, route)
}

func (r *Router) Match(verb string, path string) *RouteMatch {
	for _, route := range r.Routes {
		if route.Verb == verb && route.Expression.MatchString(path) {
			match := &RouteMatch{}
			match.Variables = make(map[string]string)

			// The first element of names and values will be either the entire
			// matched string or an empty name.
			names := route.Expression.SubexpNames()[1:]
			values := route.Expression.FindStringSubmatch(path)[1:]

			for i := 0; i < len(names); i++ {
				match.Variables[names[i]] = values[i]
			}

			match.Path = path
			match.Name = route.Name
			match.Verb = route.Verb
			match.Expression = route.Expression

			return match
		}
	}

	return nil
}
