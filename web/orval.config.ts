import { defineConfig } from 'orval'

// orval reads the OpenAPI document the server emitted via
// `./task gen-spec` (which lands at web/openapi.json) and generates a
// typed TypeScript client into src/generated/. Every operation declared
// in server/internal/api/handlers/ shows up as an exported function +
// type definition on the frontend.
//
// Regenerate with `./task gen` after changing any server-side handler.
export default defineConfig({
  app: {
    input: {
      target: './openapi.json',
    },
    output: {
      mode: 'single',
      target: './src/generated/api.ts',
      client: 'fetch',
      override: {
        // Fetch with credentials so the yauth_session cookie travels
        // along with every request from the browser.
        operations: {},
        mutator: {
          path: './src/api/fetcher.ts',
          name: 'apiFetch',
        },
      },
      mock: false,
    },
  },
})
