<template>
  <AppLayout>
    <div v-if="loading" class="card p-10">
      <div class="flex items-center justify-center py-16">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>
    </div>

    <div v-else-if="!nativeMarketplaceAvailable" class="card p-10 text-center">
      <div class="mx-auto flex max-w-md flex-col items-center">
        <div class="rounded-full bg-gray-100 p-4 text-gray-400 dark:bg-dark-700 dark:text-dark-300">
          <Icon name="grid" size="xl" />
        </div>
        <h1 class="mt-5 text-xl font-semibold text-gray-900 dark:text-white">
          {{ t('marketplace.notEnabledTitle') }}
        </h1>
        <p class="mt-3 text-sm leading-6 text-gray-500 dark:text-dark-400">
          {{ t('marketplace.notEnabledDesc') }}
        </p>
      </div>
    </div>

    <div v-else class="space-y-6">
      <div class="card p-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-sm font-medium text-primary-600 dark:text-primary-400">
              {{ t('marketplace.badge') }}
            </p>
            <h1 class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">
              {{ t('marketplace.heroTitle') }}
            </h1>
            <p class="mt-3 max-w-3xl text-sm leading-6 text-gray-600 dark:text-dark-300">
              {{ t('marketplace.heroDescription') }}
            </p>
          </div>

          <div class="flex flex-wrap gap-3">
            <RouterLink to="/subscriptions" class="btn btn-secondary">
              {{ t('marketplace.viewSubscriptions') }}
            </RouterLink>
            <RouterLink v-if="ordersEnabled" to="/orders" class="btn btn-secondary">
              {{ t('marketplace.viewOrders') }}
            </RouterLink>
            <RouterLink v-if="walletEnabled" to="/wallet" class="btn btn-secondary">
              {{ t('marketplace.viewWallet') }}
            </RouterLink>
            <a
              v-if="legacyPurchaseUrl"
              :href="legacyPurchaseUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-secondary"
            >
              {{ t('marketplace.openLegacyPurchase') }}
            </a>
          </div>
        </div>
      </div>

      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <div class="card p-5">
          <div class="flex items-center gap-3">
            <div
              class="rounded-xl bg-primary-50 p-3 text-primary-600 dark:bg-primary-500/10 dark:text-primary-300"
            >
              <Icon name="sparkles" size="lg" />
            </div>
            <div>
              <p class="text-sm text-gray-500 dark:text-dark-400">
                {{ t('marketplace.summary.models') }}
              </p>
              <h2 class="text-2xl font-semibold text-gray-900 dark:text-white">
                {{ models.length }}
              </h2>
            </div>
          </div>
        </div>

        <div class="card p-5">
          <div class="flex items-center gap-3">
            <div
              class="rounded-xl bg-emerald-50 p-3 text-emerald-600 dark:bg-emerald-500/10 dark:text-emerald-300"
            >
              <Icon name="creditCard" size="lg" />
            </div>
            <div>
              <p class="text-sm text-gray-500 dark:text-dark-400">
                {{ t('marketplace.summary.products') }}
              </p>
              <h2 class="text-2xl font-semibold text-gray-900 dark:text-white">
                {{ products.length }}
              </h2>
            </div>
          </div>
        </div>

        <div class="card p-5">
          <div class="flex items-center gap-3">
            <div
              class="rounded-xl bg-amber-50 p-3 text-amber-600 dark:bg-amber-500/10 dark:text-amber-300"
            >
              <Icon name="document" size="lg" />
            </div>
            <div>
              <p class="text-sm text-gray-500 dark:text-dark-400">
                {{ t('marketplace.summary.modules') }}
              </p>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                {{ moduleSummary }}
              </h2>
            </div>
          </div>
        </div>
      </div>

      <section v-if="paymentProviders.length" class="space-y-4">
        <div>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('marketplace.paymentMethodsTitle') }}
          </h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
            {{ t('marketplace.paymentMethodsDesc') }}
          </p>
        </div>

        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <div
            v-for="provider in paymentProviders"
            :key="provider.code"
            class="card border border-gray-200/80 p-5 dark:border-dark-700"
          >
            <div class="flex items-start gap-3">
              <div
                class="flex h-12 w-12 shrink-0 items-center justify-center overflow-hidden rounded-2xl bg-gray-100 text-sm font-semibold text-gray-600 dark:bg-dark-700 dark:text-dark-200"
              >
                <img
                  v-if="provider.icon"
                  :src="provider.icon"
                  :alt="provider.name"
                  class="h-full w-full object-cover"
                />
                <span v-else>{{ (provider.name || provider.code).slice(0, 1).toUpperCase() }}</span>
              </div>

              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                    {{ provider.name || provider.code }}
                  </h3>
                  <span
                    class="rounded-full bg-primary-50 px-2.5 py-1 text-xs font-medium text-primary-700 dark:bg-primary-500/10 dark:text-primary-300"
                  >
                    {{ provider.code }}
                  </span>
                </div>
                <p
                  v-if="provider.description"
                  class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300"
                >
                  {{ provider.description }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <div
        v-if="errorMessage"
        class="card border border-red-200 bg-red-50 p-4 text-sm text-red-600 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-300"
      >
        {{ errorMessage }}
      </div>

      <div
        v-if="!errorMessage && models.length === 0 && standaloneProducts.length === 0"
        class="card p-10 text-center"
      >
        <div class="mx-auto flex max-w-md flex-col items-center">
          <div class="rounded-full bg-gray-100 p-4 text-gray-400 dark:bg-dark-700 dark:text-dark-300">
            <Icon name="grid" size="xl" />
          </div>
          <h2 class="mt-5 text-xl font-semibold text-gray-900 dark:text-white">
            {{ t('marketplace.emptyTitle') }}
          </h2>
          <p class="mt-3 text-sm leading-6 text-gray-500 dark:text-dark-400">
            {{ t('marketplace.emptyDesc') }}
          </p>
        </div>
      </div>

      <section v-if="models.length" class="space-y-4">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('marketplace.modelsSectionTitle') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              {{ t('marketplace.modelsSectionDesc') }}
            </p>
          </div>
        </div>

        <div class="grid gap-4 xl:grid-cols-2">
          <div
            v-for="model in models"
            :key="model.id"
            class="card overflow-hidden border border-gray-200/80 p-5 dark:border-dark-700"
          >
            <div class="flex items-start justify-between gap-4">
              <div class="flex min-w-0 items-start gap-3">
                <div
                  class="flex h-12 w-12 shrink-0 items-center justify-center overflow-hidden rounded-2xl bg-gray-100 text-base font-semibold text-gray-600 dark:bg-dark-700 dark:text-dark-200"
                >
                  <img
                    v-if="model.icon"
                    :src="model.icon"
                    :alt="model.display_name"
                    class="h-full w-full object-cover"
                  />
                  <span v-else>{{ model.display_name.slice(0, 1).toUpperCase() }}</span>
                </div>

                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                      {{ model.display_name }}
                    </h3>
                    <span
                      v-if="model.recommended"
                      class="rounded-full bg-primary-50 px-2.5 py-1 text-xs font-medium text-primary-700 dark:bg-primary-500/10 dark:text-primary-300"
                    >
                      {{ t('marketplace.recommended') }}
                    </span>
                  </div>
                  <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                    {{ vendorLabel(model.vendor) }}
                  </p>
                  <p
                    v-if="model.description"
                    class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300"
                  >
                    {{ model.description }}
                  </p>
                </div>
              </div>
            </div>

            <div v-if="model.tags.length" class="mt-4 flex flex-wrap gap-2">
              <span
                v-for="tag in model.tags"
                :key="tag"
                class="rounded-full bg-gray-100 px-2.5 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-dark-300"
              >
                {{ tag }}
              </span>
            </div>

            <div v-if="model.offers.length" class="mt-5 space-y-3">
              <div
                v-for="offer in model.offers"
                :key="offer.binding_id"
                class="rounded-2xl border border-gray-200/80 p-4 dark:border-dark-700"
              >
                <div class="flex flex-wrap items-start justify-between gap-3">
                  <div class="min-w-0">
                    <div class="flex flex-wrap items-center gap-2">
                      <h4 class="text-base font-semibold text-gray-900 dark:text-white">
                        {{ offer.product.name }}
                      </h4>
                      <span
                        class="rounded-full bg-gray-100 px-2.5 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-dark-300"
                      >
                        {{ bindingTypeLabel(offer.binding_type) }}
                      </span>
                      <span
                        class="rounded-full bg-gray-100 px-2.5 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-dark-300"
                      >
                        {{ productTypeLabel(offer.product.product_type) }}
                      </span>
                    </div>
                    <p
                      v-if="offer.product.description"
                      class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300"
                    >
                      {{ offer.product.description }}
                    </p>
                  </div>

                  <a
                    v-if="legacyPurchaseUrl"
                    :href="legacyPurchaseUrl"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="btn btn-secondary btn-sm"
                  >
                    {{ t('marketplace.openLegacyPurchase') }}
                  </a>
                </div>

                <div v-if="offer.product.tags.length" class="mt-3 flex flex-wrap gap-2">
                  <span
                    v-for="tag in offer.product.tags"
                    :key="`${offer.binding_id}-${tag}`"
                    class="rounded-full bg-primary-50 px-2.5 py-1 text-xs text-primary-700 dark:bg-primary-500/10 dark:text-primary-300"
                  >
                    {{ tag }}
                  </span>
                </div>

                <div class="mt-4 space-y-3">
                  <div>
                    <p class="mb-2 text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">
                      {{ t('marketplace.prices') }}
                    </p>
                    <div class="flex flex-wrap gap-2">
                      <span
                        v-for="price in offer.product.prices"
                        :key="price.id"
                        class="rounded-full bg-emerald-50 px-3 py-1.5 text-sm font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300"
                      >
                        {{ formatPrice(price.amount, price.currency) }} / {{ priceTypeLabel(price.price_type) }}
                      </span>
                    </div>
                    <div class="mt-3 flex flex-wrap gap-2">
                      <template v-if="paymentProviders.length === 0">
                        <button
                          v-for="price in offer.product.prices"
                          :key="`order-${offer.product.id}-${price.id}`"
                          type="button"
                          class="btn btn-primary btn-sm"
                          :disabled="isCreatingOrder(offer.product.id, price.id)"
                          @click="handleCreateOrder(offer.product.id, price.id)"
                        >
                          {{
                            isCreatingOrder(offer.product.id, price.id)
                              ? t('marketplace.creatingOrder')
                              : orderButtonLabel(price)
                          }}
                        </button>
                      </template>
                      <template v-else>
                        <template
                          v-for="price in offer.product.prices"
                          :key="`order-group-${offer.product.id}-${price.id}`"
                        >
                          <button
                            v-for="provider in paymentProviders"
                            :key="`order-${offer.product.id}-${price.id}-${provider.code}`"
                            type="button"
                            class="btn btn-primary btn-sm"
                            :disabled="isCreatingOrder(offer.product.id, price.id, provider.code)"
                            @click="handleCreateOrder(offer.product.id, price.id, provider.code)"
                          >
                            {{
                              isCreatingOrder(offer.product.id, price.id, provider.code)
                                ? t('marketplace.creatingOrder')
                                : orderButtonLabel(price, provider)
                            }}
                          </button>
                        </template>
                      </template>
                    </div>
                  </div>

                  <div v-if="offer.product.grants.length">
                    <p class="mb-2 text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">
                      {{ t('marketplace.entitlements') }}
                    </p>
                    <div class="flex flex-wrap gap-2">
                      <span
                        v-for="grant in offer.product.grants"
                        :key="grant.id"
                        class="rounded-full bg-amber-50 px-3 py-1.5 text-sm text-amber-700 dark:bg-amber-500/10 dark:text-amber-300"
                      >
                        {{ grantLabel(grant) }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div
              v-else
              class="mt-5 rounded-2xl border border-dashed border-gray-200 p-4 text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400"
            >
              {{ t('marketplace.noActiveOffers') }}
            </div>
          </div>
        </div>
      </section>

      <section v-if="standaloneProducts.length" class="space-y-4">
        <div>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('marketplace.standaloneSectionTitle') }}
          </h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
            {{ t('marketplace.standaloneSectionDesc') }}
          </p>
        </div>

        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <div
            v-for="product in standaloneProducts"
            :key="product.id"
            class="card overflow-hidden border border-gray-200/80 p-5 dark:border-dark-700"
          >
            <div class="flex items-start gap-3">
              <div
                class="flex h-12 w-12 shrink-0 items-center justify-center overflow-hidden rounded-2xl bg-gray-100 text-base font-semibold text-gray-600 dark:bg-dark-700 dark:text-dark-200"
              >
                <img
                  v-if="product.cover_image"
                  :src="product.cover_image"
                  :alt="product.name"
                  class="h-full w-full object-cover"
                />
                <Icon v-else name="creditCard" size="lg" />
              </div>

              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                    {{ product.name }}
                  </h3>
                  <span
                    v-if="product.recommended"
                    class="rounded-full bg-primary-50 px-2.5 py-1 text-xs font-medium text-primary-700 dark:bg-primary-500/10 dark:text-primary-300"
                  >
                    {{ t('marketplace.recommended') }}
                  </span>
                </div>
                <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-dark-300">
                  {{ product.description || productTypeLabel(product.product_type) }}
                </p>
              </div>
            </div>

            <div v-if="product.prices.length" class="mt-4 flex flex-wrap gap-2">
              <span
                v-for="price in product.prices"
                :key="price.id"
                class="rounded-full bg-emerald-50 px-3 py-1.5 text-sm font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300"
              >
                {{ formatPrice(price.amount, price.currency) }} / {{ priceTypeLabel(price.price_type) }}
              </span>
            </div>

            <div v-if="product.prices.length" class="mt-4 flex flex-wrap gap-2">
              <template v-if="paymentProviders.length === 0">
                <button
                  v-for="price in product.prices"
                  :key="`standalone-order-${product.id}-${price.id}`"
                  type="button"
                  class="btn btn-primary btn-sm"
                  :disabled="isCreatingOrder(product.id, price.id)"
                  @click="handleCreateOrder(product.id, price.id)"
                >
                  {{
                    isCreatingOrder(product.id, price.id)
                      ? t('marketplace.creatingOrder')
                      : orderButtonLabel(price)
                  }}
                </button>
              </template>
              <template v-else>
                <template
                  v-for="price in product.prices"
                  :key="`standalone-order-group-${product.id}-${price.id}`"
                >
                  <button
                    v-for="provider in paymentProviders"
                    :key="`standalone-order-${product.id}-${price.id}-${provider.code}`"
                    type="button"
                    class="btn btn-primary btn-sm"
                    :disabled="isCreatingOrder(product.id, price.id, provider.code)"
                    @click="handleCreateOrder(product.id, price.id, provider.code)"
                  >
                    {{
                      isCreatingOrder(product.id, price.id, provider.code)
                        ? t('marketplace.creatingOrder')
                        : orderButtonLabel(price, provider)
                    }}
                  </button>
                </template>
              </template>
            </div>

            <div v-if="product.grants.length" class="mt-4 flex flex-wrap gap-2">
              <span
                v-for="grant in product.grants"
                :key="grant.id"
                class="rounded-full bg-amber-50 px-3 py-1.5 text-sm text-amber-700 dark:bg-amber-500/10 dark:text-amber-300"
              >
                {{ grantLabel(grant) }}
              </span>
            </div>

            <a
              v-if="legacyPurchaseUrl"
              :href="legacyPurchaseUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-secondary btn-sm mt-4"
            >
              {{ t('marketplace.openLegacyPurchase') }}
            </a>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { commerceAPI } from '@/api'
