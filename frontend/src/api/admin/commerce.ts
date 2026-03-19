/**
 * Admin commerce catalog API endpoints
 */

import { apiClient } from '../client'
import type {
  CommerceCatalog,
  CommerceCatalogModel,
  CommerceCatalogProduct,
  CommerceOrder,
  CommerceOrderManualCompleteRequest,
  CommerceWalletLedger,
  CommerceModelUpsertRequest,
  CommerceProductUpsertRequest,
  PaginatedResponse
} from '@/types'

export async function getCatalog(): Promise<CommerceCatalog> {
  const { data } = await apiClient.get<CommerceCatalog>('/admin/commerce/catalog')
  return data
}

export async function createProduct(
  request: CommerceProductUpsertRequest
): Promise<CommerceCatalogProduct> {
  const { data } = await apiClient.post<CommerceCatalogProduct>('/admin/commerce/products', request)
  return data
}

export async function updateProduct(
  id: number,
  request: CommerceProductUpsertRequest
): Promise<CommerceCatalogProduct> {
  const { data } = await apiClient.put<CommerceCatalogProduct>(`/admin/commerce/products/${id}`, request)
  return data
}

export async function createModel(
  request: CommerceModelUpsertRequest
): Promise<CommerceCatalogModel> {
  const { data } = await apiClient.post<CommerceCatalogModel>('/admin/commerce/models', request)
  return data
}

export async function updateModel(
  id: number,
  request: CommerceModelUpsertRequest
): Promise<CommerceCatalogModel> {
  const { data } = await apiClient.put<CommerceCatalogModel>(`/admin/commerce/models/${id}`, request)
  return data
}

export interface AdminCommerceOrderQuery {
  page?: number
  page_size?: number
  user_id?: number
  status?: string
  payment_status?: string
  order_no?: string
}

export interface AdminCommerceWalletLedgerQuery {
  page?: number
  page_size?: number
  user_id?: number
  order_id?: number
  direction?: string
  reason_type?: string
}

export async function listOrders(
  params: AdminCommerceOrderQuery = {}
): Promise<PaginatedResponse<CommerceOrder>> {
  const { data } = await apiClient.get<PaginatedResponse<CommerceOrder>>('/admin/commerce/orders', {
    params
  })
  return data
}

export async function getWalletLedger(
  params: AdminCommerceWalletLedgerQuery = {}
): Promise<PaginatedResponse<CommerceWalletLedger>> {
  const { data } = await apiClient.get<PaginatedResponse<CommerceWalletLedger>>('/admin/commerce/wallet/ledger', {
    params
  })
  return data
}

export async function manualCompleteOrder(
  id: number,
  request: CommerceOrderManualCompleteRequest = {}
): Promise<CommerceOrder> {
  const { data } = await apiClient.post<CommerceOrder>(`/admin/commerce/orders/${id}/manual-complete`, request)
  return data
}

const commerceAPI = {
  getCatalog,
  createProduct,
  updateProduct,
  createModel,
  updateModel,
  listOrders,
  getWalletLedger,
  manualCompleteOrder
}

export default commerceAPI
