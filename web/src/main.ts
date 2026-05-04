// Initialize OpenTelemetry FIRST so FetchInstrumentation is in place
// before any other module fires off network requests.
import './otel'

import { YAuthPlugin } from '@yackey-labs/yauth-ui-vue'
import { createApp } from 'vue'
import App from './App.vue'
import { router } from './router'
import './style.css'

createApp(App)
  .use(router)
  .use(YAuthPlugin, { baseUrl: '/api/auth' })
  .mount('#app')
