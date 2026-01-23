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

// API response types
export interface ErrorResponse {
  error: string
}

export interface SuccessResponse {
  message: string
}
