<template>
  <svg
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-linecap="round"
    stroke-linejoin="round"
    :stroke-width="strokeWidth"
    aria-hidden="true"
    v-bind="$attrs"
  >
    <template v-if="name === 'signal'">
      <path d="M18 20v-10"></path>
      <path d="M14 20v-6"></path>
      <path d="M10 20v-3"></path>
      <path d="M6 20v-1"></path>
    </template>

    <template v-else-if="name === 'wifi'">
      <path d="M5 12.55a11 11 0 0 1 14.08 0"></path>
      <path d="M8.53 16.11a6 6 0 0 1 6.95 0"></path>
      <path d="M12 20h.01"></path>
    </template>

    <template v-else-if="name === 'battery'">
      <rect x="2" y="7" width="18" height="10" rx="2"></rect>
      <path d="M22 11v2"></path>
      <path d="M6 10h8v4H6z"></path>
    </template>

    <template v-else-if="name === 'shield-check'">
      <path d="M12 3l7 4v5c0 4.5-2.8 8.4-7 9-4.2-.6-7-4.5-7-9V7l7-4z"></path>
      <path d="m9 12 2 2 4-4"></path>
    </template>

    <template v-else-if="name === 'bell'">
      <path d="M10.27 21a2 2 0 0 0 3.46 0"></path>
      <path d="M18 8a6 6 0 1 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9"></path>
    </template>

    <template v-else-if="name === 'check'">
      <path d="M20 6 9 17l-5-5"></path>
    </template>

    <template v-else-if="name === 'bot'">
      <path d="M12 8V4"></path>
      <path d="M9 4h6"></path>
      <rect x="4" y="8" width="16" height="11" rx="3"></rect>
      <path d="M8 12h.01"></path>
      <path d="M16 12h.01"></path>
      <path d="M9 16h6"></path>
    </template>

    <template v-else-if="name === 'calendar'">
      <rect x="3" y="4" width="18" height="18" rx="2"></rect>
      <path d="M16 2v4"></path>
      <path d="M8 2v4"></path>
      <path d="M3 10h18"></path>
    </template>

    <template v-else-if="name === 'line-chart'">
      <path d="M3 3v18h18"></path>
      <path d="m19 9-5 5-4-4-3 3"></path>
    </template>

    <template v-else-if="name === 'heart'">
      <path d="m12 20-1.45-1.32C5.4 14.05 2 10.97 2 7.2 2 4.12 4.42 2 7.5 2c1.74 0 3.41.81 4.5 2.09C13.09 2.81 14.76 2 16.5 2 19.58 2 22 4.12 22 7.2c0 3.77-3.4 6.85-8.55 11.49z"></path>
    </template>

    <template v-else-if="name === 'play-circle'">
      <circle cx="12" cy="12" r="10"></circle>
      <path d="m10 8 6 4-6 4z"></path>
    </template>

    <template v-else-if="name === 'search'">
      <circle cx="11" cy="11" r="7"></circle>
      <path d="m21 21-4.35-4.35"></path>
    </template>

    <template v-else-if="name === 'alert-octagon'">
      <path d="M7.86 2h8.28L22 7.86v8.28L16.14 22H7.86L2 16.14V7.86z"></path>
      <path d="M12 8v4"></path>
      <path d="M12 16h.01"></path>
    </template>

    <template v-else-if="name === 'inbox'">
      <path d="M22 12h-6l-2 3h-4l-2-3H2"></path>
      <path d="M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"></path>
    </template>

    <template v-else-if="name === 'home'">
      <path d="m3 10 9-7 9 7"></path>
      <path d="M5 10v10h14V10"></path>
      <path d="M9 20v-6h6v6"></path>
    </template>

    <template v-else-if="name === 'message-square'">
      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
    </template>

    <template v-else-if="name === 'maximize'">
      <path d="M8 3H3v5"></path>
      <path d="M16 3h5v5"></path>
      <path d="M8 21H3v-5"></path>
      <path d="M16 21h5v-5"></path>
    </template>

    <template v-else-if="name === 'users'">
      <path d="M16 21v-2a4 4 0 0 0-4-4H7a4 4 0 0 0-4 4v2"></path>
      <circle cx="9.5" cy="7" r="4"></circle>
      <path d="M22 21v-2a4 4 0 0 0-3-3.87"></path>
      <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
    </template>
  </svg>
</template>

<script setup>
defineProps({
  name: {
    type: String,
    required: true
  },
  size: {
    type: [Number, String],
    default: 24
  },
  strokeWidth: {
    type: [Number, String],
    default: 2
  }
});
</script>
