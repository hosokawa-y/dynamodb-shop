// User types
export interface User {
  id: string
  email: string
  name: string
  createdAt: string
  updatedAt: string
}

export interface RegisterRequest {
  email: string
  name: string
  password: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface AuthResponse {
  token: string
  user: User
}

// Product types
export interface Product {
  id: string
  name: string
  description: string
  price: number
  category: string
  stock: number
  imageUrl: string
  version: number
  createdAt: string
  updatedAt: string
}

export interface CreateProductRequest {
  name: string
  description: string
  price: number
  category: string
  stock: number
  imageUrl: string
}

export interface UpdateProductRequest {
  name: string
  description: string
  price: number
  category: string
  imageUrl: string
  version: number
}

// Cart types
export interface CartItem {
  userId: string
  productId: string
  productName: string
  price: number
  quantity: number
  version: number
  addedAt: string
  updatedAt: string
}

export interface Cart {
  items: CartItem[]
  totalPrice: number
  itemCount: number
}

export interface AddToCartRequest {
  productId: string
  quantity: number
}

export interface UpdateCartRequest {
  quantity: number
  version: number
}

// Order types
export interface OrderItem {
  orderId: string
  productId: string
  productName: string
  price: number
  quantity: number
  subtotal: number
}

export interface Order {
  id: string
  userId: string
  status: string
  totalAmount: number
  itemCount: number
  items?: OrderItem[]
  createdAt: string
  updatedAt: string
}

// Price History types
export interface PriceHistory {
  productId: string
  price: number
  changedBy: string
  timestamp: string
}

// Inventory Log types
export interface InventoryLog {
  productId: string
  changeType: string // IN, OUT, ADJUST
  quantity: number
  previousStock: number
  newStock: number
  reason: string
  orderId?: string
  timestamp: string
}

export interface UpdatePriceRequest {
  price: number
}

export interface AdjustStockRequest {
  changeType: string // IN, OUT, ADJUST
  quantity: number
  reason: string
}

// API response types
export interface ErrorResponse {
  error: string
}

export interface SuccessResponse {
  message: string
}
