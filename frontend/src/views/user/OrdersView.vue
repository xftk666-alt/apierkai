<template>
  <AppLayout>
    <div v-if="loading" class="card p-10">
      <div class="flex items-center justify-center py-16">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>
    </div>

    <div v-else-if="!ordersEnabled" class="card p-10 text-center">
      <div class="mx-auto flex max-w-md flex-col items-center">
        <div class="rounded-full bg-gray-100 p-4 text-gray-400 dark:bg-dark-700 dark:text-dark-300">
          <Icon name="document" size="xl" />
        </div>
        <h1 class="mt-5 text-xl font-semibold text-gray-900 dark:text-white">
          {{ t('orders.notEnabledTitle') }}
        </h1>
        <p class="mt-3 text-sm leading-6 text-gray-500 dark:text-dark-400">
          {{ t('orders.notEnabledDesc') }}
        </p>
      </div>
    </div>

    <div v-else class="space-y-6">
      <div class="card p-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ t('orders.title') }}
            </h1>
            <p class="mt-3 max-w-3xl text-sm leading-6 text-gray-600 dark:text-dark-300">
              {{ t('orders.description') }}
            </p>
          </div>

          <RouterLink to="/marketplace" class="btn btn-secondary">
            {{ t('orders.goMarketplace') }}
          </RouterLink>
        </div>
      </div>

      <div class="grid gap-4 md:grid-cols-3">
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('orders.summary.total') }}</p>
          <h2 class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ pagination.total }}</h2>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('orders.summary.pending') }}</p>
          <h2 class="mt-2 text-2xl font-semibold text-amber-600 dark:text-amber-300">{{ pendingCount }}</h2>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('orders.summary.completed') }}</p>
          <h2 class="mt-2 text-2xl font-semibold text-emerald-600 dark:text-emerald-300">{{ completedCount }}</h2>
        </div>
      </div>

      <div
        v-if="errorMessage"
        class="card border border-red-200 bg-red-50 p-4 text-sm text-red-600 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-300"
      >
        {{ errorMessage }}
      </div>

      <div v-if="!errorMessage && orders.length === 0" class="card p-10 text-center">
        <div class="mx-auto flex max-w-md flex-col items-center">
          <div class="rounded-full bg-gray-100 p-4 text-gray-400 dark:bg-dark-700 dark:text-dark-300">
            <Icon name="document" size="xl" />
          </div>
          <h2 class="mt-5 text-xl font-semibold text-gray-900 dark:text-white">
            {{ t('orders.emptyTitle') }}
          </h2>
          <p class="mt-3 text-sm leading-6 text-gray-500 dark:text-dark-400">
            {{ t('orders.emptyDesc') }}
          </p>
          <RouterLink to="/marketplace" class="btn btn-primary mt-5">
            {{ t('orders.goMarketplace') }}
          </RouterLink>
        </div>
      </div>

      <div v-else class="space-y-4">
        <article
          v-for="order in orders"
          :key="order.id"
          class="card border border-gray-200/80 p-5 dark:border-dark-700"
        >
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ order.snapshot.product.name }}
                </h2>
                <span :class="statusBadgeClass(order.status)">
                  {{ orderStatusLabel(order.status) }}
                </span>
                <span :class="paymentBadgeClass(order.payment_status)">
                  {{ paymentStatusLabel(order.payment_status) }}
                </span>
              </div>

              <p class="mt-2 font-mono text-xs text-gray-500 dark:text-dark-400">
                {{ t('orders.orderNo') }}: {{ order.order_no }}
              </p>

              <div class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
                <div>
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('orders.price') }}
                  </p>
                  <p class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
                    {{ formatPrice(order.amount, order.currency) }} / {{ priceTypeLabel(order.snapshot.price.price_type) }}
                  </p>
                </div>
                <div>
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('orders.createdAt') }}
                  </p>
                  <p class="mt-1 text-sm text-gray-900 dark:text-white">
                    {{ formatDateTime(order.created_at) }}
                  </p>
                </div>
                <div>
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('orders.paidAt') }}
                  </p>
                  <p class="mt-1 text-sm text-gray-900 dark:text-white">
                    {{ order.paid_at ? formatDateTime(order.paid_at) : '-' }}
                  </p>
                </div>
                <div>
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('orders.completedAt') }}
                  </p>
                  <p class="mt-1 text-sm text-gray-900 dark:text-white">
                    {{ order.completed_at ? formatDateTime(order.completed_at) : '-' }}
                  </p>
                </div>
              </div>

              <div class="mt-4">
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('orders.entitlements') }}
                </p>
                <div v-if="order.snapshot.grants.length" class="mt-2 flex flex-wrap gap-2">
                  <span
                    v-for="grant in order.snapshot.grants"
                    :key="`${order.id}-${grant.id}-${grant.group_id}`"
                    class="rounded-full bg-amber-50 px-3 py-1.5 text-sm text-amber-700 dark:bg-amber-500/10 dark:text-amber-300"
                  >
                    {{ grantLabel(grant) }}
                  </span>
                </div>
                <p v-else class="mt-2 text-sm text-gray-500 dark:text-dark-400">-</p>
              </div>
            </div>
          </div>
        </article>

        <div v-if="pagination.pages > 1" class="flex items-center justify-end gap-3">
          <button type="button" class="btn btn-secondary btn-sm" :disabled="pagination.page <= 1" @click="changePage(pagination.page - 1)">
            {{ t('orders.previous') }}
          </button>
          <span class="text-sm text-gray-500 dark:text-dark-400">
            {{ pagination.page }} / {{ pagination.pages }}
          </span>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="pagination.page >= pagination.pages" @click="changePage(pagination.page + 1)">
            {{ t('orders.next') }}
          </button>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { commerceAPI } from '@/api'
