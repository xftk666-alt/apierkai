<template>
  <AppLayout>
    <div v-if="loading" class="card p-10">
      <div class="flex items-center justify-center py-16">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>
    </div>

    <div v-else-if="!walletEnabled" class="card p-10 text-center">
      <div class="mx-auto flex max-w-md flex-col items-center">
        <div class="rounded-full bg-gray-100 p-4 text-gray-400 dark:bg-dark-700 dark:text-dark-300">
          <Icon name="creditCard" size="xl" />
        </div>
        <h1 class="mt-5 text-xl font-semibold text-gray-900 dark:text-white">
          {{ t('wallet.notEnabledTitle') }}
        </h1>
        <p class="mt-3 text-sm leading-6 text-gray-500 dark:text-dark-400">
          {{ t('wallet.notEnabledDesc') }}
        </p>
      </div>
    </div>

    <div v-else class="space-y-6">
      <div class="card p-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ t('wallet.title') }}
            </h1>
            <p class="mt-3 max-w-3xl text-sm leading-6 text-gray-600 dark:text-dark-300">
              {{ t('wallet.description') }}
            </p>
          </div>

          <RouterLink to="/marketplace" class="btn btn-secondary">
            {{ t('wallet.goMarketplace') }}
          </RouterLink>
        </div>
      </div>

      <div class="grid gap-4 md:grid-cols-3">
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('wallet.summary.total') }}</p>
          <h2 class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ pagination.total }}</h2>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('wallet.summary.credit') }}</p>
          <h2 class="mt-2 text-2xl font-semibold text-emerald-600 dark:text-emerald-300">{{ totalCredit.toFixed(2) }}</h2>
        </div>
        <div class="card p-5">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('wallet.summary.debit') }}</p>
          <h2 class="mt-2 text-2xl font-semibold text-red-600 dark:text-red-300">{{ totalDebit.toFixed(2) }}</h2>
        </div>
      </div>

      <div
        v-if="errorMessage"
        class="card border border-red-200 bg-red-50 p-4 text-sm text-red-600 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-300"
      >
        {{ errorMessage }}
      </div>

      <div v-if="!errorMessage && records.length === 0" class="card p-10 text-center">
        <div class="mx-auto flex max-w-md flex-col items-center">
          <div class="rounded-full bg-gray-100 p-4 text-gray-400 dark:bg-dark-700 dark:text-dark-300">
            <Icon name="creditCard" size="xl" />
          </div>
          <h2 class="mt-5 text-xl font-semibold text-gray-900 dark:text-white">
            {{ t('wallet.emptyTitle') }}
          </h2>
          <p class="mt-3 text-sm leading-6 text-gray-500 dark:text-dark-400">
            {{ t('wallet.emptyDesc') }}
          </p>
        </div>
      </div>

      <div v-else class="space-y-4">
        <article
          v-for="item in records"
          :key="item.id"
          class="card border border-gray-200/80 p-5 dark:border-dark-700"
        >
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ walletDirectionLabel(item.direction) }}
                </h2>
                <span :class="directionBadgeClass(item.direction)">
                  {{ signedAmount(item.direction, item.change_amount) }}
                </span>
              </div>

              <div class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
                <div>
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('wallet.balanceBefore') }}
                  </p>
                  <p class="mt-1 text-sm text-gray-900 dark:text-white">{{ item.balance_before.toFixed(2) }}</p>
                </div>
                <div>
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('wallet.balanceAfter') }}
                  </p>
                  <p class="mt-1 text-sm text-gray-900 dark:text-white">{{ item.balance_after.toFixed(2) }}</p>
                </div>
                <div>
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('wallet.reason') }}
                  </p>
                  <p class="mt-1 text-sm text-gray-900 dark:text-white">{{ walletReasonLabel(item.reason_type) }}</p>
                </div>
                <div>
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('wallet.createdAt') }}
                  </p>
                  <p class="mt-1 text-sm text-gray-900 dark:text-white">{{ formatDateTime(item.created_at) }}</p>
                </div>
              </div>

              <div class="mt-4 text-sm text-gray-500 dark:text-dark-400">
                <span>{{ item.reason_detail }}</span>
                <span v-if="item.order_no" class="ml-2 font-mono">#{{ item.order_no }}</span>
              </div>
            </div>
          </div>
        </article>

        <div v-if="pagination.pages > 1" class="flex items-center justify-end gap-3">
          <button type="button" class="btn btn-secondary btn-sm" :disabled="pagination.page <= 1" @click="changePage(pagination.page - 1)">
            {{ t('wallet.previous') }}
          </button>
          <span class="text-sm text-gray-500 dark:text-dark-400">
            {{ pagination.page }} / {{ pagination.pages }}
          </span>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="pagination.page >= pagination.pages" @click="changePage(pagination.page + 1)">
            {{ t('wallet.next') }}
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
import type { CommerceWalletLedger, PaginatedResponse } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const errorMessage = ref('')
const records = ref<CommerceWalletLedger[]>([])
const pagination = ref<PaginatedResponse<CommerceWalletLedger>>({
  items: [],
  total: 0,
  page: 1,
  page_size: 10,
  pages: 1
})

const walletEnabled = computed(() => {
  const settings = appStore.cachedPublicSettings
  return Boolean(
    settings?.purchase_subscription_enabled &&
      settings.native_purchase_mode === 'native' &&
      settings.native_wallet_enabled
  )
})

const totalCredit = computed(() =>
  records.value
    .filter((item) => item.direction === 'credit')
    .reduce((sum, item) => sum + item.change_amount, 0)
)

const totalDebit = computed(() =>
  records.value
    .filter((item) => item.direction === 'debit')
    .reduce((sum, item) => sum + item.change_amount, 0)
)

onMounted(async () => {
  loading.value = true
  try {
    if (!appStore.publicSettingsLoaded) {
      await appStore.fetchPublicSettings()
    }
    if (!walletEnabled.value) {
      return
    }
    await loadWalletLedger(1)
  } catch (error: any) {
    errorMessage.value = error?.message || t('wallet.loadFailed')
    appStore.showError(errorMessage.value)
  } finally {
    loading.value = false
  }
})

async function loadWalletLedger(page: number): Promise<void> {
  const response = await commerceAPI.getWalletLedger(page, pagination.value.page_size || 10)
  pagination.value = response
  records.value = response.items
}

async function changePage(page: number): Promise<void> {
  if (page < 1 || page > pagination.value.pages) {
    return
  }
  loading.value = true
  try {
    await loadWalletLedger(page)
  } catch (error: any) {
    appStore.showError(error?.message || t('wallet.loadFailed'))
  } finally {
    loading.value = false
  }
}

function signedAmount(direction: string, amount: number): string {
  const prefix = direction === 'debit' ? '-' : '+'
  return `${prefix}${amount.toFixed(2)}`
}

function formatDateTime(value: string): string {
  try {
    return new Date(value).toLocaleString()
  } catch {
    return value
  }
}

function walletDirectionLabel(direction: string): string {
  return direction === 'debit' ? t('wallet.directions.debit') : t('wallet.directions.credit')
}

function walletReasonLabel(reasonType: string): string {
  if (reasonType === 'commerce_order_topup') {
    return t('wallet.reasonTypes.orderTopup')
  }
  return reasonType
}

function directionBadgeClass(direction: string): string {
  return direction === 'debit'
    ? 'rounded-full bg-red-50 px-2.5 py-1 text-xs font-medium text-red-700 dark:bg-red-500/10 dark:text-red-300'
    : 'rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'
}
</script>
