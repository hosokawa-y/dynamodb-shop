<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { productsApi } from '@/api/products'
import type { Product, InventoryLog, AdjustStockRequest } from '@/api/types'

const products = ref<Product[]>([])
const selectedProduct = ref<Product | null>(null)
const inventoryLogs = ref<InventoryLog[]>([])
const loading = ref(false)
const logsLoading = ref(false)
const error = ref<string | null>(null)

// Stock adjustment form
const adjustForm = ref<AdjustStockRequest>({
  changeType: 'IN',
  quantity: 0,
  reason: '',
})
const adjusting = ref(false)
const adjustError = ref<string | null>(null)
const adjustSuccess = ref(false)

// Price update form
const newPrice = ref(0)
const updatingPrice = ref(false)
const priceError = ref<string | null>(null)
const priceSuccess = ref(false)

async function fetchProducts() {
  loading.value = true
  error.value = null
  try {
    products.value = await productsApi.list()
  } catch {
    error.value = 'Failed to load products'
  } finally {
    loading.value = false
  }
}

async function selectProduct(product: Product) {
  selectedProduct.value = product
  newPrice.value = product.price
  adjustForm.value = { changeType: 'IN', quantity: 0, reason: '' }
  adjustError.value = null
  adjustSuccess.value = false
  priceError.value = null
  priceSuccess.value = false
  await fetchInventoryLogs(product.id)
}

async function fetchInventoryLogs(productId: string) {
  logsLoading.value = true
  try {
    inventoryLogs.value = await productsApi.getInventoryLogs(productId, { limit: 20 })
  } catch {
    inventoryLogs.value = []
  } finally {
    logsLoading.value = false
  }
}

async function adjustStock() {
  if (!selectedProduct.value || adjusting.value) return

  adjusting.value = true
  adjustError.value = null
  adjustSuccess.value = false

  try {
    await productsApi.adjustStock(selectedProduct.value.id, adjustForm.value)
    adjustSuccess.value = true
    adjustForm.value = { changeType: 'IN', quantity: 0, reason: '' }

    // Refresh product and logs
    const updatedProduct = await productsApi.getById(selectedProduct.value.id)
    selectedProduct.value = updatedProduct
    const index = products.value.findIndex((p) => p.id === updatedProduct.id)
    if (index !== -1) {
      products.value[index] = updatedProduct
    }
    await fetchInventoryLogs(selectedProduct.value.id)

    setTimeout(() => {
      adjustSuccess.value = false
    }, 3000)
  } catch {
    adjustError.value = 'Failed to adjust stock'
  } finally {
    adjusting.value = false
  }
}

async function updatePrice() {
  if (!selectedProduct.value || updatingPrice.value) return

  updatingPrice.value = true
  priceError.value = null
  priceSuccess.value = false

  try {
    await productsApi.updatePrice(selectedProduct.value.id, { price: newPrice.value })
    priceSuccess.value = true

    // Refresh product
    const updatedProduct = await productsApi.getById(selectedProduct.value.id)
    selectedProduct.value = updatedProduct
    const index = products.value.findIndex((p) => p.id === updatedProduct.id)
    if (index !== -1) {
      products.value[index] = updatedProduct
    }

    setTimeout(() => {
      priceSuccess.value = false
    }, 3000)
  } catch {
    priceError.value = 'Failed to update price'
  } finally {
    updatingPrice.value = false
  }
}

function formatPrice(price: number): string {
  return new Intl.NumberFormat('ja-JP', {
    style: 'currency',
    currency: 'JPY',
  }).format(price)
}

function formatDateTime(dateString: string): string {
  return new Date(dateString).toLocaleString('ja-JP')
}

function getChangeTypeLabel(type: string): string {
  switch (type) {
    case 'IN':
      return 'In'
    case 'OUT':
      return 'Out'
    case 'ADJUST':
      return 'Adjust'
    default:
      return type
  }
}

function getChangeTypeClass(type: string): string {
  switch (type) {
    case 'IN':
      return 'type-in'
    case 'OUT':
      return 'type-out'
    case 'ADJUST':
      return 'type-adjust'
    default:
      return ''
  }
}

onMounted(() => {
  fetchProducts()
})
</script>

