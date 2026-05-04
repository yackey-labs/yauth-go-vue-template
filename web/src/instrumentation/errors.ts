// Browser error capture → OpenTelemetry spans.
//
// Three sources cover the realistic failure modes:
//   • window.error            — uncaught synchronous throws
//   • unhandledrejection      — promise rejections that escape every catch
//   • console.error override  — explicit error logs (incl. framework noise)
//
// Vue framework errors are captured separately in main.ts via
// `app.config.errorHandler`, which has access to the failing component
// instance and richer info than the generic window.error path.
//
// Each capture starts a short span, calls span.recordException(err)
// (which adds an exception event per OTel semantic conventions), then
// sets the span status to ERROR. The span ends immediately — it's an
// event marker, not a duration measurement.
//
// We do NOT pre-emptively re-throw, suppress, or transform the error.
// The handler is purely additive: errors still bubble to the browser
// console for local debugging.

import { SpanStatusCode, trace } from '@opentelemetry/api'

const tracer = trace.getTracer('browser-errors')

function recordError(name: string, attributes: Record<string, string | number>, err: unknown) {
  const span = tracer.startSpan(name, { attributes })
  if (err instanceof Error) {
    span.recordException(err)
    span.setStatus({ code: SpanStatusCode.ERROR, message: err.message })
  } else {
    span.setStatus({ code: SpanStatusCode.ERROR, message: String(err) })
  }
  span.end()
}

export function installBrowserErrorHandlers() {
  window.addEventListener('error', (event) => {
    recordError(
      'browser.error.uncaught',
      {
        'code.filepath': event.filename ?? '',
        'code.lineno': event.lineno ?? 0,
        'code.column': event.colno ?? 0,
      },
      event.error ?? event.message,
    )
  })

  window.addEventListener('unhandledrejection', (event) => {
    recordError('browser.error.unhandled_rejection', {}, event.reason)
  })

  // console.error override — keep the original behaviour intact and
  // tee a span for each call. We purposely don't wrap console.warn /
  // .info / .debug; .error is the signal the developer chose to log
  // as a problem.
  const original = console.error.bind(console)
  console.error = (...args: unknown[]) => {
    original(...args)
    const err = args.find((a) => a instanceof Error) as Error | undefined
    recordError(
      'browser.error.console',
      { 'console.message': args.map(String).join(' ') },
      err ?? args.map(String).join(' '),
    )
  }
}
