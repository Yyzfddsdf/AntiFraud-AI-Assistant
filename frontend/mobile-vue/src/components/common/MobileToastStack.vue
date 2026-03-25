<template>
  <div class="fixed top-4 left-4 right-4 z-[999] pointer-events-none flex flex-col gap-2">
    <transition-group
      enter-active-class="transition duration-300 ease-out"
      enter-from-class="transform -translate-y-4 opacity-0"
      leave-active-class="transition duration-200 ease-in"
      leave-to-class="transform -translate-y-2 opacity-0"
    >
      <div
        v-for="toast in state.toasts"
        :key="toast.id"
        :class="[
          'px-4 py-3 rounded-xl shadow-lg border flex items-center gap-3 pointer-events-auto backdrop-blur-md',
          toast.type === 'error'
            ? 'bg-white border-red-100 text-red-600'
            : toast.type === 'warning'
              ? 'bg-white border-amber-100 text-amber-700'
              : 'bg-white border-gray-100 text-black'
        ]"
      >
        <div
          :class="[
            'w-2 h-2 rounded-full',
            toast.type === 'error' ? 'bg-red-500' : toast.type === 'warning' ? 'bg-amber-500' : 'bg-black'
          ]"
        ></div>
        <span class="text-sm font-medium">{{ toast.message }}</span>
      </div>
    </transition-group>
  </div>
</template>

<script setup>
defineProps({
  state: {
    type: Object,
    required: true
  }
});
</script>
