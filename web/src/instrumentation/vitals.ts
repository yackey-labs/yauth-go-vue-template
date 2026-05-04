// Core Web Vitals → OpenTelemetry spans.
//
// Each metric (LCP / INP / CLS / FCP / TTFB) becomes a short span on
// the "web-vitals" tracer with attributes following the convention
// described in the Honeycomb / SigNoz patterns:
//   web_vital.name, .value, .delta, .rating, .id, .navigation_type
//
// We use traces rather than the metrics SDK because:
//   1. Vitals correlate 1:1 with the same session that generated the
//      surrounding traces — querying "vitals for this user" is a span
//      filter, not a separate metric stream.
//   2. We avoid pulling in @opentelemetry/sdk-metrics (~30 KB gzipped)
//      and a separate exporter pipeline.
//   3. The web_vital.value attribute is a number, so percentile
//      aggregations work in the collector or query layer.
//
// `web-vitals` (Google's library) emits each metric ONCE per page load,
// which keeps span volume low. INP replaces the deprecated FID in
// web-vitals v4+.

import { context, SpanKind, trace } from '@opentelemetry/api'
import { onCLS, onFCP, onINP, onLCP, onTTFB, type Metric } from 'web-vitals'

const tracer = trace.getTracer('web-vitals')

function report(metric: Metric) {
  // Anchor the vital span at the document-load context so backends can
  // group all five vitals under the same root request.
  const span = tracer.startSpan(
    `vital.${metric.name.toLowerCase()}`,
    { kind: SpanKind.INTERNAL },
    context.active(),
  )
  span.setAttributes({
    'web_vital.name': metric.name,
    'web_vital.value': metric.value,
    'web_vital.delta': metric.delta,
    'web_vital.rating': metric.rating,
    'web_vital.id': metric.id,
    'web_vital.navigation_type': metric.navigationType,
  })
  span.end()
}

export function installWebVitals() {
  onLCP(report)
  onINP(report)
  onCLS(report)
  onFCP(report)
  onTTFB(report)
}
