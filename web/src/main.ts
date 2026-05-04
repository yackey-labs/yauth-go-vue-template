// Initialize OpenTelemetry FIRST so FetchInstrumentation is in place
// before any other module fires off network requests, and so the
// global TracerProvider exists by the time the vitals/error
// instrumentation registers spans.
import './otel'

import { SpanStatusCode, trace } from '@opentelemetry/api'
import { YAuthPlugin } from '@yackey-labs/yauth-ui-vue'
import { createApp } from 'vue'
import App from './App.vue'
import { installBrowserErrorHandlers } from './instrumentation/errors'
import { installWebVitals } from './instrumentation/vitals'
import { router } from './router'
import './style.css'

installWebVitals()
installBrowserErrorHandlers()

const app = createApp(App)

// Vue framework errors — caught from inside any component's render,
// lifecycle hook, or watcher. Recorded as an OTel exception with the
// component name + lifecycle hook (`info`) for downstream querying.
app.config.errorHandler = (err, instance, info) => {
  const span = trace.getTracer('browser-errors').startSpan('browser.error.vue', {
    attributes: {
      'vue.component': instance?.$options?.name ?? 'unknown',
      'vue.info': info,
    },
  })
  if (err instanceof Error) {
    span.recordException(err)
    span.setStatus({ code: SpanStatusCode.ERROR, message: err.message })
  } else {
    span.setStatus({ code: SpanStatusCode.ERROR, message: String(err) })
  }
  span.end()
  // Re-throw in dev so the browser shows the unhandled error, hidden
  // in prod so the user doesn't see a stacktrace.
  if (import.meta.env.DEV) {
    throw err
  }
}

app
  .use(router)
  .use(YAuthPlugin, { baseUrl: '/api/auth' })
  .mount('#app')
