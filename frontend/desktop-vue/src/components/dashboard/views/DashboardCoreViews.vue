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

<style scoped>
.analysis-upload-card {
  position: relative;
  height: 200px;
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 18px;
  cursor: pointer;
  border: 2px dashed #cacaca;
  background-color: rgba(255, 255, 255, 1);
  padding: 1.5rem;
  border-radius: 10px;
  box-shadow: 0 48px 35px -48px rgba(0, 0, 0, 0.1);
  transition: all 0.25s ease;
}

.analysis-upload-card:hover {
  transform: translateY(-2px);
  border-color: #bdbdbd;
  box-shadow: 0 52px 38px -48px rgba(0, 0, 0, 0.14);
}

.analysis-upload-card input {
  display: none;
}

.analysis-upload-card__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.analysis-upload-card__icon svg {
  width: 84px;
  height: 84px;
  fill: rgba(75, 85, 99, 1);
}

.analysis-upload-card__marker {
  position: absolute;
  right: -4px;
  bottom: 4px;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: #ffffff;
  border: 1px solid #d7dbe2;
  box-shadow: 0 8px 18px rgba(0, 0, 0, 0.08);
}

.analysis-upload-card__marker svg {
  width: 16px;
  height: 16px;
  fill: none;
  stroke: rgba(75, 85, 99, 1);
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.analysis-upload-card__text {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.analysis-upload-card__text span {
  font-weight: 500;
  color: rgba(75, 85, 99, 1);
}
</style>
