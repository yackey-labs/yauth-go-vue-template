<script setup lang="ts">
import { useSession } from '@yackey-labs/yauth-ui-vue'
import { RouterLink, RouterView } from 'vue-router'

const { user, isAuthenticated } = useSession()
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <header class="border-b border-border">
      <nav class="max-w-3xl mx-auto px-6 py-3 flex items-center gap-4 text-sm">
        <RouterLink
          to="/login"
          class="text-foreground hover:underline"
          active-class="font-medium"
        >
          Login
        </RouterLink>
        <RouterLink
          to="/register"
          class="text-foreground hover:underline"
          active-class="font-medium"
        >
          Register
        </RouterLink>
        <RouterLink
          to="/dashboard"
          class="text-foreground hover:underline"
          active-class="font-medium"
        >
          Dashboard
        </RouterLink>
        <span class="ml-auto text-muted-foreground">
          <template v-if="isAuthenticated">
            Signed in as <strong class="text-foreground">{{ user?.email }}</strong>
          </template>
          <template v-else>Not signed in</template>
        </span>
      </nav>
    </header>
    <main class="flex-1 max-w-md w-full mx-auto px-6 py-10">
      <RouterView />
    </main>
  </div>
</template>
