<template>
  <div v-show="state.activeTab === 'profile'" class="bg-slate-50 pt-3 pb-32">
    <!-- Profile Header Card -->
    <div class="px-4 mb-6 mt-4">
      <div class="relative bg-white rounded-[32px] p-6 shadow-[0_8px_30px_rgb(0,0,0,0.04)] border border-slate-100/50 overflow-hidden">
        <!-- Decorative Background Elements -->
        <div class="absolute top-0 right-0 w-32 h-32 bg-gradient-to-br from-emerald-50 to-emerald-100/30 rounded-full blur-2xl -mr-10 -mt-10"></div>
        <div class="absolute bottom-0 left-0 w-24 h-24 bg-gradient-to-tr from-blue-50 to-blue-100/30 rounded-full blur-xl -ml-8 -mb-8"></div>
        
        <div class="relative flex items-center gap-5">
          <div class="relative">
            <div class="w-20 h-20 rounded-[24px] bg-gradient-to-br from-slate-100 to-slate-50 shadow-inner flex items-center justify-center text-3xl font-black text-slate-700 border-2 border-white ring-1 ring-slate-100">
              {{ state.user.username ? state.user.username.substring(0,1).toUpperCase() : 'U' }}
            </div>
            <div class="absolute -bottom-1 -right-1 bg-gradient-to-r from-emerald-500 to-emerald-400 text-white text-[10px] font-black px-2.5 py-0.5 rounded-lg shadow-sm border-2 border-white tracking-wider">
              {{ state.user.role || 'User' }}
            </div>
          </div>
          <div class="flex-1">
            <h2 class="text-2xl font-black text-slate-800 tracking-tight">{{ state.user.username }}</h2>
            <div class="flex items-center gap-1.5 mt-1.5">
              <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
              <p class="text-[13px] font-bold text-slate-400">欢迎使用反诈护航 AI</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Menu List -->
    <div class="px-4 space-y-4">
      <!-- Privacy Profile Button -->
      <div class="bg-white rounded-[28px] shadow-[0_4px_20px_rgb(0,0,0,0.03)] border border-slate-100/50 overflow-hidden transition-all active:scale-[0.98]">
        <button type="button" @click="state.openProfilePrivacyPage" class="w-full p-5 flex justify-between items-center bg-white active:bg-slate-50/50 transition-colors text-left group">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 rounded-[18px] bg-gradient-to-br from-blue-50 to-blue-100/50 text-blue-500 flex items-center justify-center group-hover:scale-105 transition-transform shadow-inner">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path></svg>
            </div>
            <div>
              <div class="text-[16px] font-black text-slate-800 tracking-tight">隐私资料</div>
              <div class="text-[13px] font-medium text-slate-400 mt-0.5">查看与管理您的个人画像数据</div>
            </div>
          </div>
          <div class="w-8 h-8 rounded-full bg-slate-50 flex items-center justify-center text-slate-400 group-active:bg-slate-100 transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M9 5l7 7-7 7"></path></svg>
          </div>
        </button>
      </div>

      <!-- Logout Button -->
      <button @click="state.logout" class="w-full py-4 text-center text-[15px] text-rose-500 font-bold bg-white rounded-[24px] shadow-[0_4px_20px_rgb(0,0,0,0.03)] border border-slate-100/50 active:bg-rose-50 active:scale-[0.98] transition-all">
        退出登录
      </button>
    </div>
  </div>

  <!-- Privacy Settings Modal/Page -->
  <div v-if="state.activeTab === 'profile_privacy'" class="fixed inset-0 z-[1000] bg-slate-50 flex flex-col animate-slide-up" style="padding-bottom: env(safe-area-inset-bottom);">
    <!-- Header -->
    <div class="shrink-0 bg-white/80 backdrop-blur-xl z-10 px-4 pt-safe pb-3 flex flex-col gap-3 sticky top-0 border-b border-slate-100/50 shadow-sm">
      <div class="flex items-center justify-between mt-2">
        <button @click="state.closeProfilePrivacyPage" class="w-10 h-10 rounded-full bg-slate-100/80 text-slate-600 flex items-center justify-center active:scale-90 transition-transform">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M15 19l-7-7 7-7"></path></svg>
        </button>
        <div class="font-black tracking-tight text-slate-800 text-[17px]">隐私资料</div>
        <div class="w-10"></div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-4 space-y-4" style="-webkit-overflow-scrolling: touch;">
      <div class="bg-white rounded-[32px] p-6 shadow-[0_8px_30px_rgb(0,0,0,0.04)] border border-slate-100/50 space-y-6">
        
        <!-- Basic Info -->
        <div class="space-y-4">
          <div>
            <div class="text-[11px] uppercase tracking-widest text-slate-400 font-bold mb-1.5 ml-1">手机号</div>
            <div class="text-[15px] font-black text-slate-700 bg-slate-50/80 rounded-2xl px-5 py-3.5 border border-slate-100">{{ state.getUserPhoneText(state.user) || '未设置' }}</div>
          </div>
          <div>
            <div class="text-[11px] uppercase tracking-widest text-slate-400 font-bold mb-1.5 ml-1">邮箱</div>
            <div class="text-[15px] font-black text-slate-700 bg-slate-50/80 rounded-2xl px-5 py-3.5 border border-slate-100 break-all">{{ state.getUserEmailText(state.user) || '未设置' }}</div>
          </div>
        </div>

        <!-- Profile Info -->
        <div class="pt-4 border-t border-slate-100">
          <div class="flex items-center justify-between gap-3 mb-4">
            <div class="text-[13px] tracking-wide text-slate-800 font-black">画像资料</div>
            <button type="button" @click="state.toggleAgeEditor" class="text-[12px] font-bold text-emerald-600 bg-emerald-50 px-4 py-1.5 rounded-full active:scale-95 transition-transform">
              {{ state.ageEditorVisible ? '收起编辑' : '编辑资料' }}
            </button>
          </div>

          <!-- Read-only View -->
          <div v-if="!state.ageEditorVisible" class="bg-slate-50/80 rounded-[24px] p-5 border border-slate-100 space-y-4">
            <div class="flex justify-between items-center">
              <span class="text-[13px] font-bold text-slate-500">年龄</span>
              <span class="text-[15px] font-black text-slate-800">{{ state.user.age || state.ageForm.age || '未设置' }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-[13px] font-bold text-slate-500">职业</span>
              <span class="text-[15px] font-black text-slate-800">{{ state.user.occupation || '未设置' }}</span>
            </div>
            <div class="flex justify-between items-start">
              <span class="text-[13px] font-bold text-slate-500 pt-0.5">位置</span>
              <span class="text-[15px] font-black text-slate-800 text-right">{{ state.user.province_name || '未设置' }} <br> <span class="text-[12px] text-slate-400 font-bold mt-1 inline-block">{{ state.user.city_name || '未设置' }} / {{ state.user.district_name || '未设置' }}</span></span>
            </div>
            <div class="pt-4 border-t border-slate-200/60 mt-2">
              <span class="text-[11px] font-bold text-slate-400 block mb-2.5">近期标签</span>
              <div class="flex flex-wrap gap-2" v-if="state.user.recent_tags && state.user.recent_tags.length">
                <span v-for="tag in state.user.recent_tags" :key="`recent-tag-${tag}`" class="px-2.5 py-1 rounded-lg bg-white shadow-sm border border-slate-100 text-[11px] font-black text-slate-600">{{ tag }}</span>
              </div>
              <div v-else class="text-xs text-slate-400 font-bold">暂无标签</div>
            </div>
          </div>

          <!-- Edit View -->
          <div v-if="state.ageEditorVisible" class="mt-4 space-y-5 bg-slate-50/80 rounded-[24px] p-5 border border-slate-100 animate-fade-in">
            <!-- Form fields remain mostly the same structurally but with updated styling -->
            <div>
              <label class="text-[11px] uppercase tracking-widest font-bold text-slate-500 block mb-2 ml-1">年龄</label>
              <input v-model.number="state.ageForm.age" type="number" min="1" max="150" class="w-full h-12 px-4 rounded-xl bg-white border border-slate-200 text-[15px] font-bold text-slate-800 focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 transition-all shadow-sm" placeholder="输入年龄">
            </div>
            
            <div>
              <label class="text-[11px] uppercase tracking-widest font-bold text-slate-500 block mb-2 ml-1">职业</label>
              <div class="m-dropdown" data-custom-dropdown>
                <button type="button" @click="state.toggleDropdown('profile_occupation')" :class="['m-dropdown-trigger h-auto py-3 px-4 rounded-xl bg-white border border-slate-200 shadow-sm', state.openDropdownKey === 'profile_occupation' ? 'is-open ring-2 ring-emerald-500 border-emerald-500' : '']">
                  <div class="min-w-0 text-left">
                    <div class="text-[15px] font-black text-slate-800">{{ state.profileForm.occupation || '选择职业' }}</div>
                    <div class="text-[11px] font-bold text-slate-400 mt-0.5">从配置枚举中选择当前职业</div>
                  </div>
                  <svg class="w-5 h-5 text-slate-400 shrink-0 transition-transform" :class="state.openDropdownKey === 'profile_occupation' ? 'rotate-180 text-emerald-500' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
                </button>
                <div v-if="state.openDropdownKey === 'profile_occupation'" class="m-dropdown-menu hide-scrollbar rounded-xl shadow-lg border border-slate-100 mt-2">
                  <button type="button" class="m-dropdown-option py-3" :class="{ 'bg-emerald-50/50': !state.profileForm.occupation }" @click="state.selectDropdownValue('profile_occupation', state.profileForm, 'occupation', '')">
                    <div class="text-left">
                      <div class="text-[14px] font-black text-slate-800">未设置</div>
                      <div class="text-[11px] font-bold text-slate-400 mt-0.5">清空职业信息</div>
                    </div>
                  </button>
                  <button v-for="item in state.occupationOptions" :key="`occupation-${item}`" type="button" class="m-dropdown-option py-3" :class="{ 'bg-emerald-50/50': state.profileForm.occupation === item }" @click="state.selectDropdownValue('profile_occupation', state.profileForm, 'occupation', item)">
                    <div class="text-[14px] font-black text-slate-800">{{ item }}</div>
                    <svg v-if="state.profileForm.occupation === item" class="w-5 h-5 text-emerald-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"></path></svg>
                  </button>
                </div>
              </div>
            </div>
            
            <div class="space-y-3 pt-2">
              <label class="text-[11px] uppercase tracking-widest font-bold text-slate-500 block ml-1">地理位置</label>
              
              <!-- Location Selectors -->
              <div class="grid grid-cols-1 gap-3">
                <!-- Province -->
                <div class="m-dropdown" data-custom-dropdown>
                  <button type="button" @click="state.toggleDropdown('profile_region_province')" :class="['m-dropdown-trigger h-auto py-3 px-4 rounded-xl bg-white border border-slate-200 shadow-sm', state.openDropdownKey === 'profile_region_province' ? 'is-open ring-2 ring-emerald-500 border-emerald-500' : '']">
                    <div class="min-w-0 text-left">
                      <div class="text-[14px] font-black text-slate-800">{{ state.profileForm.provinceName || '选择省' }}</div>
                    </div>
                    <svg class="w-4 h-4 text-slate-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
                  </button>
                  <div v-if="state.openDropdownKey === 'profile_region_province'" class="m-dropdown-menu hide-scrollbar rounded-xl shadow-lg border border-slate-100 mt-1 max-h-48">
                    <button v-for="item in state.provinceOptions" :key="`province-${item.code}`" type="button" class="m-dropdown-option py-2.5" :class="{ 'bg-emerald-50/50': state.profileForm.provinceCode === item.code }" @click="state.selectProvinceValue(item.code)">
                      <div class="text-[14px] font-black text-slate-800">{{ item.name }}</div>
                      <svg v-if="state.profileForm.provinceCode === item.code" class="w-4 h-4 text-emerald-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"></path></svg>
                    </button>
                  </div>
                </div>
                
                <!-- City -->
                <div class="m-dropdown" data-custom-dropdown>
                  <button type="button" @click="state.profileForm.provinceCode && state.toggleDropdown('profile_region_city')" :class="['m-dropdown-trigger h-auto py-3 px-4 rounded-xl bg-white border border-slate-200 shadow-sm', state.openDropdownKey === 'profile_region_city' ? 'is-open ring-2 ring-emerald-500 border-emerald-500' : '', !state.profileForm.provinceCode ? 'opacity-50 bg-slate-50' : '']" :disabled="!state.profileForm.provinceCode">
                    <div class="min-w-0 text-left">
                      <div class="text-[14px] font-black text-slate-800">{{ state.profileForm.cityName || '选择市' }}</div>
                    </div>
                    <svg class="w-4 h-4 text-slate-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
                  </button>
                  <div v-if="state.openDropdownKey === 'profile_region_city'" class="m-dropdown-menu hide-scrollbar rounded-xl shadow-lg border border-slate-100 mt-1 max-h-48">
                    <button v-for="item in state.cityOptions" :key="`city-${item.code}`" type="button" class="m-dropdown-option py-2.5" :class="{ 'bg-emerald-50/50': state.profileForm.cityCode === item.code }" @click="state.selectCityValue(item.code)">
                      <div class="text-[14px] font-black text-slate-800">{{ item.name }}</div>
                      <svg v-if="state.profileForm.cityCode === item.code" class="w-4 h-4 text-emerald-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"></path></svg>
                    </button>
                  </div>
                </div>
                
                <!-- District -->
                <div class="m-dropdown" data-custom-dropdown>
                  <button type="button" @click="state.profileForm.cityCode && state.toggleDropdown('profile_region_district')" :class="['m-dropdown-trigger h-auto py-3 px-4 rounded-xl bg-white border border-slate-200 shadow-sm', state.openDropdownKey === 'profile_region_district' ? 'is-open ring-2 ring-emerald-500 border-emerald-500' : '', !state.profileForm.cityCode ? 'opacity-50 bg-slate-50' : '']" :disabled="!state.profileForm.cityCode">
                    <div class="min-w-0 text-left">
                      <div class="text-[14px] font-black text-slate-800">{{ state.profileForm.districtName || '选择区/县' }}</div>
                    </div>
                    <svg class="w-4 h-4 text-slate-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
                  </button>
                  <div v-if="state.openDropdownKey === 'profile_region_district'" class="m-dropdown-menu hide-scrollbar rounded-xl shadow-lg border border-slate-100 mt-1 max-h-48">
                    <button v-for="item in state.districtOptions" :key="`district-${item.code}`" type="button" class="m-dropdown-option py-2.5" :class="{ 'bg-emerald-50/50': state.profileForm.districtCode === item.code }" @click="state.selectDistrictValue(item.code)">
                      <div class="text-[14px] font-black text-slate-800">{{ item.name }}</div>
                      <svg v-if="state.profileForm.districtCode === item.code" class="w-4 h-4 text-emerald-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"></path></svg>
                    </button>
                  </div>
                </div>
              </div>

              <!-- Auto Location Button -->
              <button type="button" @click="state.requestCurrentRegion" :disabled="state.locationResolving" class="mt-2 w-full flex items-center justify-center gap-2 text-[14px] font-black text-emerald-700 bg-emerald-50/80 px-4 py-3.5 rounded-xl border border-emerald-100 active:bg-emerald-100 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                <svg v-if="!state.locationResolving" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>
                <svg v-else class="w-5 h-5 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                {{ state.locationResolving ? '定位中...' : '自动定位当前位置' }}
              </button>
            </div>

            <!-- Action Buttons -->
            <div class="flex items-center gap-3 pt-4 border-t border-slate-200/50 mt-2">
              <button type="button" @click="state.updateUserProfile" :disabled="state.profileSaving" class="flex-1 text-[15px] font-black text-white bg-emerald-500 shadow-lg shadow-emerald-500/20 px-4 py-3.5 rounded-xl disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98] transition-all">
                {{ state.profileSaving ? '保存中...' : '保存修改' }}
              </button>
              <button type="button" @click="state.cancelAgeEditor" class="text-[15px] font-bold text-slate-600 bg-white border border-slate-200 px-6 py-3.5 rounded-xl active:bg-slate-50 transition-colors">
                取消
              </button>
            </div>
          </div>
        </div>

        <!-- Account Actions -->
        <div class="pt-6 border-t border-slate-100">
          <button type="button" @click="state.deleteAccount" class="w-full flex justify-center items-center gap-2 px-4 py-3.5 rounded-xl bg-rose-50/50 text-rose-500 border border-rose-100/50 active:bg-rose-100/50 transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
            <span class="text-[14px] font-bold">永久注销账户</span>
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