import { useAppStore } from '@/stores'
import type {
  CommerceCatalog,
  CommerceCatalogGrant,
  CommerceCatalogModel,
  CommerceCatalogPrice,
  CommerceCatalogProduct,
  CommercePaymentProviderOption
} from '@/types'

const { t } = useI18n()
const appStore = useAppStore()
const router = useRouter()

const loading = ref(true)
const errorMessage = ref('')
const catalog = ref<CommerceCatalog | null>(null)
const creatingOrderKey = ref('')

const nativeMarketplaceAvailable = computed(() => {
  const settings = appStore.cachedPublicSettings
  return Boolean(
    settings?.purchase_subscription_enabled &&
      settings.native_purchase_mode === 'native' &&
      settings.native_marketplace_enabled
  )
})

const models = computed<CommerceCatalogModel[]>(() => catalog.value?.models ?? [])
const products = computed<CommerceCatalogProduct[]>(() => catalog.value?.products ?? [])
const paymentProviders = computed<CommercePaymentProviderOption[]>(
  () => catalog.value?.features.payment_providers ?? []
)

const linkedProductIds = computed(() => {
  const ids = new Set<number>()
  for (const model of models.value) {
    for (const offer of model.offers) {
      ids.add(offer.product.id)
    }
  }
  return ids
})

