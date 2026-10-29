<script setup lang="ts">
import NavBar from '@/components/NavBar.vue'
</script>

<template>
  <div class="flex flex-col h-screen">
    <NavBar />
    <RouterView />
  </div>
</template>
