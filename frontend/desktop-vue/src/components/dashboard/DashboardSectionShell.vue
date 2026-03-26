<template>
  <div class="w-full h-full min-h-0 min-w-0 animate-fade-in">
    <div class="grid h-full min-h-0 grid-cols-1 gap-4 lg:grid-cols-[240px_minmax(0,1fr)]">
      <aside class="min-h-0 overflow-hidden rounded-sm border border-slate-200 bg-white shadow-sm lg:flex lg:flex-col">
        <div class="shrink-0 border-b border-slate-200 px-4 py-3">
          <div class="flex items-center gap-2 text-brand-700">
            <span class="w-1.5 h-7 bg-brand-500 rounded-sm"></span>
            <div>
              <div class="text-xs font-bold uppercase tracking-[0.24em] text-slate-400">{{ eyebrow }}</div>
              <h1 class="text-base font-extrabold tracking-tight text-slate-900 mt-1">{{ title }}</h1>
            </div>
          </div>
        </div>

        <nav class="space-y-1 p-2.5 lg:flex-1">
          <slot name="nav" />
        </nav>
      </aside>

      <div class="min-h-0 overflow-hidden rounded-sm border border-slate-200 bg-white shadow-sm">
        <div class="custom-scrollbar flex h-full min-h-0 flex-col overflow-y-auto bg-slate-100/60 p-3">
          <div class="shrink-0 rounded-sm border border-slate-200 bg-white p-3 shadow-sm">
            <div class="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
              <div>
                <div class="flex flex-wrap items-center gap-2">
                  <span class="px-2 py-0.5 rounded-sm bg-slate-100 border border-slate-200 text-slate-700 text-[11px] font-bold uppercase tracking-[0.2em]">{{ title }}</span>
                  <slot name="badges" />
                </div>
                <h2 v-if="heading" class="text-lg font-extrabold text-slate-900 tracking-tight mt-1.5">{{ heading }}</h2>
                <p v-if="description" class="text-xs text-slate-500 mt-1">{{ description }}</p>
              </div>

              <div class="flex flex-wrap items-center gap-3">
                <slot name="toolbar" />
              </div>
            </div>
          </div>

          <div class="min-h-0 flex-1 pt-3">
            <slot />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'DashboardSectionShell',
  props: {
    eyebrow: {
      type: String,
      default: ''
    },
    title: {
      type: String,
      required: true
    },
    heading: {
      type: String,
      required: true
    },
    description: {
      type: String,
      default: ''
    }
  }
};
</script>
