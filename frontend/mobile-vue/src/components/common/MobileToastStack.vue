<template>
  <div class="fixed left-4 right-4 z-[2000] pointer-events-none flex flex-col gap-3 items-center" style="top: max(1rem, env(safe-area-inset-top))">
    <transition-group
      enter-active-class="transition-all duration-400 ease-out"
      enter-from-class="transform -translate-y-6 opacity-0 scale-95"
      enter-to-class="transform translate-y-0 opacity-100 scale-100"
      leave-active-class="transition-all duration-300 ease-in absolute w-full"
      leave-from-class="transform translate-y-0 opacity-100 scale-100"
      leave-to-class="transform -translate-y-4 opacity-0 scale-95"
    >
      <div
        v-for="toast in state.toasts"
        :key="toast.id"
        :class="[
          'px-4 py-3.5 rounded-2xl shadow-[0_8px_30px_rgba(0,0,0,0.08)] border flex items-center gap-3 pointer-events-auto backdrop-blur-xl transition-all max-w-[90vw]',
          toast.type === 'error'
            ? 'bg-white/90 border-red-100/50 text-slate-800'
            : toast.type === 'warning'
              ? 'bg-white/90 border-amber-100/50 text-slate-800'
              : 'bg-slate-900/90 border-slate-800 text-white'
        ]"
      >
        <div
          v-if="toast.type === 'error'"
          class="w-6 h-6 rounded-full bg-red-100 flex items-center justify-center shrink-0"
        >
          <svg class="w-3.5 h-3.5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
        </div>
        <div
          v-else-if="toast.type === 'warning'"
          class="w-6 h-6 rounded-full bg-amber-100 flex items-center justify-center shrink-0"
        >
          <svg class="w-3.5 h-3.5 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
        </div>
        <div
          v-else
          class="w-6 h-6 rounded-full bg-slate-800 flex items-center justify-center shrink-0"
        >
          <svg class="w-3.5 h-3.5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"></path></svg>
        </div>
        <span class="text-[14px] font-medium tracking-wide">{{ toast.message }}</span>
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
