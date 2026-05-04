<script setup lang="ts">
import { RegisterForm } from '@yackey-labs/yauth-ui-vue'
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

const router = useRouter()
const successMessage = ref<string | null>(null)

function handleSuccess(message: string) {
  successMessage.value = message || 'Account created. Redirecting to login...'
  window.setTimeout(() => router.push('/login'), 1500)
}
</script>

<template>
  <h1 class="text-2xl font-semibold mb-6">Create account</h1>
  <div
    v-if="successMessage"
    class="rounded-md bg-primary/5 border border-primary/20 px-3 py-2 text-sm text-foreground mb-4"
  >
    {{ successMessage }}
  </div>
  <RegisterForm v-else :on-success="handleSuccess" />
  <p class="mt-4 text-sm text-muted-foreground">
    Already have one?
    <RouterLink to="/login" class="text-foreground underline underline-offset-4">
      Sign in
    </RouterLink>
  </p>
</template>
