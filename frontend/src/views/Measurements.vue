<template>
  <div>
    <div class="toolbar">
      <div>
        <h2 class="page-title">器物度量单</h2>
        <p class="page-sub">按登记号检索文物，记录并查看历次度量（按时间倒序）</p>
      </div>
    </div>

    <div class="card">
      <div class="search-row">
        <label>
          登记号
          <input
            v-model.trim="keyword"
            placeholder="如 EL-2024-0001"
            @keyup.enter="search"
          />
        </label>
        <button class="btn" @click="search">检索</button>
      </div>
      <p v-if="searchError" class="error">{{ searchError }}</p>
      <div v-if="candidates.length > 1" class="candidates">
        <div
          v-for="f in candidates"
          :key="f.id"
          class="candidate"
          :class="{ active: currentFind && currentFind.id === f.id }"
          @click="selectFind(f)"
        >
          <span class="tag">{{ f.registerNo }}</span>
          {{ f.artifactType }} · 探方 {{ f.unit?.code || '-' }} · {{ f.materialName || '未指定材质' }}
        </div>
      </div>
    </div>

    <template v-if="currentFind">
      <div class="card find-card">
        <div class="find-head">
          <div>
            <h3 class="find-title">{{ currentFind.registerNo }} · {{ currentFind.artifactType }}</h3>
            <p class="page-sub">
              探方 {{ currentFind.unit?.code || '-' }}
              · 材质 {{ currentFind.materialName || currentFind.material?.name || '-' }}
              · {{ currentFind.completeness || '-' }}
              · 存放 {{ currentFind.storageLoc || '-' }}
            </p>
          </div>
          <button class="btn" @click="openCreate">新增度量</button>
        </div>
        <p v-if="currentFind.description" class="find-desc">描述：{{ currentFind.description }}</p>
        <p v-if="currentFind.latestDimsSummary" class="latest-line">
          最近度量：{{ formatTime(currentFind.latestMeasuredAt) }} · {{ currentFind.latestDimsSummary }}
          <template v-if="currentFind.latestWeightG != null"> · 重 {{ currentFind.latestWeightG }} g</template>
        </p>
        <p v-else class="page-sub">暂无度量记录</p>
      </div>

      <div class="card">
        <h3 class="history-title">历史度量（{{ measurements.length }}）</h3>
        <table class="table">
          <thead>
            <tr>
              <th>度量时间</th>
              <th>长 (mm)</th>
              <th>宽 (mm)</th>
              <th>高 (mm)</th>
              <th>重量 (g)</th>
              <th>操作人</th>
              <th>卡尺备注</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="m in measurements" :key="m.id">
              <td>{{ formatTime(m.measuredAt) }}</td>
              <td>{{ m.lengthMm }}</td>
              <td>{{ m.widthMm }}</td>
              <td>{{ m.heightMm }}</td>
              <td>{{ m.weightG == null ? '-' : m.weightG }}</td>
              <td>{{ m.operatorName }}</td>
              <td>{{ m.caliperNote || '-' }}</td>
              <td>
                <button class="btn secondary small" @click="openEdit(m)">编辑</button>
                <button class="btn danger small" @click="remove(m)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
        <p v-if="!measurements.length" class="page-sub">暂无度量记录，点击「新增度量」录入第一份度量单</p>
        <p v-if="listError" class="error">{{ listError }}</p>
      </div>
    </template>

    <div v-if="showModal" class="modal-mask" @click.self="showModal = false">
      <div class="modal">
        <h3>{{ form.id ? '编辑度量单' : '新增度量单' }} — {{ currentFind?.registerNo }}</h3>
        <div class="form-grid">
          <label>
            度量时间
            <input v-model="form.measuredAt" type="datetime-local" />
          </label>
          <label>
            操作人
            <input v-model="form.operatorName" placeholder="测量人姓名" />
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
            重量 (g，可空)
            <input v-model.number="form.weightG" type="number" min="0" step="0.1" />
          </label>
          <label class="full">
            卡尺备注
            <textarea v-model="form.caliperNote" placeholder="量具、测量方式、复测原因等" />
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
import { useRoute } from 'vue-router'
import api from '../api/http'

const route = useRoute()

const keyword = ref('')
const candidates = ref([])
const currentFind = ref(null)
const measurements = ref([])
const searchError = ref('')
const listError = ref('')
const formError = ref('')
const showModal = ref(false)

const form = reactive({
  id: null,
  measuredAt: '',
  lengthMm: null,
  widthMm: null,
  heightMm: null,
  weightG: '',
  caliperNote: '',
  operatorName: ''
})

function pad(n) {
  return String(n).padStart(2, '0')
}

