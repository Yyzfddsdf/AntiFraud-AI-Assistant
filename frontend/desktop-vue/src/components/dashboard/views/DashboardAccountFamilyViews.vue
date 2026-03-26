<template src="./DashboardAccountFamilyViews.template.html"></template>

<script>
import { computed, ref } from 'vue';
import CustomSelect from '../../common/CustomSelect.vue';
import DashboardSectionShell from '../DashboardSectionShell.vue';

export default {
  name: 'DashboardAccountFamilyViews',
  components: {
    CustomSelect,
    DashboardSectionShell
  },
  props: {
    app: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    const profileCenterSection = ref('basic');
    const setProfileCenterSection = (nextSection) => {
      const normalized = String(nextSection || '').trim().toLowerCase();
      if (normalized === 'account' && props.app.user?.role === 'admin') {
        profileCenterSection.value = 'basic';
        return;
      }
      profileCenterSection.value = ['basic', 'account', 'danger'].includes(normalized) ? normalized : 'basic';
    };
    const profileHeading = computed(() => {
      if (profileCenterSection.value === 'account') return '账户能力与权限';
      if (profileCenterSection.value === 'danger') return '安全操作';
      return '基本资料';
    });
    const profileDescription = computed(() => {
      if (profileCenterSection.value === 'account') return '处理管理员升级相关设置。';
      if (profileCenterSection.value === 'danger') return '执行高风险账户操作。';
      return '维护年龄、职业、标签和地区信息。';
    });

    return {
      ...props.app,
      profileCenterSection,
      setProfileCenterSection,
      profileHeading,
      profileDescription
    };
  }
};
</script>
