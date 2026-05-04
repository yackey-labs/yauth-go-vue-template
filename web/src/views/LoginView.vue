<script setup lang="ts">
import { LoginForm } from '@yackey-labs/yauth-ui-vue'
import { RouterLink, useRouter } from 'vue-router'

const router = useRouter()
</script>

<template>
  <h1 class="text-2xl font-semibold mb-6">Sign in</h1>
  <LoginForm :on-success="() => router.push('/dashboard')" />
  <p class="mt-4 text-sm text-muted-foreground">
    No account?
    <RouterLink to="/register" class="text-foreground underline underline-offset-4">
      Register
    </RouterLink>
  </p>
</template>
