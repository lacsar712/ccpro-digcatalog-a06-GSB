<template>
  <div>
    <div class="toolbar">
      <div>
        <h2 class="page-title">器物度量单</h2>
        <p class="page-sub">一件文物可多次度量，按测量时间倒序保留完整历史</p>
      </div>
    </div>

    <div class="card">
      <div class="filters">
        <label class="search-label">
          登记号检索
          <input
            v-model="keyword"
            placeholder="如 EL-2024-0001"
            @keyup.enter="search"
          />
        </label>
        <button class="btn" @click="search">检索</button>
      </div>
      <p v-if="searchError" class="error">{{ searchError }}</p>

      <div v-if="searchResults.length && !current" class="result-list">
        <table class="table">
          <thead>
            <tr>
              <th>登记号</th>
              <th>器物类型</th>
              <th>材质</th>
              <th>完整度</th>
              <th>最近度量</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="f in searchResults" :key="f.id">
              <td>{{ f.registerNo }}</td>
              <td><span class="tag">{{ f.artifactType }}</span></td>
              <td>{{ f.materialName || f.material?.name || '-' }}</td>
              <td>{{ f.completeness || '-' }}</td>
              <td>{{ latestSummary(f) }}</td>
              <td><button class="btn secondary small" @click="selectFind(f)">查看度量单</button></td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-else-if="searched && !searchResults.length && !current" class="page-sub">
        未找到匹配登记号的文物
      </p>
    </div>

    <template v-if="current">
      <div class="card find-head">
        <button class="btn secondary small back-btn" @click="clearCurrent">← 返回检索</button>
        <div class="find-meta">
          <h3 class="find-reg">{{ current.registerNo }}</h3>
          <p class="page-sub">
            {{ current.artifactType }} · {{ current.materialName || current.material?.name || '材质未指定' }}
            · {{ current.completeness || '完整度未填' }}
          </p>
          <div v-if="current.lastMeasuredAt" class="latest-box">
            <span class="latest-title">最近度量（{{ formatDateTime(current.lastMeasuredAt) }}）：</span>
            长 {{ current.lastLengthMm }} mm × 宽 {{ current.lastWidthMm }} mm ×
            高 {{ current.lastHeightMm }} mm<template v-if="current.lastWeightG != null">
              · 重 {{ current.lastWeightG }} g</template>
          </div>
          <p v-else class="page-sub">尚无比量记录</p>
        </div>
        <button class="btn" @click="openCreate">新增度量</button>
      </div>

      <div class="card">
        <h3>历史度量（{{ sheets.length }}）</h3>
        <table class="table">
          <thead>
            <tr>
              <th>测量时间 ↓</th>
              <th>长 (mm)</th>
              <th>宽 (mm)</th>
              <th>高 (mm)</th>
              <th>重 (g)</th>
              <th>测量人</th>
              <th>卡尺备注</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in sheets" :key="s.id" :class="{ latest: isLatest(s) }">
              <td>{{ formatDateTime(s.measuredAt) }}</td>
              <td>{{ s.lengthMm }}</td>
              <td>{{ s.widthMm }}</td>
              <td>{{ s.heightMm }}</td>
              <td>{{ s.weightG == null ? '-' : s.weightG }}</td>
              <td>{{ s.operatorName }}</td>
              <td class="note-cell">{{ s.caliperNote || '-' }}</td>
            </tr>
          </tbody>
        </table>
        <p v-if="!sheets.length" class="page-sub">该文物暂无度量记录，点击「新增度量」开始登记</p>
        <p v-if="listError" class="error">{{ listError }}</p>
      </div>
    </template>

    <div v-if="showModal" class="modal-mask" @click.self="showModal = false">
      <div class="modal">
        <h3>新增度量 · {{ current.registerNo }}</h3>
        <div class="form-grid">
          <label>
            测量时间
            <input v-model="form.measuredAt" type="datetime-local" />
          </label>
          <label>
            测量人
            <input v-model="form.operatorName" placeholder="必填" />
          </label>
          <label>
            长 (mm)
            <input v-model.number="form.lengthMm" type="number" min="0" step="0.1" />
          </label>
          <label>
            宽 (mm)
            <input v-model.number="form.widthMm" type="number" min="0" step="0.1" />
          </label>
          <label>
            高 (mm)
            <input v-model.number="form.heightMm" type="number" min="0" step="0.1" />
          </label>
          <label>
            重 (g，可空)
            <input v-model="form.weightG" type="number" min="0" step="0.1" placeholder="未称重可留空" />
          </label>
          <label class="full">
            卡尺测量备注
            <textarea v-model="form.caliperNote" placeholder="测量基准、仪器、异常情况等" />
          </label>
        </div>
        <p v-if="formError" class="error">{{ formError }}</p>
        <div class="modal-actions">
          <button class="btn secondary" @click="showModal = false">取消</button>
          <button class="btn" @click="save">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api/http'

