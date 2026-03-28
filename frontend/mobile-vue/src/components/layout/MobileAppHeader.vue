<template>
  <div class="bg-white">
    <header class="px-5 py-4 flex justify-between items-center glass">
      <div class="flex items-center gap-2.5">
        <div class="w-9 h-9 bg-primary rounded-xl flex items-center justify-center text-white shadow-lg shadow-primary/20">
          <i data-lucide="shield-check" size="20"></i>
        </div>
        <div>
          <h1 class="text-[16px] font-bold text-slate-900 leading-none">反诈卫士</h1>
          <span class="text-[9px] text-slate-400 font-bold tracking-widest uppercase">Sentinel AI</span>
        </div>
      </div>

      <button type="button" class="w-9 h-9 rounded-full bg-slate-50 flex items-center justify-center text-slate-500 border border-slate-100" @click="state.activeTab = 'alerts'">
        <i data-lucide="bell" size="18"></i>
      </button>
    </header>
  </div>
</template>

<script setup>
import { nextTick, onMounted, onUpdated } from 'vue';

defineProps({
  state: {
    type: Object,
    required: true
  }
});

const refreshLucideIcons = () => {
  nextTick(() => {
    if (window.lucide && typeof window.lucide.createIcons === 'function') {
      window.lucide.createIcons();
    }
  });
};

onMounted(refreshLucideIcons);
onUpdated(refreshLucideIcons);
</script>
