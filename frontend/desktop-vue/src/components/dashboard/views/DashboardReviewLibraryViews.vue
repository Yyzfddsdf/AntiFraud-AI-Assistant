<template src="./DashboardReviewLibraryViews.template.html"></template>

<script>
import CustomSelect from '../../common/CustomSelect.vue';

export default {
  name: 'DashboardReviewLibraryViews',
  components: {
    CustomSelect
  },
  props: {
    app: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    return props.app;
  }
};
</script>

<style scoped>
.case-download-trigger {
  color: rgba(51, 65, 85, 0.82);
  padding: 2px;
  display: inline-flex;
  justify-content: center;
  align-items: center;
  background: transparent;
  overflow: visible;
  font-size: 0.8rem;
  font-weight: 600;
  gap: 0;
  border-radius: 4px;
  margin: 0;
  transition: 0.2s;
  border: 0;
  box-shadow: none;
}

.case-download-trigger:hover {
  color: rgb(79, 70, 229);
  transform: translateY(-2px) scale(1.04);
}

.case-download-trigger:active {
  transform: translateY(0) scale(0.98);
  box-shadow: none;
}

.case-download-trigger:focus-visible {
  outline: 2px solid rgba(99, 102, 241, 0.48);
  outline-offset: 3px;
}
</style>