import { useAppStore } from '@/stores'
import type { CommerceCatalogGrant, CommerceOrder, PaginatedResponse } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const errorMessage = ref('')
const orders = ref<CommerceOrder[]>([])
const pagination = ref<PaginatedResponse<CommerceOrder>>({
  items: [],
  total: 0,
  page: 1,
  page_size: 10,
  pages: 1
})

const ordersEnabled = computed(() => {
  const settings = appStore.cachedPublicSettings
  return Boolean(
    settings?.purchase_subscription_enabled &&
      settings.native_purchase_mode === 'native' &&
      settings.native_orders_enabled
  )
})

const pendingCount = computed(() => orders.value.filter((item) => item.status === 'pending').length)
const completedCount = computed(() => orders.value.filter((item) => item.status === 'completed').length)

onMounted(async () => {
  loading.value = true
  try {
    if (!appStore.publicSettingsLoaded) {
      await appStore.fetchPublicSettings()
    }
    if (!ordersEnabled.value) {
      return
    }
    await loadOrders(1)
  } catch (error: any) {
    errorMessage.value = error?.message || t('orders.loadFailed')
    appStore.showError(errorMessage.value)
  } finally {
    loading.value = false
  }
})

async function loadOrders(page: number): Promise<void> {
  const response = await commerceAPI.listOrders(page, pagination.value.page_size || 10)
  pagination.value = response
  orders.value = response.items
}

async function changePage(page: number): Promise<void> {
  if (page < 1 || page > pagination.value.pages) {
    return
  }
  loading.value = true
  try {
    await loadOrders(page)
  } catch (error: any) {
    appStore.showError(error?.message || t('orders.loadFailed'))
  } finally {
    loading.value = false
  }
}

function formatPrice(amount: number, currency: string): string {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
      maximumFractionDigits: 2
    }).format(amount)
  } catch {
    return `${currency} ${amount.toFixed(2)}`
  }
}

function formatDateTime(value: string): string {
  try {
    return new Date(value).toLocaleString()
  } catch {
    return value
  }
}

function priceTypeLabel(type: string): string {
  switch (type) {
    case 'one_time':
      return t('marketplace.priceTypes.oneTime')
    case 'monthly':
      return t('marketplace.priceTypes.monthly')
    case 'quarterly':
      return t('marketplace.priceTypes.quarterly')
    case 'yearly':
      return t('marketplace.priceTypes.yearly')
    default:
      return type
  }
}

function orderStatusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: t('orders.statuses.pending'),
    paid: t('orders.statuses.paid'),
    completed: t('orders.statuses.completed'),
    expired: t('orders.statuses.expired'),
    cancelled: t('orders.statuses.cancelled')
  }
  return map[status] || status
}

function paymentStatusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: t('orders.paymentStatuses.pending'),
    succeeded: t('orders.paymentStatuses.succeeded'),
    failed: t('orders.paymentStatuses.failed'),
    refunded: t('orders.paymentStatuses.refunded')
  }
  return map[status] || status
}

function grantLabel(grant: CommerceCatalogGrant): string {
  const parts = [grant.group_name || `#${grant.group_id}`]
  switch (grant.grant_type) {
    case 'subscription':
      parts.push(t('marketplace.grantTypes.subscription'))
      break
    case 'allowed_group':
      parts.push(t('marketplace.grantTypes.allowedGroup'))
      break
    default:
      parts.push(grant.grant_type)
      break
  }
  if (grant.validity_days > 0) {
    parts.push(`${grant.validity_days}${t('marketplace.daysSuffix')}`)
  }
  return parts.join(' / ')
}

function statusBadgeClass(status: string): string {
  switch (status) {
    case 'completed':
      return 'rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'
    case 'pending':
      return 'rounded-full bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-700 dark:bg-amber-500/10 dark:text-amber-300'
    default:
      return 'rounded-full bg-gray-100 px-2.5 py-1 text-xs font-medium text-gray-600 dark:bg-dark-700 dark:text-dark-300'
  }
}

function paymentBadgeClass(status: string): string {
  switch (status) {
    case 'succeeded':
      return 'rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'
    case 'pending':
      return 'rounded-full bg-primary-50 px-2.5 py-1 text-xs font-medium text-primary-700 dark:bg-primary-500/10 dark:text-primary-300'
    case 'failed':
      return 'rounded-full bg-red-50 px-2.5 py-1 text-xs font-medium text-red-700 dark:bg-red-500/10 dark:text-red-300'
    default:
      return 'rounded-full bg-gray-100 px-2.5 py-1 text-xs font-medium text-gray-600 dark:bg-dark-700 dark:text-dark-300'
  }
}
</script>
