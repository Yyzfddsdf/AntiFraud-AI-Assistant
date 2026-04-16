<template>
  <DashboardSectionShell
    eyebrow="Simulation"
    title="反诈模拟"
    :heading="simulationSidebarSection === 'assistant' ? '' : examViewMode ? '答题界面' : '模拟题目'"
    description=""
  >
        <template #nav>
          <button @click="setSimulationSidebarSection('quiz')" :class="['w-full flex items-center gap-3 rounded-sm px-3 py-2 text-left text-sm font-bold transition-all border', simulationSidebarSection === 'quiz' ? 'bg-white text-brand-700 border-slate-200' : 'bg-white text-slate-700 border-transparent hover:border-slate-200 hover:text-brand-700 hover:bg-slate-50']">
            <span :class="simulationSidebarSection === 'quiz' ? 'bg-brand-600' : 'bg-slate-300'" class="w-2 h-2 rounded-full shrink-0"></span>
            <span class="min-w-0 truncate">题目</span>
          </button>
          <button @click="openAssistantPanel" :class="['w-full flex items-center gap-3 rounded-sm px-3 py-2 text-left text-sm font-bold transition-all border', simulationSidebarSection === 'assistant' ? 'bg-white text-brand-700 border-slate-200' : 'bg-white text-slate-700 border-transparent hover:border-slate-200 hover:text-brand-700 hover:bg-slate-50']">
            <span :class="simulationSidebarSection === 'assistant' ? 'bg-brand-600' : 'bg-slate-300'" class="w-2 h-2 rounded-full shrink-0"></span>
            <span class="min-w-0 truncate">助手</span>
          </button>
        </template>

        <template #badges>
          <span class="px-2 py-0.5 rounded-sm bg-brand-50 border border-brand-100 text-brand-700 text-[11px] font-bold">{{ simulationSidebarSection === 'assistant' ? '助手' : '题目' }}</span>
          <span class="px-2 py-0.5 rounded-sm bg-slate-100 border border-slate-200 text-slate-700 text-[11px] font-bold">{{ statusLabel }}</span>
        </template>

    <div v-if="simulationSidebarSection === 'quiz' && !examViewMode" class="space-y-4">
      <div class="rounded-sm border border-slate-200 bg-white p-6 shadow-sm">
        <div class="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
          <div>
            <div class="text-[11px] font-black uppercase tracking-[0.32em] text-brand-600">Sentinel AI</div>
            <h1 class="mt-2 text-3xl font-black tracking-tight text-slate-950">反诈模拟答题</h1>
          </div>

          <div class="grid grid-cols-2 gap-3 xl:w-[280px]">
            <div class="rounded-sm border border-slate-200 bg-slate-50 px-4 py-3">
              <div class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">状态</div>
              <div class="mt-1 text-sm font-black text-slate-900">{{ statusLabel }}</div>
            </div>
            <div class="rounded-sm border border-slate-200 bg-slate-50 px-4 py-3">
              <div class="text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">报告数</div>
              <div class="mt-1 text-sm font-black text-slate-900">{{ completedSessionCount }}</div>
            </div>
          </div>
        </div>
      </div>

      <div class="grid gap-4 2xl:grid-cols-[minmax(0,1.2fr)_420px]">
        <div class="space-y-4">
          <div class="rounded-sm border border-slate-200 bg-white p-5 shadow-sm">
            <div class="flex items-center justify-between gap-3">
              <div>
                <div class="text-[11px] font-black uppercase tracking-[0.22em] text-slate-400">Generator</div>
                <h3 class="mt-1 text-xl font-black text-slate-950">创建一套新的模拟试卷</h3>
              </div>
              <span class="rounded-sm border border-slate-200 bg-slate-50 px-3 py-1 text-[11px] font-bold text-slate-500">固定 10 步</span>
            </div>

            <div class="mt-4 grid gap-4 md:grid-cols-2">
              <label class="block">
                <div class="mb-2 text-xs font-bold uppercase tracking-[0.18em] text-slate-500">案件类型</div>
                <input
                  v-model="simulationForm.caseType"
                  type="text"
                  class="w-full rounded-sm border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700 focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
                  placeholder="例如：冒充客服"
                />
              </label>
              <label class="block">
                <div class="mb-2 text-xs font-bold uppercase tracking-[0.18em] text-slate-500">目标人群</div>
                <input
                  v-model="simulationForm.targetPersona"
                  type="text"
                  class="w-full rounded-sm border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700 focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
                  placeholder="例如：普通居民"
                />
              </label>
            </div>

            <div class="mt-4 flex flex-wrap gap-3">
              <button @click="simulationForm.difficulty = 'easy'" :class="difficultyButtonClass('easy')">简单</button>
              <button @click="simulationForm.difficulty = 'medium'" :class="difficultyButtonClass('medium')">中等</button>
              <button @click="simulationForm.difficulty = 'hard'" :class="difficultyButtonClass('hard')">困难</button>
            </div>

            <div class="mt-5 flex flex-wrap gap-3">
              <button @click="generateSimulationPack" :disabled="simulationGenerating" class="min-w-36 rounded-sm bg-brand-600 px-5 py-3 text-sm font-black text-white shadow-sm transition-colors hover:bg-brand-700 disabled:opacity-60">{{ simulationGenerating ? '生成中...' : '生成题包' }}</button>
              <button @click="resetSimulation" class="rounded-sm border border-slate-200 bg-white px-5 py-3 text-sm font-black text-slate-600 transition-colors hover:bg-slate-50">重置状态</button>
            </div>
          </div>

          <div class="rounded-sm border border-slate-200 bg-white p-5 shadow-sm">
            <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
              <div class="min-w-0">
                <div class="text-[11px] font-black uppercase tracking-[0.22em] text-slate-400">Current Pack</div>
                <h3 class="mt-1 text-xl font-black text-slate-950">{{ activePackTitle }}</h3>
                <p class="mt-2 text-sm leading-6 text-slate-600">{{ activePackIntro }}</p>
              </div>
              <div class="grid grid-cols-2 gap-3 lg:w-[260px]">
                <div class="rounded-sm border border-slate-200 bg-slate-50 px-3 py-3">
                  <div class="text-[11px] font-bold uppercase tracking-[0.14em] text-slate-400">进度</div>
                  <div class="mt-1 text-sm font-black text-slate-900">{{ progressMeta.current }}/{{ progressMeta.total || defaultStepCount }}</div>
                </div>
                <div class="rounded-sm border border-slate-200 bg-slate-50 px-3 py-3">
                  <div class="text-[11px] font-bold uppercase tracking-[0.14em] text-slate-400">题型</div>
                  <div class="mt-1 text-sm font-black text-slate-900">{{ simulationCurrentStep?.step_type || '待开始' }}</div>
                </div>
              </div>
            </div>

            <div class="mt-4 flex flex-wrap gap-3">
              <button v-if="simulationStatus === 'pack_ready'" @click="startExamFromCurrentPack" :disabled="simulationSubmitting" class="rounded-sm bg-brand-600 px-5 py-3 text-sm font-black text-white transition-colors hover:bg-brand-700 disabled:opacity-60">进入答题</button>
              <button v-if="simulationStatus === 'in_progress'" @click="openExamView" class="rounded-sm bg-brand-600 px-5 py-3 text-sm font-black text-white transition-colors hover:bg-brand-700">继续答题</button>
              <button v-if="simulationStatus === 'completed' && simulationResult" @click="openExamView" class="rounded-sm border border-brand-200 bg-brand-50 px-5 py-3 text-sm font-black text-brand-700 transition-colors hover:bg-brand-100">查看结果</button>
              <span class="inline-flex items-center rounded-sm border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-bold text-slate-500">{{ currentStepLabel }}</span>
            </div>
          </div>
        </div>

        <div class="space-y-4">
          <div class="rounded-sm border border-slate-200 bg-white p-5 shadow-sm">
            <div class="flex items-center justify-between gap-3">
              <div>
                <div class="text-[11px] font-black uppercase tracking-[0.22em] text-slate-400">Reports</div>
                <h3 class="mt-1 text-lg font-black text-slate-950">报告历史</h3>
              </div>
              <button @click="fetchSimulationSessions" class="rounded-sm border border-slate-200 bg-slate-50 px-3 py-2 text-xs font-black text-slate-600 transition hover:bg-slate-100">刷新</button>
            </div>
            <div class="mt-4 space-y-3 max-h-[22rem] overflow-y-auto pr-1">
              <div v-for="item in simulationSessionList" :key="`session-list-${item.pack_id}`" class="rounded-sm border border-slate-200 bg-slate-50 p-4">
                <div class="flex items-start gap-3">
                  <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-sm text-xs font-black" :class="scoreClass(item.score)">
                    {{ item.score }}
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="text-sm font-black leading-5 text-slate-900">{{ item.title || '未命名题包' }}</div>
                    <div class="mt-1 text-xs leading-5 text-slate-500">{{ item.level || '未评分' }} · {{ item.status === 'completed' ? '已完成' : '未完成' }}</div>
                  </div>
                </div>
                <div class="mt-3 flex items-center justify-end gap-2">
                  <button
                    v-if="item.status !== 'completed'"
                    @click="startExamSession(item.pack_id)"
                    :disabled="simulationSubmitting"
                    class="rounded-sm border border-brand-200 bg-brand-50 px-3 py-2 text-xs font-black text-brand-700 transition hover:bg-brand-100 disabled:opacity-50"
                  >
                    继续
                  </button>
                  <button @click="deleteSimulationSession(item.pack_id)" class="rounded-sm border border-rose-200 bg-rose-50 px-3 py-2 text-xs font-black text-rose-600 transition hover:bg-rose-100">
                    删除
                  </button>
                </div>
              </div>
              <div v-if="!simulationSessionList.length" class="rounded-sm border border-dashed border-slate-200 px-4 py-8 text-center text-sm text-slate-400">暂无报告</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="simulationSidebarSection === 'quiz'" class="rounded-sm border border-slate-200 bg-white shadow-sm overflow-hidden">
      <div class="border-b border-slate-200 bg-slate-50 px-6 py-4">
        <div class="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
          <div class="min-w-0">
            <button @click="closeExamView" class="inline-flex items-center gap-2 rounded-sm border border-slate-200 bg-white px-3 py-2 text-xs font-black uppercase tracking-[0.14em] text-slate-600 transition hover:bg-slate-50">
              返回概览
            </button>
            <h2 class="mt-3 text-3xl font-black tracking-tight text-slate-950">{{ activePackTitle }}</h2>
          </div>

          <div class="grid grid-cols-3 gap-3 xl:w-[460px]">
            <div class="rounded-sm border border-slate-200 bg-white px-4 py-3">
              <div class="text-[11px] font-bold uppercase tracking-[0.14em] text-slate-400">状态</div>
              <div class="mt-1 text-sm font-black text-slate-900">{{ statusLabel }}</div>
            </div>
            <div class="rounded-sm border border-slate-200 bg-white px-4 py-3">
              <div class="text-[11px] font-bold uppercase tracking-[0.14em] text-slate-400">进度</div>
              <div class="mt-1 text-sm font-black text-slate-900">{{ progressMeta.current }}/{{ progressMeta.total || defaultStepCount }}</div>
            </div>
            <div class="rounded-sm border border-slate-200 bg-white px-4 py-3">
              <div class="text-[11px] font-bold uppercase tracking-[0.14em] text-slate-400">当前分数</div>
              <div class="mt-1 text-sm font-black" :class="scoreToneClass">{{ simulationCurrentScore }}</div>
            </div>
          </div>
        </div>

        <div class="mt-4">
          <div class="flex items-center justify-between text-[11px] font-bold uppercase tracking-[0.16em] text-slate-400">
            <span>答题进度</span>
            <span>{{ progressMeta.percent }}%</span>
          </div>
          <div class="mt-2 h-2 overflow-hidden rounded-full bg-slate-200">
            <div class="h-full rounded-full bg-brand-500 transition-all duration-500 ease-out" :style="{ width: `${progressMeta.percent}%` }"></div>
          </div>
        </div>
      </div>

      <div class="px-6 py-6">
        <div v-if="simulationCurrentStep && simulationStatus === 'in_progress'" class="w-full space-y-5">
          <div class="inline-flex items-center rounded-sm border border-slate-200 bg-slate-50 px-3 py-1 text-[11px] font-black uppercase tracking-[0.18em] text-slate-500">{{ simulationCurrentStep.step_type }}</div>
          <div class="rounded-sm border border-slate-200 bg-slate-50 p-5 text-sm leading-7 text-slate-600">{{ simulationCurrentStep.narrative }}</div>
          <div class="text-3xl font-black leading-10 tracking-tight text-slate-950">{{ simulationCurrentStep.question }}</div>
          <div class="grid gap-3 xl:grid-cols-2">
            <button
              v-for="option in simulationCurrentStep.options"
              :key="`exam-option-${simulationCurrentStep.step_id}-${option.key}`"
              @click="submitSimulationAnswer(option.key)"
              :disabled="simulationSubmitting || simulationStatus !== 'in_progress'"
              class="group flex items-start gap-4 rounded-sm border border-slate-200 bg-white px-5 py-4 text-left shadow-sm transition-all hover:border-brand-300 hover:bg-brand-50 disabled:opacity-60"
            >
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-sm bg-slate-100 text-sm font-black text-slate-700 transition-colors group-hover:bg-brand-100 group-hover:text-brand-700">{{ option.key }}</div>
              <div class="min-w-0">
                <div class="text-base font-bold leading-7 text-slate-900">{{ option.text }}</div>
              </div>
            </button>
          </div>
        </div>

        <div v-else-if="simulationStatus === 'completed' && simulationResult" class="grid w-full gap-5 2xl:grid-cols-[minmax(0,1.15fr)_minmax(320px,0.85fr)]">
          <div class="rounded-sm border border-brand-200 bg-brand-50 p-6">
            <h4 class="text-2xl font-black text-brand-900">答题完成 · {{ simulationResult.level }}</h4>
            <p class="mt-2 text-sm text-brand-800">总分：{{ simulationResult.total_score }}</p>
            <div v-if="simulationResult.advice && simulationResult.advice.length" class="mt-4 space-y-3 text-sm leading-6 text-slate-700">
              <div v-for="(advice, idx) in simulationResult.advice" :key="`result-advice-${idx}`" class="flex gap-3">
                <span class="font-black text-brand-700">{{ idx + 1 }}.</span>
                <span>{{ advice }}</span>
              </div>
            </div>
          </div>

          <div class="grid gap-4">
            <div class="rounded-sm border border-slate-200 bg-white p-5">
              <div class="text-[11px] font-black uppercase tracking-[0.16em] text-slate-400">Strengths</div>
              <div v-if="simulationResult.strengths && simulationResult.strengths.length" class="mt-3 space-y-2 text-sm leading-6 text-slate-600">
                <div v-for="(item, idx) in simulationResult.strengths" :key="`sim-strength-${idx}`" class="rounded-sm bg-emerald-50 px-3 py-2 text-emerald-800">{{ item }}</div>
              </div>
              <div v-else class="mt-3 text-sm text-slate-400">暂无优势项总结</div>
            </div>

            <div class="rounded-sm border border-slate-200 bg-white p-5">
              <div class="text-[11px] font-black uppercase tracking-[0.16em] text-slate-400">Weaknesses</div>
              <div v-if="simulationResult.weaknesses && simulationResult.weaknesses.length" class="mt-3 space-y-2 text-sm leading-6 text-slate-600">
                <div v-for="(item, idx) in simulationResult.weaknesses" :key="`sim-weakness-${idx}`" class="rounded-sm bg-amber-50 px-3 py-2 text-amber-800">{{ item }}</div>
              </div>
              <div v-else class="mt-3 text-sm text-slate-400">暂无薄弱项总结</div>
            </div>
          </div>
        </div>

        <div v-else class="w-full rounded-sm border border-dashed border-slate-200 bg-slate-50 px-8 py-16 text-center">
          <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-sm border border-slate-200 bg-white">
            <span class="text-2xl font-black text-slate-400">Q</span>
          </div>
          <h4 class="mt-4 text-xl font-black text-slate-900">答题界面待启动</h4>
          <p class="mt-2 text-sm leading-6 text-slate-500">请返回概览页，从题目列表或当前题包进入答题。</p>
        </div>
      </div>
    </div>

    <DesktopChatView v-else :app="app" embedded @back="setSimulationSidebarSection('quiz')" />
  </DashboardSectionShell>