const standaloneProducts = computed(() =>
  products.value.filter((product) => !linkedProductIds.value.has(product.id))
)

const legacyPurchaseUrl = computed(() => {
  const fromCatalog = catalog.value?.features.legacy_purchase_url?.trim()
  if (fromCatalog) {
    return fromCatalog
  }
  return (appStore.cachedPublicSettings?.purchase_subscription_url || '').trim()
})

const ordersEnabled = computed(() => {
  if (catalog.value) {
    return catalog.value.features.orders_enabled
  }
  return Boolean(appStore.cachedPublicSettings?.native_orders_enabled)
})

const walletEnabled = computed(() => {
  if (catalog.value) {
    return catalog.value.features.wallet_enabled
  }
  return Boolean(appStore.cachedPublicSettings?.native_wallet_enabled)
})

const moduleSummary = computed(() => {
  const features = catalog.value?.features
  const parts: string[] = []
  if (features?.wallet_enabled) {
    parts.push(t('marketplace.cards.wallet.title'))
  }
  if (features?.orders_enabled) {
    parts.push(t('marketplace.cards.orders.title'))
  }
  if (parts.length === 0) {
    return t('marketplace.comingSoon')
  }
  return parts.join(' / ')
})

onMounted(async () => {
  loading.value = true
  try {
    if (!appStore.publicSettingsLoaded) {
      await appStore.fetchPublicSettings()
    }
    if (!nativeMarketplaceAvailable.value) {
      return
    }
    catalog.value = await commerceAPI.getCatalog()
  } catch (error: any) {
    errorMessage.value = error?.message || t('marketplace.loadFailed')
    appStore.showError(errorMessage.value)
  } finally {
    loading.value = false
  }
})

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

