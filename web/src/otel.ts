// Browser-side OpenTelemetry. Mirrors the server's setup
// (internal/telemetry) but for the SPA: a WebTracerProvider with the
// OTLP/HTTP exporter, plus FetchInstrumentation so every typed-client
// call automatically becomes a client span and carries `traceparent`
// to the backend. The backend's otelhttp wrapper then continues the
// trace under that parent — full browser → server visibility.
//
// Spans are POSTed to a same-origin `/v1/traces`. In dev, the Vite
// proxy in vite.config.ts forwards that to the configured collector.
// In prod, your SPA's reverse proxy / CDN must do the same — this
// avoids browser CORS hassles entirely.
//
// Disable by leaving VITE_OTEL_EXPORTER_OTLP_ENDPOINT unset; this
// module becomes a no-op (no provider registered, no instrumentation
// installed, no exporter created).
//
// IMPORTANT: import this module FIRST in main.ts so the SDK is set up
// before any code that performs fetches or starts spans.

import { ZoneContextManager } from '@opentelemetry/context-zone'
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http'
import { registerInstrumentations } from '@opentelemetry/instrumentation'
import { FetchInstrumentation } from '@opentelemetry/instrumentation-fetch'
import { resourceFromAttributes } from '@opentelemetry/resources'
import { BatchSpanProcessor, WebTracerProvider } from '@opentelemetry/sdk-trace-web'
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions'

// Endpoint + service name follow a "VITE_-prefixed wins, otherwise
// fall back to the bare OTEL_* var" convention. `vite.config.ts`'s
// envPrefix lets us read both. Default behavior: one OTEL_* var in
// .env feeds backend AND frontend. Override per side by setting the
// VITE_-prefixed counterpart.
const endpoint =
  (import.meta.env.VITE_OTEL_EXPORTER_OTLP_ENDPOINT as string | undefined) ??
  (import.meta.env.OTEL_EXPORTER_OTLP_ENDPOINT as string | undefined)

if (endpoint) {
  const serviceName =
    (import.meta.env.VITE_OTEL_SERVICE_NAME as string | undefined) ??
    (import.meta.env.OTEL_SERVICE_NAME as string | undefined) ??
    'yauth-go-vue-template-web'

  const provider = new WebTracerProvider({
    resource: resourceFromAttributes({
      [ATTR_SERVICE_NAME]: serviceName,
    }),
    spanProcessors: [
      new BatchSpanProcessor(
        new OTLPTraceExporter({
          // Same-origin path; the dev proxy / prod reverse proxy
          // forwards to the actual collector (configured by
          // VITE_OTEL_EXPORTER_OTLP_ENDPOINT in vite.config.ts).
          url: '/v1/traces',
        }),
      ),
    ],
  })

  provider.register({
    contextManager: new ZoneContextManager(),
  })

  registerInstrumentations({
    instrumentations: [
      new FetchInstrumentation({
        // Inject traceparent on every outbound fetch — same-origin
        // (Vite proxy) and cross-origin (split-origin prod). The
        // backend's CORS allowlist already permits the header.
        propagateTraceHeaderCorsUrls: [/.*/],
      }),
    ],
  })
}
