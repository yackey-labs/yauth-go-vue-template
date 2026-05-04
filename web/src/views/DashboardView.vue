<script setup lang="ts">
import { useSession } from '@yackey-labs/yauth-ui-vue'
import { onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

const { user, loading, isAuthenticated, logout } = useSession()
const router = useRouter()

interface Me {
  id: string
  email: string
  role: string
  email_verified: boolean
  auth_method: string
}

const me = ref<Me | null>(null)
const meError = ref<string | null>(null)

// Demonstrates how to call your own protected route. The session cookie is
// included automatically because credentials: 'include' is the default for
// same-origin requests through the Vite proxy.
async function loadMe() {
  meError.value = null
  try {
    const r = await fetch('/api/me', { credentials: 'include' })
    if (!r.ok) {
      meError.value = `GET /api/me → ${r.status}`
      return
    }
    me.value = (await r.json()) as Me
  } catch (e) {
    meError.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(() => {
  if (isAuthenticated.value) {
    void loadMe()
  }
})

async function handleLogout() {
  await logout()
  router.push('/login')
}
</script>

<template>
  <h1 class="text-2xl font-semibold mb-6">Dashboard</h1>

  <div v-if="loading" class="text-muted-foreground">Loading session...</div>

  <div v-else-if="isAuthenticated" class="space-y-6">
    <section>
      <h2 class="text-sm font-medium text-muted-foreground mb-2">
        useSession()
      </h2>
      <p>
        Hello,
        <strong>{{ user?.email }}</strong>
      </p>
      <p class="text-sm text-muted-foreground">
        Role: {{ user?.role }} ·
        Email verified: {{ user?.email_verified ? 'yes' : 'no' }}
      </p>
    </section>

    <section>
      <h2 class="text-sm font-medium text-muted-foreground mb-2">
        Protected fetch — <code>GET /api/me</code>
      </h2>
      <pre
        v-if="me"
        class="bg-muted px-3 py-2 rounded-md text-xs overflow-x-auto"
      >{{ JSON.stringify(me, null, 2) }}</pre>
      <p v-else-if="meError" class="text-sm text-destructive">{{ meError }}</p>
      <p v-else class="text-sm text-muted-foreground">Loading...</p>
    </section>

    <button
      type="button"
      class="rounded-md bg-primary px-3 py-2 text-sm text-primary-foreground hover:opacity-90"
      @click="handleLogout"
    >
      Logout
    </button>
  </div>

  <div v-else class="text-sm text-muted-foreground">
    Not signed in.
    <RouterLink to="/login" class="text-foreground underline underline-offset-4">
      Go to login
    </RouterLink>
  </div>
</template>
