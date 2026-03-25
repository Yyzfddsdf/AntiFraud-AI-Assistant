<template>
  <div v-show="state.activeTab === 'profile'" class="h-full overflow-y-auto bg-slate-50 pt-3 pb-28" style="-webkit-overflow-scrolling: touch;">
    <div class="px-5 mb-6 mt-2 flex items-center gap-4">
      <div class="w-20 h-20 rounded-[24px] bg-white shadow-md border border-slate-100 flex items-center justify-center text-3xl font-black text-slate-300 relative">
        {{ state.user.username ? state.user.username.substring(0,1).toUpperCase() : 'U' }}
        <div class="absolute -bottom-2 -right-2 bg-emerald-500 text-white text-[10px] font-black px-2 py-0.5 rounded-lg shadow-sm border-2 border-white">{{ state.user.role || 'User' }}</div>
      </div>
      <div class="flex-1">
        <h2 class="text-2xl font-black text-slate-900">{{ state.user.username }}</h2>
        <p class="text-xs font-bold text-slate-400 mt-1">欢迎使用反诈护航 AI</p>
      </div>
    </div>

    <div class="px-4 space-y-4">
      <div class="bg-white rounded-[24px] shadow-sm border border-slate-100 overflow-hidden">
        <button type="button" @click="state.openProfilePrivacyPage" class="w-full p-5 flex justify-between items-center active:bg-slate-50 transition-colors text-left group">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 rounded-full bg-blue-50 text-blue-500 flex items-center justify-center group-hover:scale-110 transition-transform">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path></svg>
            </div>
            <div>
              <div class="text-[15px] font-black text-slate-900">隐私资料</div>
              <div class="text-xs font-medium text-slate-400 mt-0.5">查看与管理您的个人画像数据</div>
            </div>
          </div>
          <svg class="w-5 h-5 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>
        </button>
      </div>

      <button @click="state.logout" class="w-full py-4 text-center text-rose-600 font-bold bg-white rounded-[24px] shadow-sm border border-slate-100 active:bg-rose-50 transition-colors">退出登录</button>
    </div>
  </div>

  <div v-if="state.activeTab === 'profile_privacy'" class="fixed inset-0 z-[1000] bg-slate-50 flex flex-col animate-slide-up" style="padding-bottom: env(safe-area-inset-bottom);">
    <div class="shrink-0 bg-white/80 backdrop-blur-md z-10 px-4 pt-safe pb-3 flex flex-col gap-3 sticky top-0 border-b border-slate-100">
      <div class="flex items-center justify-between mt-2">
        <button @click="state.closeProfilePrivacyPage" class="w-8 h-8 rounded-full bg-slate-100 text-slate-500 flex items-center justify-center active:scale-90 transition-transform">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path></svg>
        </button>
        <div class="font-bold text-slate-800 text-base">隐私资料</div>
        <div class="w-8"></div>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto p-4 space-y-4" style="-webkit-overflow-scrolling: touch;">
      <div class="bg-white rounded-[24px] p-5 shadow-sm border border-slate-100 space-y-5">
        <div>
          <div class="text-[11px] uppercase tracking-[0.2em] text-slate-400 font-bold mb-1">手机号</div>
          <div class="text-[15px] font-black text-slate-800 bg-slate-50 rounded-xl px-4 py-3 border border-slate-100">{{ state.getUserPhoneText(state.user) || '未设置' }}</div>
        </div>
        <div>
          <div class="text-[11px] uppercase tracking-[0.2em] text-slate-400 font-bold mb-1">邮箱</div>
          <div class="text-[15px] font-black text-slate-800 bg-slate-50 rounded-xl px-4 py-3 border border-slate-100 break-all">{{ state.getUserEmailText(state.user) || '未设置' }}</div>
        </div>

        <div class="pt-2">
          <div class="flex items-start justify-between gap-3 mb-3">
            <div class="text-[11px] uppercase tracking-[0.2em] text-slate-400 font-bold">画像资料</div>
            <button type="button" @click="state.toggleAgeEditor" class="text-[11px] font-black text-emerald-600 bg-emerald-50 px-3 py-1.5 rounded-full active:scale-95 transition-transform">
              {{ state.ageEditorVisible ? '收起编辑' : '编辑资料' }}
            </button>
          </div>

          <div v-if="!state.ageEditorVisible" class="bg-slate-50 rounded-2xl p-4 border border-slate-100 space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-xs font-bold text-slate-500">年龄</span>
              <span class="text-[14px] font-black text-slate-800">{{ state.user.age || state.ageForm.age || '未设置' }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-xs font-bold text-slate-500">职业</span>
              <span class="text-[14px] font-black text-slate-800">{{ state.user.occupation || '未设置' }}</span>
            </div>
            <div class="flex justify-between items-start">
              <span class="text-xs font-bold text-slate-500">位置</span>
              <span class="text-[14px] font-black text-slate-800 text-right">{{ state.user.province_name || '未设置' }} <br> <span class="text-xs text-slate-500 font-medium">{{ state.user.city_name || '未设置' }} / {{ state.user.district_name || '未设置' }}</span></span>
            </div>
            <div class="pt-2 border-t border-slate-200/60">
              <span class="text-[10px] font-bold text-slate-400 block mb-2">近期标签</span>
              <div class="flex flex-wrap gap-2" v-if="state.user.recent_tags && state.user.recent_tags.length">
                <span v-for="tag in state.user.recent_tags" :key="`recent-tag-${tag}`" class="px-2 py-1 rounded-md bg-white border border-slate-200 text-[10px] font-black text-slate-600">{{ tag }}</span>
              </div>
              <div v-else class="text-xs text-slate-400 font-bold">暂无标签</div>
            </div>
          </div>

          <div v-if="state.ageEditorVisible" class="mt-3 space-y-4 bg-slate-50 rounded-2xl p-4 border border-slate-100">
            <div>
              <label class="text-[10px] uppercase tracking-widest font-bold text-slate-500 block mb-1.5">年龄</label>
              <input v-model.number="state.ageForm.age" type="number" min="1" max="150" class="w-full h-11 px-3 rounded-xl bg-white border border-slate-200 text-sm font-bold focus:ring-2 focus:ring-emerald-500" placeholder="输入年龄">
            </div>
            <div>
              <label class="text-[10px] uppercase tracking-widest font-bold text-slate-500 block mb-1.5">职业</label>
              <div class="m-dropdown" data-custom-dropdown>
                <button type="button" @click="state.toggleDropdown('profile_occupation')" :class="['m-dropdown-trigger', state.openDropdownKey === 'profile_occupation' ? 'is-open' : '']">
                  <div class="min-w-0">
                    <div class="text-sm font-semibold text-slate-800">{{ state.profileForm.occupation || '选择职业' }}</div>
                    <div class="text-[11px] text-slate-400 mt-1">从配置枚举中选择当前职业</div>
                  </div>
                  <svg class="w-4 h-4 text-slate-400 shrink-0 transition-transform" :class="state.openDropdownKey === 'profile_occupation' ? 'rotate-180' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
                </button>
                <div v-if="state.openDropdownKey === 'profile_occupation'" class="m-dropdown-menu hide-scrollbar">
                  <button type="button" class="m-dropdown-option" :class="{ 'is-selected': !state.profileForm.occupation }" @click="state.selectDropdownValue('profile_occupation', state.profileForm, 'occupation', '')">
                    <div>
                      <div class="text-sm font-semibold text-slate-800">未设置</div>
                      <div class="text-[11px] text-slate-400 mt-1">清空职业信息</div>
                    </div>
                  </button>
                  <button v-for="item in state.occupationOptions" :key="`occupation-${item}`" type="button" class="m-dropdown-option" :class="{ 'is-selected': state.profileForm.occupation === item }" @click="state.selectDropdownValue('profile_occupation', state.profileForm, 'occupation', item)">
                    <div class="text-sm font-semibold text-slate-800">{{ item }}</div>
                    <svg v-if="state.profileForm.occupation === item" class="w-4 h-4 text-emerald-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
                  </button>
                </div>
              </div>
            </div>
            <div class="space-y-2">
              <div class="text-[11px] uppercase tracking-[0.18em] text-slate-400 font-bold">地理位置</div>
              <div class="grid grid-cols-1 gap-2">
                <div class="m-dropdown" data-custom-dropdown>
                  <button type="button" @click="state.toggleDropdown('profile_region_province')" :class="['m-dropdown-trigger', state.openDropdownKey === 'profile_region_province' ? 'is-open' : '']">
                    <div class="min-w-0">
                      <div class="text-sm font-semibold text-slate-800">{{ state.profileForm.provinceName || '选择省' }}</div>
                      <div class="text-[11px] text-slate-400 mt-1">先选择省级行政区</div>
                    </div>
                    <svg class="w-4 h-4 text-slate-400 shrink-0 transition-transform" :class="state.openDropdownKey === 'profile_region_province' ? 'rotate-180' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
                  </button>
                  <div v-if="state.openDropdownKey === 'profile_region_province'" class="m-dropdown-menu hide-scrollbar">
                    <button v-for="item in state.provinceOptions" :key="`province-${item.code}`" type="button" class="m-dropdown-option" :class="{ 'is-selected': state.profileForm.provinceCode === item.code }" @click="state.selectProvinceValue(item.code)">
                      <div class="text-sm font-semibold text-slate-800">{{ item.name }}</div>
                      <svg v-if="state.profileForm.provinceCode === item.code" class="w-4 h-4 text-emerald-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
                    </button>
                  </div>
                </div>
                <div class="m-dropdown" data-custom-dropdown>
                  <button type="button" @click="state.profileForm.provinceCode && state.toggleDropdown('profile_region_city')" :class="['m-dropdown-trigger', state.openDropdownKey === 'profile_region_city' ? 'is-open' : '', !state.profileForm.provinceCode ? 'opacity-50' : '']" :disabled="!state.profileForm.provinceCode">
                    <div class="min-w-0">
                      <div class="text-sm font-semibold text-slate-800">{{ state.profileForm.cityName || '选择市' }}</div>
                      <div class="text-[11px] text-slate-400 mt-1">根据已选省份加载城市</div>
                    </div>
                    <svg class="w-4 h-4 text-slate-400 shrink-0 transition-transform" :class="state.openDropdownKey === 'profile_region_city' ? 'rotate-180' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
                  </button>
                  <div v-if="state.openDropdownKey === 'profile_region_city'" class="m-dropdown-menu hide-scrollbar">
                    <button v-for="item in state.cityOptions" :key="`city-${item.code}`" type="button" class="m-dropdown-option" :class="{ 'is-selected': state.profileForm.cityCode === item.code }" @click="state.selectCityValue(item.code)">
                      <div class="text-sm font-semibold text-slate-800">{{ item.name }}</div>
                      <svg v-if="state.profileForm.cityCode === item.code" class="w-4 h-4 text-emerald-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
                    </button>
                  </div>
                </div>
                <div class="m-dropdown" data-custom-dropdown>
                  <button type="button" @click="state.profileForm.cityCode && state.toggleDropdown('profile_region_district')" :class="['m-dropdown-trigger', state.openDropdownKey === 'profile_region_district' ? 'is-open' : '', !state.profileForm.cityCode ? 'opacity-50' : '']" :disabled="!state.profileForm.cityCode">
                    <div class="min-w-0">
                      <div class="text-sm font-semibold text-slate-800">{{ state.profileForm.districtName || '选择末级行政区' }}</div>
                      <div class="text-[11px] text-slate-400 mt-1">选择当前地区可用的最后一级行政区</div>
                    </div>
                    <svg class="w-4 h-4 text-slate-400 shrink-0 transition-transform" :class="state.openDropdownKey === 'profile_region_district' ? 'rotate-180' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
                  </button>
                  <div v-if="state.openDropdownKey === 'profile_region_district'" class="m-dropdown-menu hide-scrollbar">
                    <button v-for="item in state.districtOptions" :key="`district-${item.code}`" type="button" class="m-dropdown-option" :class="{ 'is-selected': state.profileForm.districtCode === item.code }" @click="state.selectDistrictValue(item.code)">
                      <div class="text-sm font-semibold text-slate-800">{{ item.name }}</div>
                      <svg v-if="state.profileForm.districtCode === item.code" class="w-4 h-4 text-emerald-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
                    </button>
                  </div>
                </div>
              </div>
              <button type="button" @click="state.requestCurrentRegion" :disabled="state.locationResolving" class="w-full text-sm font-bold text-emerald-700 bg-emerald-50 px-4 py-2.5 rounded-lg border border-emerald-100 disabled:opacity-50 disabled:cursor-not-allowed">
                {{ state.locationResolving ? '定位中...' : '自动定位上传' }}
              </button>
              <div class="text-[11px] text-slate-400">支持手动选择行政区，也支持点击按钮自动识别到末级行政区。</div>
              <div class="text-[11px] text-slate-500">当前选择：{{ state.profileForm.provinceName || '未选省级' }} / {{ state.profileForm.cityName || '未选中间层级' }} / {{ state.profileForm.districtName || '未选末级行政区' }}</div>
            </div>
            <div class="rounded-xl bg-slate-50 border border-slate-100 px-4 py-3">
              <div class="text-[11px] uppercase tracking-[0.18em] text-slate-400 font-bold">近期标签</div>
              <div v-if="state.user.recent_tags && state.user.recent_tags.length" class="mt-2 flex flex-wrap gap-2">
                <span v-for="tag in state.user.recent_tags" :key="`recent-tag-readonly-${tag}`" class="px-2.5 py-1 rounded-full bg-white border border-slate-200 text-xs font-medium text-slate-600">{{ tag }}</span>
              </div>
              <div v-else class="mt-1 text-xs text-slate-400">近期标签未设置</div>
            </div>
            <div class="flex items-center gap-2">
              <button type="button" @click="state.updateUserProfile" :disabled="state.profileSaving" class="text-sm font-bold text-white bg-emerald-600 px-4 py-2 rounded-lg whitespace-nowrap disabled:opacity-50 disabled:cursor-not-allowed">{{ state.profileSaving ? '保存中...' : '保存' }}</button>
              <button type="button" @click="state.cancelAgeEditor" class="text-sm font-bold text-slate-600 bg-slate-100 px-4 py-2 rounded-lg whitespace-nowrap">取消</button>
            </div>
          </div>
        </div>
        <div class="pt-2 border-t border-slate-100">
          <button type="button" @click="state.deleteAccount" class="w-full flex justify-between items-center text-left px-4 py-3 rounded-xl bg-red-50 text-red-600 active:bg-red-100">
            <span class="text-sm font-medium">删除账户</span>
          </button>
        </div>
      </div>
    </div>
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