const route = useRoute()
const router = useRouter()

const keyword = ref('')
const searched = ref(false)
const searchResults = ref([])
const searchError = ref('')
const current = ref(null)
const sheets = ref([])
const listError = ref('')
const showModal = ref(false)
const formError = ref('')

const form = reactive({
  measuredAt: '',
  operatorName: '',
  lengthMm: 0,
  widthMm: 0,
  heightMm: 0,
  weightG: '',
  caliperNote: ''
})

function pad(n) {
  return String(n).padStart(2, '0')
}

function nowLocalInput() {
  const d = new Date()
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function formatDateTime(v) {
  if (!v) return '-'
  return String(v).replace('T', ' ').slice(0, 16)
}

function latestSummary(f) {
  if (!f.lastMeasuredAt) return '无'
  const w = f.lastWeightG == null ? '' : ` / ${f.lastWeightG}g`
  return `${f.lastLengthMm}×${f.lastWidthMm}×${f.lastHeightMm}mm${w}`
}

function isLatest(s) {
  return sheets.value.length > 0 && sheets.value[0].id === s.id
}

async function search() {
  searchError.value = ''
  searched.value = false
  searchResults.value = []
  if (!keyword.value.trim()) {
    searchError.value = '请输入登记号'
    return
  }
  try {
    const { data } = await api.get('/finds', { params: { registerNo: keyword.value.trim() } })
    searchResults.value = data
    searched.value = true
  } catch (e) {
    searchError.value = e.response?.data?.error || '检索失败'
  }
}

async function loadSheets() {
  listError.value = ''
  try {
    const { data } = await api.get(`/finds/${current.value.id}/measurements`)
    sheets.value = data
  } catch (e) {
    listError.value = e.response?.data?.error || '度量历史加载失败'
  }
}

async function selectFind(f) {
  current.value = f
  searchResults.value = []
  searched.value = false
  router.replace({ name: 'measurements', query: { findId: f.id } })
  await loadSheets()
}

function clearCurrent() {
  current.value = null
  sheets.value = []
  router.replace({ name: 'measurements', query: {} })
}

function openCreate() {
  Object.assign(form, {
    measuredAt: nowLocalInput(),
    operatorName: '',
    lengthMm: 0,
    widthMm: 0,
    heightMm: 0,
    weightG: '',
    caliperNote: ''
  })
  formError.value = ''
  showModal.value = true
}

async function save() {
  formError.value = ''
  if (!form.measuredAt) {
    formError.value = '请选择测量时间'
    return
  }
  if (!form.operatorName.trim()) {
    formError.value = '测量人必填'
    return
  }
  const len = Number(form.lengthMm)
  const wid = Number(form.widthMm)
  const hei = Number(form.heightMm)
  if ([len, wid, hei].some((v) => Number.isNaN(v) || v < 0)) {
    formError.value = '长、宽、高必须为非负数'
    return
  }
  let weight = null
  if (form.weightG !== '' && form.weightG != null) {
    weight = Number(form.weightG)
    if (Number.isNaN(weight) || weight < 0) {
      formError.value = '重量必须为非负数'
      return
    }
  }
  try {
    await api.post(`/finds/${current.value.id}/measurements`, {
      measuredAt: form.measuredAt,
      operatorName: form.operatorName.trim(),
      lengthMm: len,
      widthMm: wid,
      heightMm: hei,
      weightG: weight,
      caliperNote: form.caliperNote
    })
    showModal.value = false
    // 刷新度量历史与文物摘要
    const { data: find } = await api.get(`/finds/${current.value.id}`)
    current.value = find
    await loadSheets()
  } catch (e) {
    formError.value = e.response?.data?.error || '保存失败'
  }
}

onMounted(async () => {
  const findId = route.query.findId
  if (findId) {
    try {
      const { data: find } = await api.get(`/finds/${findId}`)
      current.value = find
      keyword.value = find.registerNo
      await loadSheets()
    } catch (e) {
      searchError.value = e.response?.data?.error || '文物加载失败'
    }
  }
})
</script>

<style scoped>
.search-label {
  min-width: 280px;
}

.result-list {
  margin-top: 0.5rem;
}

.find-head {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  margin-top: 1rem;
}

.back-btn {
  flex-shrink: 0;
}

.find-meta {
  flex: 1;
}

.find-reg {
  margin: 0 0 0.25rem;
}

.latest-box {
  margin-top: 0.5rem;
  padding: 0.6rem 0.8rem;
  border-radius: 8px;
  background: rgba(196, 165, 116, 0.14);
  font-size: 0.92rem;
}

.latest-title {
  font-weight: 600;
}

tr.latest td {
  background: rgba(196, 165, 116, 0.12);
  font-weight: 600;
}

.note-cell {
  max-width: 320px;
  white-space: pre-wrap;
}
</style>