function vendorLabel(vendor: string): string {
  return vendor ? `${t('marketplace.vendor')}: ${vendor}` : t('marketplace.vendorUnknown')
}

function productTypeLabel(type: string): string {
  switch (type) {
    case 'topup_balance':
      return t('marketplace.productTypes.topupBalance')
    case 'subscription_group':
      return t('marketplace.productTypes.subscriptionGroup')
    case 'standard_group_access':
      return t('marketplace.productTypes.standardGroupAccess')
    case 'combo':
      return t('marketplace.productTypes.combo')
    default:
      return type
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

function bindingTypeLabel(type: string): string {
  switch (type) {
    case 'primary':
      return t('marketplace.bindingTypes.primary')
    case 'upsell':
      return t('marketplace.bindingTypes.upsell')
    case 'topup':
      return t('marketplace.bindingTypes.topup')
    default:
      return type
  }
}

function grantLabel(grant: CommerceCatalogGrant): string {
  const pieces = [grant.group_name || `#${grant.group_id}`]
  switch (grant.grant_type) {
    case 'subscription':
      pieces.push(t('marketplace.grantTypes.subscription'))
      break
    case 'allowed_group':
      pieces.push(t('marketplace.grantTypes.allowedGroup'))
      break
    default:
      pieces.push(grant.grant_type)
      break
  }
  if (grant.validity_days > 0) {
    pieces.push(`${grant.validity_days}${t('marketplace.daysSuffix')}`)
  }
  return pieces.join(' / ')
}

function orderButtonLabel(
  price: CommerceCatalogPrice,
  paymentProvider?: CommercePaymentProviderOption
): string {
  const summary = `${formatPrice(price.amount, price.currency)} / ${priceTypeLabel(price.price_type)}`
  if (!paymentProvider) {
    return `${t('marketplace.createOrder')} · ${summary}`
  }
  return `${paymentProvider.name || paymentProvider.code} · ${summary}`
}

function orderActionKey(productId: number, priceId: number, paymentProvider?: string): string {
  return `${productId}:${priceId}:${paymentProvider || ''}`
}

function isCreatingOrder(productId: number, priceId: number, paymentProvider?: string): boolean {
  return creatingOrderKey.value === orderActionKey(productId, priceId, paymentProvider)
}

async function handleCreateOrder(
  productId: number,
  priceId: number,
  paymentProvider?: string
): Promise<void> {
  const actionKey = orderActionKey(productId, priceId, paymentProvider)
  creatingOrderKey.value = actionKey
  try {
    const order = await commerceAPI.createOrder({
      product_id: productId,
      price_id: priceId,
      payment_provider: paymentProvider
    })
    appStore.showSuccess(t('marketplace.orderCreated', { orderNo: order.order_no }))

    const checkoutURL = order.payment_action?.checkout_url?.trim()
    if (checkoutURL) {
      const checkoutWindow = window.open('', '_blank')
      if (checkoutWindow) {
        checkoutWindow.opener = null
        checkoutWindow.location.href = checkoutURL
        if (ordersEnabled.value) {
          await router.push('/orders')
        }
        return
      }

      window.location.href = checkoutURL
      return
    }

    if (ordersEnabled.value) {
      await router.push('/orders')
    }
  } catch (error: any) {
    appStore.showError(error?.response?.data?.detail || error?.message || t('marketplace.orderCreateFailed'))
  } finally {
    if (creatingOrderKey.value === actionKey) {
      creatingOrderKey.value = ''
    }
  }
}
</script>
