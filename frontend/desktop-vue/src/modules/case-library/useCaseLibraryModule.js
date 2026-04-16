import { ref, watch } from 'vue';

export function useCaseLibraryModule(deps) {
  const caseLibrary = ref([]);
  const caseLibraryPagination = ref({
    total: 0,
    page: 1,
    page_size: 20,
    total_pages: 0,
    has_next: false,
    has_prev: false
  });
  const pendingReviews = ref([]);
  const startingCaseCollection = ref(false);
  const selectedReview = ref(null);
  const showReviewDetailModal = ref(false);
  const scamTypeOptions = ref([]);
  const targetGroupOptions = ref([]);
  const selectedCase = ref(null);
  const showCaseModal = ref(false);
  const submittingCase = ref(false);
  const duplicateCaseConflict = ref(null);

  const caseForm = ref({
    title: '',
    target_group: '',
    risk_level: '',
    scam_type: '',
    case_description: '',
    typical_scripts_raw: '',
    keywords_raw: '',
    violated_law: '',
    suggestion: ''
  });

  const caseCollectionForm = ref({
    query: '',
    case_count: 5
  });

  const fetchCaseLibrary = async (page = caseLibraryPagination.value.page, pageSize = caseLibraryPagination.value.page_size) => {
    if (!deps.isAuthenticated.value || deps.user.value.role !== 'admin') return;
    const normalizedPage = Number.isInteger(page) && page > 0 ? page : 1;
    const normalizedPageSize = Number.isInteger(pageSize) && pageSize > 0 ? pageSize : caseLibraryPagination.value.page_size;
    const res = await deps.request(`/scam/case-library/cases?page=${normalizedPage}&page_size=${normalizedPageSize}`);
    if (res && res.cases) {
      deps.replaceListIfChanged(caseLibrary, res.cases);
      caseLibraryPagination.value = {
        total: Number(res.total) || 0,
        page: Number(res.page) || normalizedPage,
        page_size: Number(res.page_size) || normalizedPageSize,
        total_pages: Number(res.total_pages) || 0,
        has_next: Boolean(res.has_next),
        has_prev: Boolean(res.has_prev)
      };
    }
  };

  const goToCaseLibraryPage = async (page) => {
    const targetPage = Number(page);
    if (!Number.isInteger(targetPage) || targetPage < 1) return;
    if (caseLibraryPagination.value.total_pages > 0 && targetPage > caseLibraryPagination.value.total_pages) return;
    await fetchCaseLibrary(targetPage, caseLibraryPagination.value.page_size);
  };

  const changeCaseLibraryPageSize = async (pageSize) => {
    const nextPageSize = Number(pageSize);
    if (!Number.isInteger(nextPageSize) || nextPageSize < 1) return;
    await fetchCaseLibrary(1, nextPageSize);
  };

  const fetchPendingReviews = async () => {
    if (!deps.isAuthenticated.value || deps.user.value.role !== 'admin') return;
    const res = await deps.request('/scam/review/cases');
    if (res && res.cases) {
      deps.replaceListIfChanged(pendingReviews, res.cases);
    }
  };

  const viewReviewDetail = async (recordId) => {
    const res = await deps.request(`/scam/review/cases/${recordId}`);
    if (res && res.case) {
      selectedReview.value = res.case;
      showReviewDetailModal.value = true;
    }
  };

  const approveReview = async (recordId) => {
    if (!confirm('确认通过该案件审核并入库知识库？')) return;
    const res = await deps.request(`/scam/review/cases/${recordId}/approve`, 'POST');
    if (res && res.case_id) {
      showReviewDetailModal.value = false;
      selectedReview.value = null;
      fetchPendingReviews();
    }
  };

  const rejectReview = async (recordId) => {
    const trimmedRecordId = String(recordId || '').trim();
    if (!trimmedRecordId) {
      deps.showToast('recordId 不能为空', 'error');
      return;
    }
    if (!confirm('确认拒绝该案件并从待审核库中删除？')) return;
    const res = await deps.request(`/scam/review/cases/${trimmedRecordId}/reject`, 'POST');
    if (res && (res.record_id || res.message)) {
      showReviewDetailModal.value = false;
      selectedReview.value = null;
      deps.showToast(res.message || '审核拒绝成功');
      await fetchPendingReviews();
      return;
    }
    deps.showToast('审核拒绝失败', 'error');
  };

  const submitCaseCollection = async () => {
    const query = String(caseCollectionForm.value.query || '').trim();
    const caseCount = Number(caseCollectionForm.value.case_count);
    if (!query) {
      deps.showToast('采集主题不能为空', 'error');
      return;
    }
    if (!Number.isInteger(caseCount) || caseCount < 1 || caseCount > 20) {
      deps.showToast('案件数量取值范围应为 1-20', 'error');
      return;
    }

    startingCaseCollection.value = true;
    try {
      const res = await deps.request('/scam/case-collection/search', 'POST', {
        query,
        case_count: caseCount
      });
      deps.showToast((res && res.message) || '案件采集任务已在后台启动');
      caseCollectionForm.value.query = '';
      setTimeout(() => fetchPendingReviews(), 1200);
    } catch (e) {
      deps.showToast(`启动失败: ${e.message}`, 'error');
    } finally {
      startingCaseCollection.value = false;
    }
  };

  const fetchCaseOptionLists = async () => {
    if (!deps.isAuthenticated.value || deps.user.value.role !== 'admin') return;
    const [scamTypeRes, targetGroupRes] = await Promise.all([
      deps.request('/scam/case-library/options/scam-types'),
      deps.request('/scam/case-library/options/target-groups')
    ]);
    if (scamTypeRes && Array.isArray(scamTypeRes.options)) {
      deps.replaceListIfChanged(scamTypeOptions, scamTypeRes.options);
    }
    if (targetGroupRes && Array.isArray(targetGroupRes.options)) {
      deps.replaceListIfChanged(targetGroupOptions, targetGroupRes.options);
    }
  };

  const openCaseModal = () => {
    duplicateCaseConflict.value = null;
    caseForm.value = {
      title: '',
      target_group: '',
      risk_level: '',
      scam_type: '',
      case_description: '',
      typical_scripts_raw: '',
      keywords_raw: '',
      violated_law: '',
      suggestion: ''
    };
    showCaseModal.value = true;
  };

  const viewCaseDetail = async (caseId) => {
    const res = await deps.request(`/scam/case-library/cases/${caseId}`);
    if (res && res.case) {
      selectedCase.value = res.case;
    }
  };

  const viewDuplicateCaseConflict = async () => {
    const caseId = String(duplicateCaseConflict.value?.case_id || '').trim();
    if (!caseId) return;
    duplicateCaseConflict.value = null;
    showCaseModal.value = false;
    await viewCaseDetail(caseId);
  };

  const minCaseDescriptionRunes = 12;
  const maxCaseDescriptionRunes = 400;
  const randomLikeAlnumChunkLimit = 16;

  const validateCaseDescriptionQualityClient = (description) => {
    const normalized = String(description || '').trim().replace(/\s+/g, ' ');
    const runes = Array.from(normalized);

    if (!normalized) return { ok: false, message: '案件描述不能为空' };
    if (runes.length < minCaseDescriptionRunes) return { ok: false, message: `案件描述过短，至少 ${minCaseDescriptionRunes} 个字符` };
    if (runes.length > maxCaseDescriptionRunes) return { ok: false, message: `案件描述过长，最多 ${maxCaseDescriptionRunes} 个字符` };

    const uniqueChars = new Set(runes);
    if (uniqueChars.size <= 2) return { ok: false, message: '案件描述疑似无效，请填写有语义的内容' };

    const hasHan = /[\u4e00-\u9fff]/.test(normalized);
    const hasSeparator = /[^A-Za-z0-9]/.test(normalized);
    let maxAlnumChunk = 0;
    let currentAlnumChunk = 0;
    let alnumCount = 0;
    let digitCount = 0;

    for (const ch of runes) {
      if (/[A-Za-z0-9]/.test(ch)) {
        currentAlnumChunk += 1;
        alnumCount += 1;
        if (/[0-9]/.test(ch)) digitCount += 1;
      } else {
        currentAlnumChunk = 0;
      }
      if (currentAlnumChunk > maxAlnumChunk) {
        maxAlnumChunk = currentAlnumChunk;
      }
    }

    if (!hasHan && !hasSeparator && maxAlnumChunk >= randomLikeAlnumChunkLimit) {
      return { ok: false, message: '案件描述疑似随机字符串，请补充有效描述' };
    }
    if (!hasHan && alnumCount >= minCaseDescriptionRunes) {
      const digitRatio = digitCount / alnumCount;
      if (maxAlnumChunk >= minCaseDescriptionRunes && digitRatio > 0.35) {
        return { ok: false, message: '案件描述疑似随机字符串，请补充有效描述' };
      }
    }
    return { ok: true, message: '' };
  };

  const submitCase = async () => {
    duplicateCaseConflict.value = null;
    if (!String(caseForm.value.title || '').trim()) return deps.showToast('案件标题不能为空', 'error');
    if (!String(caseForm.value.target_group || '').trim()) return deps.showToast('目标人群不能为空', 'error');
    if (!String(caseForm.value.risk_level || '').trim()) return deps.showToast('风险等级不能为空', 'error');
    if (!String(caseForm.value.scam_type || '').trim()) return deps.showToast('诈骗类型不能为空', 'error');

    const descriptionValidation = validateCaseDescriptionQualityClient(caseForm.value.case_description);
    if (!descriptionValidation.ok) {
      deps.showToast(descriptionValidation.message, 'error');
      return;
    }

    submittingCase.value = true;
    try {
      const payload = {
        title: String(caseForm.value.title || '').trim(),
        target_group: String(caseForm.value.target_group || '').trim(),
        risk_level: String(caseForm.value.risk_level || '').trim(),
        scam_type: String(caseForm.value.scam_type || '').trim(),
        case_description: String(caseForm.value.case_description || '').trim(),
        typical_scripts: String(caseForm.value.typical_scripts_raw || '').split('\n').filter(s => s.trim()),
        keywords: String(caseForm.value.keywords_raw || '').split(/[,，]/).map(s => s.trim()).filter(Boolean),
        violated_law: String(caseForm.value.violated_law || '').trim(),
        suggestion: String(caseForm.value.suggestion || '').trim()
      };

      const res = await deps.request('/scam/case-library/cases', 'POST', payload, { silent: true, throwOnError: true });
      if (res) {
        duplicateCaseConflict.value = null;
        deps.showToast('案件录入成功');
        showCaseModal.value = false;
        fetchCaseLibrary(caseLibraryPagination.value.page, caseLibraryPagination.value.page_size);
        fetchCaseOptionLists();
        await deps.refreshAdminStats(true);
      }
    } catch (e) {
      if (e?.status === 409 && e?.data?.duplicate_case) {
        duplicateCaseConflict.value = e.data.duplicate_case;
        deps.showToast(e.data.message || '检测到高度相似案件，请先核对已有记录', 'warning');
        return;
      }
      deps.showToast(`录入失败: ${e.message}`, 'error');
    } finally {
      submittingCase.value = false;
    }
  };

  const deleteCase = async (item) => {
    if (!item || !item.case_id) return;
    if (!confirm(`确定删除案件 ${item.title} 吗？此操作不可恢复。`)) return;
    try {
      const res = await deps.request(`/scam/case-library/cases/${item.case_id}`, 'DELETE');
      if (res) {
        deps.showToast(res.message || '案件已删除');
        const fallbackPage = caseLibrary.value.length === 1 && caseLibraryPagination.value.page > 1
          ? caseLibraryPagination.value.page - 1
          : caseLibraryPagination.value.page;
        fetchCaseLibrary(fallbackPage, caseLibraryPagination.value.page_size);
        await deps.refreshAdminStats(true);
        if (selectedCase.value && selectedCase.value.case_id === item.case_id) {
          selectedCase.value = null;
        }
      }
    } catch (e) {
      deps.showToast(`删除失败: ${e.message}`, 'error');
    }
  };

  watch(showCaseModal, (visible) => {
    if (!visible) {
      duplicateCaseConflict.value = null;
    }
  });

  return {
    caseLibrary,
    caseLibraryPagination,
    pendingReviews,
    startingCaseCollection,
    selectedReview,
    showReviewDetailModal,
    scamTypeOptions,
    targetGroupOptions,
    selectedCase,
    showCaseModal,
    submittingCase,
    duplicateCaseConflict,
    caseForm,
    caseCollectionForm,
    fetchCaseLibrary,
    goToCaseLibraryPage,
    changeCaseLibraryPageSize,
    fetchPendingReviews,
    viewReviewDetail,
    approveReview,
    rejectReview,
    submitCaseCollection,
    fetchCaseOptionLists,
    openCaseModal,
    viewDuplicateCaseConflict,
    submitCase,
    viewCaseDetail,
    deleteCase
  };
}
