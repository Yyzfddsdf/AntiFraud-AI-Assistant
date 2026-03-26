<template>
  <div class="w-full h-full min-w-0 animate-fade-in">
    <div class="rounded-sm border border-slate-200 bg-white shadow-sm overflow-hidden h-full">
      <div class="flex flex-col lg:flex-row h-full">
        <aside class="w-full lg:w-56 lg:h-full shrink-0 border-b lg:border-b-0 lg:border-r border-slate-200 bg-white overflow-hidden">
          <div class="p-4 border-b border-slate-200">
            <div class="flex items-center gap-2 text-brand-700">
              <span class="w-1.5 h-7 bg-brand-500 rounded-sm"></span>
              <div>
                <div class="text-xs font-bold uppercase tracking-[0.24em] text-slate-400">{{ eyebrow }}</div>
                <h1 class="text-lg font-extrabold tracking-tight text-slate-900 mt-1">{{ title }}</h1>
              </div>
            </div>
          </div>

          <nav class="p-2.5 space-y-1">
            <slot name="nav" />
          </nav>
        </aside>

        <div class="min-w-0 flex-1 bg-slate-100/60 p-4 lg:h-full lg:overflow-y-auto flex flex-col gap-4">
          <div class="bg-white rounded-sm border border-slate-200 shadow-sm p-4">
            <div class="flex flex-col xl:flex-row xl:items-center xl:justify-between gap-3">
              <div>
                <div class="flex flex-wrap items-center gap-2">
                  <span class="px-2 py-0.5 rounded-sm bg-slate-100 border border-slate-200 text-slate-700 text-[11px] font-bold uppercase tracking-[0.2em]">{{ title }}</span>
                  <slot name="badges" />
                </div>
                <h2 class="text-xl font-extrabold text-slate-900 tracking-tight mt-2">{{ heading }}</h2>
                <p class="text-sm text-slate-500 mt-1">{{ description }}</p>
              </div>

              <div class="flex flex-wrap items-center gap-3">
                <slot name="toolbar" />
              </div>
            </div>
          </div>

          <slot />
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