<template>
  <div class="inventory-management">
    <h1>Inventory Management</h1>

    <div v-if="loading" class="loading">Loading products...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else class="management-container">
      <!-- Product List -->
      <div class="product-list">
        <h2>Products</h2>
        <ul>
          <li
            v-for="product in products"
            :key="product.id"
            :class="{ selected: selectedProduct?.id === product.id }"
            @click="selectProduct(product)"
          >
            <span class="product-name">{{ product.name }}</span>
            <span class="product-stock" :class="{ 'low-stock': product.stock < 10 }">
              Stock: {{ product.stock }}
            </span>
          </li>
        </ul>
      </div>

      <!-- Product Detail / Management -->
      <div v-if="selectedProduct" class="product-management">
        <h2>{{ selectedProduct.name }}</h2>

        <div class="current-info">
          <p><strong>Current Stock:</strong> {{ selectedProduct.stock }}</p>
          <p><strong>Current Price:</strong> {{ formatPrice(selectedProduct.price) }}</p>
        </div>

        <!-- Price Update -->
        <div class="management-section">
          <h3>Update Price</h3>
          <div class="form-row">
            <label>
              New Price:
              <input v-model.number="newPrice" type="number" min="0" />
            </label>
            <button :disabled="updatingPrice || newPrice <= 0" @click="updatePrice">
              {{ updatingPrice ? 'Updating...' : 'Update Price' }}
            </button>
          </div>
          <p v-if="priceSuccess" class="success">Price updated successfully!</p>
          <p v-if="priceError" class="error">{{ priceError }}</p>
        </div>

        <!-- Stock Adjustment -->
        <div class="management-section">
          <h3>Adjust Stock</h3>
          <div class="form-group">
            <label>
              Change Type:
              <select v-model="adjustForm.changeType">
                <option value="IN">In (Add)</option>
                <option value="OUT">Out (Remove)</option>
                <option value="ADJUST">Adjust (Set)</option>
              </select>
            </label>
          </div>
          <div class="form-group">
            <label>
              {{ adjustForm.changeType === 'ADJUST' ? 'New Stock:' : 'Quantity:' }}
              <input v-model.number="adjustForm.quantity" type="number" min="0" />
            </label>
          </div>
          <div class="form-group">
            <label>
              Reason:
              <input v-model="adjustForm.reason" type="text" placeholder="Enter reason..." />
            </label>
          </div>
          <button :disabled="adjusting || !adjustForm.reason || adjustForm.quantity < 0" @click="adjustStock">
            {{ adjusting ? 'Adjusting...' : 'Adjust Stock' }}
          </button>
          <p v-if="adjustSuccess" class="success">Stock adjusted successfully!</p>
          <p v-if="adjustError" class="error">{{ adjustError }}</p>
        </div>

        <!-- Inventory Logs -->
        <div class="management-section">
          <h3>Inventory History</h3>
          <div v-if="logsLoading" class="loading">Loading...</div>
          <div v-else-if="inventoryLogs.length === 0" class="no-data">No inventory history</div>
          <table v-else class="logs-table">
            <thead>
              <tr>
                <th>Date</th>
                <th>Type</th>
                <th>Qty</th>
                <th>Before</th>
                <th>After</th>
                <th>Reason</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(log, index) in inventoryLogs" :key="index">
                <td>{{ formatDateTime(log.timestamp) }}</td>
                <td>
                  <span :class="['type-badge', getChangeTypeClass(log.changeType)]">
                    {{ getChangeTypeLabel(log.changeType) }}
                  </span>
                </td>
                <td>{{ log.quantity }}</td>
                <td>{{ log.previousStock }}</td>
                <td>{{ log.newStock }}</td>
                <td>{{ log.reason }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div v-else class="no-selection">
        <p>Select a product to manage inventory</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.inventory-management {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.inventory-management h1 {
  margin: 0 0 1.5rem;
  color: #333;
}

.loading,
.error,
.no-data,
.no-selection {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.error {
  color: #c00;
}

.success {
  color: #27ae60;
  font-size: 0.9rem;
  margin-top: 0.5rem;
}

.management-container {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: 2rem;
}

@media (max-width: 768px) {
  .management-container {
    grid-template-columns: 1fr;
  }
}

.product-list {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  padding: 1.5rem;
}

.product-list h2 {
  margin: 0 0 1rem;
  font-size: 1.1rem;
  color: #333;
}

.product-list ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.product-list li {
  padding: 0.75rem;
  border-bottom: 1px solid #eee;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: background 0.2s;
}

.product-list li:hover {
  background: #f5f5f5;
}

.product-list li.selected {
  background: #e8f4fd;
}

.product-name {
  font-weight: 500;
}

.product-stock {
  font-size: 0.85rem;
  color: #666;
}

.product-stock.low-stock {
  color: #e74c3c;
  font-weight: 500;
}

.product-management {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  padding: 1.5rem;
}

.product-management h2 {
  margin: 0 0 1rem;
  color: #333;
}

.current-info {
  background: #f9f9f9;
  padding: 1rem;
  border-radius: 4px;
  margin-bottom: 1.5rem;
}

.current-info p {
  margin: 0.25rem 0;
}

.management-section {
  border-top: 1px solid #eee;
  padding-top: 1.5rem;
  margin-top: 1.5rem;
}

.management-section h3 {
  margin: 0 0 1rem;
  font-size: 1rem;
  color: #333;
}

.form-row {
  display: flex;
  gap: 1rem;
  align-items: flex-end;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label,
.form-row label {
  display: block;
  font-size: 0.9rem;
  color: #666;
}

.form-group input,
.form-group select,
.form-row input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
  margin-top: 0.25rem;
}

.form-row input {
  width: 150px;
}

button {
  padding: 0.5rem 1rem;
  background: #4a90d9;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: background 0.2s;
}

button:hover:not(:disabled) {
  background: #3a7bc8;
}

button:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.logs-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.logs-table th,
.logs-table td {
  padding: 0.5rem;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.logs-table th {
  color: #666;
  font-weight: 500;
}

.type-badge {
  display: inline-block;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
}

.type-in {
  background: #d4edda;
  color: #155724;
}

.type-out {
  background: #f8d7da;
  color: #721c24;
}

.type-adjust {
  background: #fff3cd;
  color: #856404;
}
</style>
