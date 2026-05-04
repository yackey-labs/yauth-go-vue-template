// Package handlers groups the application's HTTP handlers. Each route
// has a `register*` function that takes a huma.API and declares its
// operation; Register() composes them all.
//
// Adding a new route:
//  1. Create a new file (e.g. posts.go) defining input/output structs
//     and a registerPosts(api huma.API) function.
//  2. Call registerPosts(api) from Register below.
//  3. If the route is protected, wire RequireAuth in api.NewRouter for
//     the new path.
package handlers

import "github.com/danielgtaylor/huma/v2"

// Register wires every app-defined operation into the Huma API.
func Register(api huma.API) {
	registerMe(api)
}
