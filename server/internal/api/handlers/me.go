package handlers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	yauthmw "github.com/yackey-labs/yauth-go/middleware"
)

// MeResponse is the JSON shape returned by GET /api/me.
//
// Tags here drive both the runtime serialization AND the generated
// OpenAPI schema, which orval consumes to type the frontend client.
type MeResponse struct {
	ID            string `json:"id" doc:"Stable user identifier (UUID)."`
	Email         string `json:"email" format:"email" doc:"Primary email address."`
	Role          string `json:"role" enum:"user,admin" doc:"Role assigned to the user."`
	EmailVerified bool   `json:"email_verified" doc:"True after the user clicks the verification link."`
	AuthMethod    string `json:"auth_method" enum:"cookie,bearer,apikey" doc:"How this request was authenticated."`
}

// meInput is intentionally empty — GET /api/me has no params, body, or
// headers we care about typing. The cookie is read via context, set by
// yauth's session middleware upstream.
type meInput struct{}

type meOutput struct {
	Body MeResponse
}

// registerMe wires GET /api/me into the Huma API. Both the runtime
// behaviour and the OpenAPI operation flow from this single declaration.
func registerMe(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getMe",
		Method:      "GET",
		Path:        "/api/me",
		Summary:     "Return the authenticated user.",
		Description: "Companion to GET /api/auth/session, but emitted by " +
			"the application's own OpenAPI document. Demonstrates how to " +
			"protect a custom route with ya.Middleware().RequireAuth.",
		Tags: []string{"app"},
		Security: []map[string][]string{
			{"sessionCookie": {}},
		},
	}, func(ctx context.Context, _ *meInput) (*meOutput, error) {
		user, ok := yauthmw.AuthUserFromContext(ctx)
		if !ok {
			return nil, huma.Error401Unauthorized("unauthenticated")
		}
		return &meOutput{
			Body: MeResponse{
				ID:            user.User.ID,
				Email:         user.User.Email,
				Role:          user.User.Role,
				EmailVerified: user.User.EmailVerified,
				AuthMethod:    user.Method,
			},
		}, nil
	})
}
