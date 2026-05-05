import { fileURLToPath } from 'node:url'

import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig, loadEnv } from 'vite'

// envDir points to the repo root so Vite picks up VITE_* values from
// the same .env that drives the Go server. Keeps endpoint
// configuration in one place.
const repoRoot = fileURLToPath(new URL('..', import.meta.url))

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, repoRoot, '')

  // Where browser-side OTel exports go. Same-origin in dev (Vite
  // proxies /v1/traces → the configured collector) so the browser
  // doesn't trip on CORS. Prod needs the same shape: a reverse proxy
  // in front of the SPA forwards /v1/traces to the collector.
  //
  // VITE_OTEL_EXPORTER_OTLP_ENDPOINT overrides OTEL_EXPORTER_OTLP_ENDPOINT
  // for the frontend specifically — handy when frontend traces should
  // go to a different collector. By default they share.
  const otelTarget =
    env.VITE_OTEL_EXPORTER_OTLP_ENDPOINT ||
    env.OTEL_EXPORTER_OTLP_ENDPOINT ||
    'https://otel-local.yackey.cloud'

  return {
    envDir: repoRoot,
    // Expose OTEL_* alongside VITE_* so the same env var can drive
    // both halves of the stack. import.meta.env reflects both.
    envPrefix: ['VITE_', 'OTEL_'],
    plugins: [vue(), tailwindcss()],
    server: {
      port: 5173,
      proxy: {
        '/api': {
          target: 'http://localhost:3000',
          changeOrigin: true,
        },
        '/v1/traces': {
          target: otelTarget,
          changeOrigin: true,
        },
      },
    },
  }
})
