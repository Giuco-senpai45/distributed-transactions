<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { userService } from '@/services/api'

const router = useRouter()
const currentUser = ref(null)

const logout = async () => {
  await userService.logout()
  router.push('/login')
}

onMounted(() => {
  currentUser.value = userService.getCurrentUser()
})
</script>

<template>
  <header>
    <img alt="Vue logo" class="logo" src="@/assets/logo.svg" width="125" height="125" />

    <div class="wrapper">
      <nav v-if="currentUser">
        <RouterLink to="/users" class="text-white">Users</RouterLink>
        <RouterLink to="/users/create" class="text-white">Create User</RouterLink>
        <RouterLink to="/accounts" class="text-white">Accounts</RouterLink>
        <RouterLink to="/accounts/create" class="text-white">Create Account</RouterLink>
        <RouterLink to="/accounts/deposit" class="text-white">Make Deposit</RouterLink>
        <RouterLink to="/audits" class="text-white">Audit Logs</RouterLink>
        <a href="#" @click.prevent="logout" class="text-white">Logout</a>
      </nav>
      <div v-if="currentUser" class="mt-2 text-white">
        Logged in as: {{ currentUser.username }}
      </div>
    </div>
  </header>

  <RouterView />
</template>

<style scoped>
header {
  line-height: 1.5;
  max-height: 100vh;
}

.logo {
  display: block;
  margin: 0 auto 2rem;
}

nav {
  width: 100%;
  font-size: 12px;
  text-align: center;
  margin-top: 2rem;
}

nav a.router-link-exact-active {
  color: var(--color-text);
}

nav a.router-link-exact-active:hover {
  background-color: transparent;
}

nav a {
  display: inline-block;
  padding: 0 1rem;
  border-left: 1px solid var(--color-border);
}

nav a:first-of-type {
  border: 0;
}

@media (min-width: 1024px) {
  header {
    display: flex;
    place-items: center;
    padding-right: calc(var(--section-gap) / 2);
  }

  .logo {
    margin: 0 2rem 0 0;
  }

  header .wrapper {
    display: flex;
    place-items: flex-start;
    flex-wrap: wrap;
  }

  nav {
    text-align: left;
    margin-left: -1rem;
    font-size: 1rem;

    padding: 1rem 0;
    margin-top: 1rem;
  }
}
</style>
