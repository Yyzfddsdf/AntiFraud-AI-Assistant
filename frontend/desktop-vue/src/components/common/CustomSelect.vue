<template>
  <div ref="rootRef" :class="wrapperClass">
    <button
      type="button"
      :disabled="disabled"
      :class="resolvedTriggerClass"
      @click="toggleOpen"
      @keydown.enter.prevent="toggleOpen"
      @keydown.space.prevent="toggleOpen"
      @keydown.esc.stop.prevent="closeMenu">
      <span class="min-w-0 truncate text-left">{{ selectedLabel }}</span>
      <svg class="w-4 h-4 shrink-0 transition-transform duration-200" :class="open ? 'rotate-180' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
      </svg>
    </button>

    <transition name="fade">
      <div v-if="open" :class="resolvedMenuClass">
        <button
          v-for="option in normalizedOptions"
          :key="option.key"
          type="button"
          :class="resolveOptionClass(option)"
          @click="selectOption(option)">
          <span class="truncate">{{ option.label }}</span>
          <svg v-if="isSelected(option.value)" :class="['w-4 h-4 shrink-0', resolvedSelectedIconClass]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
          </svg>
        </button>
      </div>
    </transition>
  </div>
</template>

<script>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';

function normalizeOption(option, index) {
  if (option && typeof option === 'object' && !Array.isArray(option)) {
    const value = Object.prototype.hasOwnProperty.call(option, 'value') ? option.value : '';
    const label = Object.prototype.hasOwnProperty.call(option, 'label') ? option.label : String(value ?? '');
    return {
      key: Object.prototype.hasOwnProperty.call(option, 'key') ? option.key : `${index}-${String(value)}`,
      value,
      label,
      disabled: Boolean(option.disabled)
    };
  }

  return {
    key: `${index}-${String(option ?? '')}`,
    value: option,
    label: String(option ?? ''),
    disabled: false
  };
}

export default {
  name: 'CustomSelect',
  props: {
    modelValue: {
      type: [String, Number, Boolean, null],
      default: ''
    },
    options: {
      type: Array,
      default: () => []
    },
    placeholder: {
      type: String,
      default: '请选择'
    },
    disabled: {
      type: Boolean,
      default: false
    },
    wrapperClass: {
      type: String,
      default: 'relative'
    },
    triggerClass: {
      type: String,
      default: ''
    },
    menuClass: {
      type: String,
      default: ''
    },
    optionClass: {
      type: String,
      default: ''
    },
    selectedOptionClass: {
      type: String,
      default: ''
    },
    selectedIconClass: {
      type: String,
      default: ''
    }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const open = ref(false);
    const rootRef = ref(null);

    const normalizedOptions = computed(() => props.options.map((option, index) => normalizeOption(option, index)));

    const selectedOption = computed(() => normalizedOptions.value.find(option => String(option.value) === String(props.modelValue)));

    const selectedLabel = computed(() => selectedOption.value?.label || props.placeholder);

    const resolvedTriggerClass = computed(() => {
      const base = 'w-full flex items-center justify-between gap-3 px-4 py-3 rounded-xl border border-slate-200 bg-white text-sm font-medium text-slate-700 shadow-sm hover:border-slate-300 focus:outline-none focus:ring-2 focus:ring-brand-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all';
      return `${base} ${props.triggerClass}`.trim();
    });

    const resolvedMenuClass = computed(() => {
      const base = 'absolute z-50 mt-2 w-full overflow-hidden rounded-2xl border border-slate-200 bg-white/98 backdrop-blur shadow-[0_20px_45px_rgba(15,23,42,0.14)] p-1.5';
      return `${base} ${props.menuClass}`.trim();
    });

    const isSelected = (value) => String(value) === String(props.modelValue);

    const closeMenu = () => {
      open.value = false;
    };

    const toggleOpen = () => {
      if (props.disabled) return;
      open.value = !open.value;
    };

    const selectOption = (option) => {
      if (option.disabled) return;
      emit('update:modelValue', option.value);
      closeMenu();
    };

    const resolveOptionClass = (option) => {
      const active = isSelected(option.value);
      const base = 'w-full flex items-center justify-between gap-3 rounded-xl px-3.5 py-2.5 text-sm text-left transition-colors';
      const state = option.disabled
        ? 'text-slate-300 cursor-not-allowed'
        : active
          ? 'bg-brand-50 text-brand-700'
          : 'text-slate-700 hover:bg-slate-100';
      const selectedState = active ? props.selectedOptionClass : '';
      return `${base} ${state} ${props.optionClass} ${selectedState}`.trim();
    };

    const handlePointerDown = (event) => {
      if (!rootRef.value) return;
      if (!rootRef.value.contains(event.target)) {
        closeMenu();
      }
    };

    onMounted(() => {
      document.addEventListener('mousedown', handlePointerDown);
    });

    onBeforeUnmount(() => {
      document.removeEventListener('mousedown', handlePointerDown);
    });

    return {
      open,
      rootRef,
      normalizedOptions,
      selectedLabel,
      resolvedTriggerClass,
      resolvedMenuClass,
      isSelected,
      closeMenu,
      toggleOpen,
      selectOption,
      resolveOptionClass,
      resolvedSelectedIconClass: computed(() => props.selectedIconClass || 'text-brand-600')
    };
  }
};
</script>
