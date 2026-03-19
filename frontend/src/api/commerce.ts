import { apiClient } from './client'
import type {
  CommerceCatalog,
  CommerceOrder,
  CommerceOrderCreateRequest,
  CommerceWalletLedger,
  PaginatedResponse
} from '@/types'

export async function getCommerceCatalog(): Promise<CommerceCatalog> {
  const { data } = await apiClient.get<CommerceCatalog>('/commerce/catalog')
  return data
}

export async function createCommerceOrder(request: CommerceOrderCreateRequest): Promise<CommerceOrder> {
  const { data } = await apiClient.post<CommerceOrder>('/commerce/orders', request)
  return data
}

export async function listCommerceOrders(
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<CommerceOrder>> {
  const { data } = await apiClient.get<PaginatedResponse<CommerceOrder>>('/commerce/orders', {
    params: {
      page,
      page_size: pageSize
    }
  })
  return data
}

export async function getCommerceWalletLedger(
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<CommerceWalletLedger>> {
  const { data } = await apiClient.get<PaginatedResponse<CommerceWalletLedger>>('/commerce/wallet/ledger', {
    params: {
      page,
      page_size: pageSize
    }
  })
  return data
}

export const commerceAPI = {
  getCatalog: getCommerceCatalog,
  createOrder: createCommerceOrder,
  listOrders: listCommerceOrders,
  getWalletLedger: getCommerceWalletLedger
}

export default commerceAPI
