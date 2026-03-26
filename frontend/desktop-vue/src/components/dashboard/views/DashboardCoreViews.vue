<template src="./DashboardCoreViews.template.html"></template>

<script>
import DesktopSimulationQuizView from './DesktopSimulationQuizView.vue';
import DesktopRiskTrendView from './DesktopRiskTrendView.vue';

export default {
  name: 'DashboardCoreViews',
  components: {
    DesktopSimulationQuizView,
    DesktopRiskTrendView
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
