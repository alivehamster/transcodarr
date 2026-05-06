<template>
  <nav class="flex items-center gap-4 border-b border-gray-200 bg-white px-6 py-3 shadow-sm">
    <RouterLink
      to="/"
      class="text-sm font-medium text-gray-600 hover:text-gray-900"
      active-class="text-blue-600"
    >
      Home
    </RouterLink>
    <RouterLink
      to="/history"
      class="text-sm font-medium text-gray-600 hover:text-gray-900"
      active-class="text-blue-600"
    >
      History
    </RouterLink>
  </nav>
</template>