</template>

<script>
import { computed, ref, unref } from 'vue';
import DesktopChatView from './DesktopChatView.vue';
import DashboardSectionShell from '../DashboardSectionShell.vue';

const STATUS_LABELS = {
  idle: '待生成',
  pack_ready: '可开始',
  in_progress: '答题中',
  completed: '已完成'
};

const DIFFICULTY_LABELS = {
  easy: '简单',
  medium: '中等',
  hard: '困难'
};

export default {
  name: 'DesktopSimulationQuizView',
  components: {
    DesktopChatView,
    DashboardSectionShell
  },
  props: {
    app: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    const simulationPackValue = computed(() => unref(props.app.simulationPack) || null);
    const simulationCurrentStepValue = computed(() => unref(props.app.simulationCurrentStep) || null);
    const simulationResultValue = computed(() => unref(props.app.simulationResult) || null);
    const simulationAnswersValue = computed(() => {
      const value = unref(props.app.simulationAnswers);
      return Array.isArray(value) ? value : [];
    });
    const simulationPackListValue = computed(() => {
      const value = unref(props.app.simulationPackList);
      return Array.isArray(value) ? value : [];
    });
    const simulationSessionListValue = computed(() => {
      const value = unref(props.app.simulationSessionList);
      return Array.isArray(value) ? value : [];
    });
    const simulationStatusValue = computed(() => String(unref(props.app.simulationStatus) || 'idle').trim() || 'idle');
    const simulationCurrentScoreValue = computed(() => Number(unref(props.app.simulationCurrentScore)) || 0);
    const defaultStepCount = 10;
    const examViewMode = ref(false);
    const simulationSidebarSection = ref('quiz');

    const totalSteps = computed(() => {
      const steps = simulationPackValue.value?.steps;
      return Array.isArray(steps) ? steps.length : 0;
    });

    const currentStepIndex = computed(() => {
      const steps = simulationPackValue.value?.steps;
      const currentStepID = simulationCurrentStepValue.value?.step_id;
      if (!Array.isArray(steps) || !currentStepID) return -1;
      return steps.findIndex((item) => item?.step_id === currentStepID);
    });

    const progressMeta = computed(() => {
      const total = totalSteps.value;
      if (!total) return { current: 0, total: 0, percent: 0 };

      const answered = simulationAnswersValue.value.length;
      let current = answered;
      if (simulationStatusValue.value === 'completed') {
        current = total;
      } else if (simulationStatusValue.value === 'in_progress' && simulationCurrentStepValue.value) {
        current = Math.min(answered + 1, total);
      } else if (simulationStatusValue.value === 'pack_ready') {
        current = Math.min(Math.max(answered, 1), total);
      }

      return {
        current,
        total,
        percent: Math.round((current / total) * 100)
      };
    });

    const statusLabel = computed(() => STATUS_LABELS[simulationStatusValue.value] || STATUS_LABELS.idle);
    const difficultyValue = computed(() => {
      const fromPack = String(simulationPackValue.value?.difficulty || '').trim();
      const fromForm = String(unref(props.app.simulationForm)?.difficulty || '').trim();
      return fromPack || fromForm || 'medium';
    });
    const difficultyLabel = computed(() => DIFFICULTY_LABELS[difficultyValue.value] || difficultyValue.value || '中等');
    const activePackTitle = computed(() => simulationPackValue.value?.title || '尚未载入题包');
    const activePackIntro = computed(() => simulationPackValue.value?.intro || '先生成题包，再从题目列表进入答题。');
    const activeKnowledgePoint = computed(() => simulationCurrentStepValue.value?.knowledge_point || '待进入题目后显示');
    const currentStepTimeLimit = computed(() => {
      const value = Number(simulationCurrentStepValue.value?.time_limit_sec || 0);
      return value > 0 ? `${value} 秒` : '未设定';
    });
    const currentStepLabel = computed(() => {
      if (simulationStatusValue.value === 'completed') return `已完成 ${totalSteps.value || defaultStepCount} / ${totalSteps.value || defaultStepCount}`;
      if (currentStepIndex.value >= 0 && totalSteps.value) return `第 ${currentStepIndex.value + 1} / ${totalSteps.value} 题`;
      if (simulationStatusValue.value === 'pack_ready') return '等待开始';
      return '尚未开始';
    });

    const completedSessionCount = computed(() => simulationSessionListValue.value.filter((item) => item?.status === 'completed').length);
    const scoreToneClass = computed(() => {
      if (simulationCurrentScoreValue.value >= 80) return 'text-emerald-600';
      if (simulationCurrentScoreValue.value >= 60) return 'text-amber-600';
      return 'text-rose-600';
    });
    const answeredAnswers = computed(() => simulationAnswersValue.value.slice().reverse());

    const difficultyText = (value) => DIFFICULTY_LABELS[String(value || '').trim()] || String(value || '未设定').trim();
    const difficultyButtonClass = (value) => (String(unref(props.app.simulationForm)?.difficulty || '').trim() === value
      ? 'rounded-sm border border-brand-200 bg-brand-50 px-4 py-2.5 text-xs font-black text-brand-700 shadow-sm'
      : 'rounded-sm border border-slate-200 bg-white px-4 py-2.5 text-xs font-black text-slate-500 transition-colors hover:bg-slate-50');
    const scoreClass = (value) => {
      const score = Number(value) || 0;
      if (score >= 80) return 'bg-emerald-100 text-emerald-700';
      if (score >= 60) return 'bg-amber-100 text-amber-700';
      return 'bg-rose-100 text-rose-700';
    };

    const openExamView = () => {
      examViewMode.value = true;
    };
    const closeExamView = () => {
      examViewMode.value = false;
    };
    const startExamSession = async (packID) => {
      await props.app.startSimulationSession(packID);
      simulationSidebarSection.value = 'quiz';
      examViewMode.value = true;
    };
    const startExamFromCurrentPack = async () => {
      await props.app.startSimulationSession();
      simulationSidebarSection.value = 'quiz';
      examViewMode.value = true;
    };
    const setSimulationSidebarSection = (nextSection) => {
      simulationSidebarSection.value = nextSection === 'assistant' ? 'assistant' : 'quiz';
    };
    const openAssistantPanel = async () => {
      simulationSidebarSection.value = 'assistant';
      await props.app.fetchChatHistory();
    };

    return {
      ...props.app,
      simulationPack: simulationPackValue,
      simulationCurrentStep: simulationCurrentStepValue,
      simulationResult: simulationResultValue,
      simulationAnswers: simulationAnswersValue,
      simulationPackList: simulationPackListValue,
      simulationSessionList: simulationSessionListValue,
      simulationStatus: simulationStatusValue,
      simulationCurrentScore: simulationCurrentScoreValue,
      defaultStepCount,
      examViewMode,
      simulationSidebarSection,
      progressMeta,
      statusLabel,
      difficultyLabel,
      activePackTitle,
      activePackIntro,
      activeKnowledgePoint,
      currentStepTimeLimit,
      currentStepLabel,
      completedSessionCount,
      scoreToneClass,
      answeredAnswers,
      difficultyText,
      difficultyButtonClass,
      scoreClass,
      openExamView,
      closeExamView,
      setSimulationSidebarSection,
      openAssistantPanel,
      startExamSession,
      startExamFromCurrentPack
    };
  }
};
</script>