function nowLocal() {
  const d = new Date()
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function formatTime(v) {
  if (!v) return '-'
  return String(v).slice(0, 16).replace('T', ' ')
}

async function search() {
  searchError.value = ''
  if (!keyword.value) {
    searchError.value = '请输入登记号'
    return
  }
  try {
    const { data } = await api.get('/finds', { params: { registerNo: keyword.value } })
    candidates.value = data
    if (!data.length) {
      currentFind.value = null
      measurements.value = []
      searchError.value = '未找到匹配登记号的文物'
      return
    }
    const exact = data.find((f) => f.registerNo === keyword.value) || data[0]
    await selectFind(exact)
  } catch (e) {
    searchError.value = e.response?.data?.error || '检索失败'
  }
}

async function selectFind(f) {
  currentFind.value = f
  await loadMeasurements()
}

async function loadMeasurements() {
  listError.value = ''
  try {
    const [{ data: list }, { data: fresh }] = await Promise.all([
      api.get(`/finds/${currentFind.value.id}/measurements`),
      api.get(`/finds/${currentFind.value.id}`)
    ])
    measurements.value = list
    currentFind.value = fresh
  } catch (e) {
    listError.value = e.response?.data?.error || '加载度量记录失败'
  }
}

function openCreate() {
  Object.assign(form, {
    id: null,
    measuredAt: nowLocal(),
    lengthMm: null,
    widthMm: null,
    heightMm: null,
    weightG: '',
    caliperNote: '',
    operatorName: ''
  })
  formError.value = ''
  showModal.value = true
}

function openEdit(m) {
  Object.assign(form, {
    id: m.id,
    measuredAt: String(m.measuredAt).slice(0, 16),
    lengthMm: m.lengthMm,
    widthMm: m.widthMm,
    heightMm: m.heightMm,
    weightG: m.weightG == null ? '' : m.weightG,
    caliperNote: m.caliperNote || '',
    operatorName: m.operatorName || ''
  })
  formError.value = ''
  showModal.value = true
}

function numOrNull(v) {
  return v === '' || v === null || v === undefined ? null : Number(v)
}

async function save() {
  formError.value = ''
  const payload = {
    measuredAt: form.measuredAt,
    lengthMm: numOrNull(form.lengthMm),
    widthMm: numOrNull(form.widthMm),
    heightMm: numOrNull(form.heightMm),
    weightG: numOrNull(form.weightG),
    caliperNote: form.caliperNote,
    operatorName: form.operatorName.trim()
  }
  if (!payload.measuredAt) {
    formError.value = '请填写度量时间'
    return
  }
  if (payload.lengthMm == null || payload.widthMm == null || payload.heightMm == null) {
    formError.value = '请填写长、宽、高'
    return
  }
  if (payload.lengthMm < 0 || payload.widthMm < 0 || payload.heightMm < 0) {
    formError.value = '长、宽、高不能为负数'
    return
  }
  if (payload.weightG != null && payload.weightG < 0) {
    formError.value = '重量不能为负数'
    return
  }
  if (!payload.operatorName) {
    formError.value = '请填写操作人'
    return
  }
  try {
    if (form.id) {
      await api.put(`/measurements/${form.id}`, payload)
    } else {
      await api.post(`/finds/${currentFind.value.id}/measurements`, payload)
    }
    showModal.value = false
    await loadMeasurements()
  } catch (e) {
    formError.value = e.response?.data?.error || '保存失败'
  }
}

async function remove(m) {
  if (!confirm(`确认删除 ${formatTime(m.measuredAt)} 的度量单？`)) return
  try {
    await api.delete(`/measurements/${m.id}`)
    await loadMeasurements()
  } catch (e) {
    alert(e.response?.data?.error || '删除失败')
  }
}

onMounted(async () => {
  const rn = route.query.registerNo
  if (rn) {
    keyword.value = String(rn)
    await search()
  }
})
</script>

<style scoped>
.search-row {
  display: flex;
  gap: 0.75rem;
  align-items: flex-end;
  flex-wrap: wrap;
}

.search-row label {
  min-width: 260px;
  flex: 1;
  max-width: 420px;
}

.candidates {
  margin-top: 0.85rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.candidate {
  padding: 0.55rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 10px;
  cursor: pointer;
  display: flex;
  gap: 0.6rem;
  align-items: center;
}

.candidate:hover,
.candidate.active {
  background: #f6edde;
  border-color: var(--accent);
}

.find-card {
  margin-top: 1rem;
}

.find-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

.find-title {
  margin: 0 0 0.35rem;
}

.find-desc {
  margin: 0.75rem 0 0;
  font-size: 0.92rem;
}

.latest-line {
  margin: 0.5rem 0 0;
  color: var(--ok);
  font-size: 0.92rem;
}

.history-title {
  margin: 0 0 0.75rem;
}

.card + .card {
  margin-top: 1rem;
}
</style>
