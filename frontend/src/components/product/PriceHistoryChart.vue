<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'
import { productsApi } from '@/api/products'
import type { PriceHistory } from '@/api/types'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend)

const props = defineProps<{
  productId: string
}>()

const priceHistory = ref<PriceHistory[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const chartData = computed(() => {
  const sortedHistory = [...priceHistory.value].sort(
    (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime(),
  )

  return {
    labels: sortedHistory.map((h) => formatDateTime(h.timestamp)),
    datasets: [
      {
        label: 'Price',
        data: sortedHistory.map((h) => h.price),
        borderColor: '#4a90d9',
        backgroundColor: 'rgba(74, 144, 217, 0.1)',
        fill: true,
        tension: 0.1,
      },
    ],
  }
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      callbacks: {
        label: (context: { parsed: { y: number | null } }) => {
          return formatPrice(context.parsed.y ?? 0)
        },
      },
    },
  },
  scales: {
    y: {
      beginAtZero: false,
      ticks: {
        callback: (value: number | string) => formatPrice(Number(value)),
      },
    },
  },
}

function formatPrice(price: number): string {
  return new Intl.NumberFormat('ja-JP', {
    style: 'currency',
    currency: 'JPY',
  }).format(price)
}

function formatDateTime(dateString: string): string {
  return new Date(dateString).toLocaleDateString('ja-JP', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function fetchPriceHistory() {
  loading.value = true
  error.value = null
  try {
    priceHistory.value = await productsApi.getPriceHistory(props.productId, { limit: 30 })
  } catch {
    error.value = 'Failed to load price history'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchPriceHistory()
})

watch(
  () => props.productId,
  () => {
    fetchPriceHistory()
  },
)
</script>

<template>
  <div class="price-history-chart">
    <h3>Price History</h3>

    <div v-if="loading" class="loading">Loading...</div>

    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else-if="priceHistory.length === 0" class="no-data">No price history available</div>

    <div v-else class="chart-container">
      <Line :data="chartData" :options="chartOptions" />
    </div>

    <div v-if="priceHistory.length > 0" class="history-list">
      <h4>Recent Changes</h4>
      <table>
        <thead>
          <tr>
            <th>Date</th>
            <th>Price</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(history, index) in priceHistory.slice(0, 5)" :key="index">
            <td>{{ formatDateTime(history.timestamp) }}</td>
            <td>{{ formatPrice(history.price) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.price-history-chart {
  background: #f9f9f9;
  border-radius: 8px;
  padding: 1.5rem;
  margin-top: 1.5rem;
}

.price-history-chart h3 {
  margin: 0 0 1rem;
  color: #333;
  font-size: 1.1rem;
}

.loading,
.error,
.no-data {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.error {
  color: #c00;
}

.chart-container {
  height: 250px;
  margin-bottom: 1.5rem;
}

.history-list h4 {
  margin: 0 0 0.75rem;
  color: #666;
  font-size: 0.9rem;
}

.history-list table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.history-list th,
.history-list td {
  padding: 0.5rem;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.history-list th {
  color: #666;
  font-weight: 500;
}
</style>
