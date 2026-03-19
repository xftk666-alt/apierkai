<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-4">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div class="flex flex-wrap items-center gap-2">
            <span class="rounded-full bg-primary-50 px-3 py-1 text-sm text-primary-700 dark:bg-primary-500/10 dark:text-primary-300">
              {{ t('marketplace.summary.products') }}: {{ products.length }}
            </span>
            <span class="rounded-full bg-emerald-50 px-3 py-1 text-sm text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300">
              {{ t('marketplace.summary.models') }}: {{ models.length }}
            </span>
            <span class="rounded-full bg-amber-50 px-3 py-1 text-sm text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">
              {{ t('admin.commerce.summary.recommended') }}: {{ recommendedCount }}
            </span>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <div class="inline-flex rounded-2xl bg-gray-100 p-1 dark:bg-dark-800">
              <button type="button" :class="tabClass('products')" @click="switchTab('products')">
                {{ t('marketplace.summary.products') }}
              </button>
              <button type="button" :class="tabClass('models')" @click="switchTab('models')">
                {{ t('marketplace.summary.models') }}
              </button>
              <button type="button" :class="tabClass('orders')" @click="switchTab('orders')">
                {{ t('nav.orders') }}
              </button>
              <button type="button" :class="tabClass('wallet')" @click="switchTab('wallet')">
                {{ t('nav.wallet') }}
              </button>
            </div>
            <button type="button" class="btn btn-secondary" :disabled="activeLoading" @click="refreshActiveTab">
              <Icon name="refresh" size="md" :class="activeLoading ? 'animate-spin' : ''" />
            </button>
            <button v-if="activeTab === 'products'" type="button" class="btn btn-primary" @click="openCreateProduct">
              <Icon name="plus" size="md" class="mr-2" />
              {{ t('admin.commerce.actions.createProduct') }}
            </button>
            <button v-else-if="activeTab === 'models'" type="button" class="btn btn-primary" @click="openCreateModel">
              <Icon name="plus" size="md" class="mr-2" />
              {{ t('admin.commerce.actions.createModel') }}
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="activeErrorMessage"
        class="card border border-red-200 bg-red-50 p-4 text-sm text-red-600 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-300"
      >
        {{ activeErrorMessage }}
      </div>

      <div v-if="loading && !catalog" class="card p-10 text-center text-gray-500 dark:text-dark-300">
        <Icon name="refresh" size="lg" class="mx-auto animate-spin" />
        <p class="mt-3">{{ t('common.loading') }}</p>
      </div>

      <div v-else-if="activeTab === 'products'" class="space-y-4">
        <div v-if="products.length" class="grid gap-4 xl:grid-cols-2">
          <article v-for="product in products" :key="product.id" class="card border border-gray-200/80 p-5 dark:border-dark-700">
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ product.name }}</h2>
                  <span :class="['badge', statusBadgeClass(product.status)]">{{ statusLabel(product.status) }}</span>
                  <span class="rounded-full bg-gray-100 px-2.5 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-dark-300">
                    {{ productTypeLabel(product.product_type) }}
                  </span>
                  <span
                    v-if="product.recommended"
                    class="rounded-full bg-primary-50 px-2.5 py-1 text-xs font-medium text-primary-700 dark:bg-primary-500/10 dark:text-primary-300"
                  >
                    {{ t('marketplace.recommended') }}
                  </span>
                </div>
                <p class="mt-1 font-mono text-xs text-gray-500 dark:text-dark-400">{{ product.code }}</p>
                <p v-if="product.description" class="mt-2 text-sm text-gray-600 dark:text-dark-300">{{ product.description }}</p>
                <div v-if="product.tags.length" class="mt-3 flex flex-wrap gap-2">
                  <span
                    v-for="tag in product.tags"
                    :key="`${product.id}-${tag}`"
                    class="rounded-full bg-primary-50 px-2.5 py-1 text-xs text-primary-700 dark:bg-primary-500/10 dark:text-primary-300"
                  >
                    {{ tag }}
                  </span>
                </div>
                <div class="mt-4 space-y-3">
                  <div>
                    <p class="mb-2 text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('marketplace.prices') }}</p>
                    <div v-if="product.prices.length" class="flex flex-wrap gap-2">
                      <span
                        v-for="price in product.prices"
                        :key="price.id"
                        class="rounded-full bg-emerald-50 px-3 py-1.5 text-sm font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300"
                      >
                        {{ formatPrice(price.amount, price.currency) }} / {{ priceTypeLabel(price.price_type) }}
                      </span>
                    </div>
                    <p v-else class="text-sm text-gray-400 dark:text-dark-400">-</p>
                  </div>
                  <div>
                    <p class="mb-2 text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('marketplace.entitlements') }}</p>
                    <div v-if="product.grants.length" class="flex flex-wrap gap-2">
                      <span
                        v-for="grant in product.grants"
                        :key="grant.id"
                        class="rounded-full bg-amber-50 px-3 py-1.5 text-sm text-amber-700 dark:bg-amber-500/10 dark:text-amber-300"
                      >
                        {{ grantLabel(grant) }}
                      </span>
                    </div>
                    <p v-else class="text-sm text-gray-400 dark:text-dark-400">-</p>
                  </div>
                </div>
              </div>

              <button type="button" class="btn btn-secondary btn-sm" @click="openEditProduct(product)">
                <Icon name="edit" size="sm" class="mr-1" />
                {{ t('common.edit') }}
              </button>
            </div>
          </article>
        </div>
        <div v-else class="card p-10">
          <EmptyState
            :title="t('admin.commerce.empty.productsTitle')"
            :description="t('admin.commerce.empty.productsDesc')"
            :action-text="t('admin.commerce.actions.createProduct')"
            @action="openCreateProduct"
          />
        </div>
      </div>

      <div v-else-if="activeTab === 'models'" class="space-y-4">
        <div v-if="models.length" class="grid gap-4 xl:grid-cols-2">
          <article v-for="model in models" :key="model.id" class="card border border-gray-200/80 p-5 dark:border-dark-700">
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ model.display_name }}</h2>
                  <span :class="['badge', statusBadgeClass(model.status)]">{{ statusLabel(model.status) }}</span>
                  <span
                    v-if="model.recommended"
                    class="rounded-full bg-primary-50 px-2.5 py-1 text-xs font-medium text-primary-700 dark:bg-primary-500/10 dark:text-primary-300"
                  >
                    {{ t('marketplace.recommended') }}
                  </span>
                </div>
                <p class="mt-1 font-mono text-xs text-gray-500 dark:text-dark-400">{{ model.model_key }}</p>
                <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ model.vendor || t('marketplace.vendorUnknown') }}</p>
                <p v-if="model.description" class="mt-2 text-sm text-gray-600 dark:text-dark-300">{{ model.description }}</p>
                <div v-if="model.tags.length" class="mt-3 flex flex-wrap gap-2">
                  <span
                    v-for="tag in model.tags"
                    :key="`${model.id}-${tag}`"
                    class="rounded-full bg-primary-50 px-2.5 py-1 text-xs text-primary-700 dark:bg-primary-500/10 dark:text-primary-300"
                  >
                    {{ tag }}
                  </span>
                </div>
                <div class="mt-4">
                  <p class="mb-2 text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ t('admin.commerce.sections.linkedProducts') }}</p>
                  <div v-if="model.offers.length" class="space-y-2">
                    <div
                      v-for="offer in model.offers"
                      :key="offer.binding_id"
                      class="rounded-2xl border border-gray-200/80 p-3 dark:border-dark-700"
                    >
                      <div class="flex flex-wrap items-center gap-2">
                        <span class="font-medium text-gray-900 dark:text-white">{{ offer.product.name }}</span>
                        <span class="rounded-full bg-gray-100 px-2.5 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-dark-300">
                          {{ bindingTypeLabel(offer.binding_type) }}
                        </span>
                      </div>
                    </div>
                  </div>
                  <p v-else class="text-sm text-gray-400 dark:text-dark-400">-</p>
                </div>
              </div>

              <button type="button" class="btn btn-secondary btn-sm" @click="openEditModel(model)">
                <Icon name="edit" size="sm" class="mr-1" />
                {{ t('common.edit') }}
              </button>
            </div>
          </article>
        </div>
        <div v-else class="card p-10">
          <EmptyState
            :title="t('admin.commerce.empty.modelsTitle')"
            :description="t('admin.commerce.empty.modelsDesc')"
            :action-text="t('admin.commerce.actions.createModel')"
            @action="openCreateModel"
          />
        </div>
      </div>

      <div v-else-if="activeTab === 'orders'" class="space-y-4">
        <div class="card p-4">
          <form class="flex flex-col gap-3 lg:flex-row lg:items-end" @submit.prevent="applyOrdersFilter">
            <div class="flex-1">
              <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('orders.orderNo') }}
              </label>
              <input
                v-model="ordersFilterOrderNoInput"
                type="text"
                class="input font-mono text-sm"
                :placeholder="t('admin.commerce.ordersFilterPlaceholder')"
              />
              <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                {{ t('admin.commerce.ordersFilterHint') }}
              </p>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <button type="submit" class="btn btn-primary btn-sm" :disabled="ordersLoading">
                {{ t('common.search') }}
              </button>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :disabled="ordersLoading || !hasOrdersFilter"
                @click="clearOrdersFilter"
              >
                {{ t('common.reset') }}
              </button>
            </div>
          </form>

          <div
            v-if="hasOrdersFilter"
            class="mt-4 rounded-xl border border-primary-100 bg-primary-50 px-4 py-3 text-sm text-primary-700 dark:border-primary-500/20 dark:bg-primary-500/10 dark:text-primary-300"
          >
            {{ t('admin.commerce.ordersFilterActive', { orderNo: normalizedOrdersFilterOrderNo }) }}
          </div>
        </div>

        <div v-if="ordersLoading && orderItems.length === 0" class="card p-10 text-center text-gray-500 dark:text-dark-300">
          <Icon name="refresh" size="lg" class="mx-auto animate-spin" />
          <p class="mt-3">{{ t('common.loading') }}</p>
        </div>
        <div v-else-if="orderItems.length" class="space-y-4">
          <article
            v-for="order in orderItems"
            :key="order.id"
            :class="[
              'card border p-5',
              isHighlightedOrder(order)
                ? 'border-primary-300 bg-primary-50/40 dark:border-primary-500/40 dark:bg-primary-500/5'
                : 'border-gray-200/80 dark:border-dark-700'
            ]"
          >
            <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                    {{ order.snapshot.product.name }}
                  </h2>
                  <span :class="['badge', statusBadgeClass(order.status)]">
                    {{ orderStatusLabel(order.status) }}
                  </span>
                  <span :class="paymentBadgeClass(order.payment_status)">
                    {{ paymentStatusLabel(order.payment_status) }}
                  </span>
                </div>

                <div class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-5">
                  <div>
                    <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                      {{ t('orders.orderNo') }}
                    </p>
                    <p class="mt-1 font-mono text-sm text-gray-900 dark:text-white">
                      {{ order.order_no }}
                    </p>
                  </div>
                  <div>
                    <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                      {{ t('admin.commerce.fields.userId') }}
                    </p>
                    <p class="mt-1 text-sm text-gray-900 dark:text-white">
                      #{{ order.user_id }}
                    </p>
                  </div>
                  <div>
                    <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                      {{ t('orders.price') }}
                    </p>
                    <p class="mt-1 text-sm text-gray-900 dark:text-white">
                      {{ formatPrice(order.amount, order.currency) }}
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

              <div class="flex flex-wrap items-center gap-2">
                <button
                  type="button"
                  class="btn btn-secondary btn-sm"
                  @click="openOrderDetail(order)"
                >
                  {{ t('admin.commerce.actions.viewOrder') }}
                </button>
                <button
                  v-if="canManualComplete(order)"
                  type="button"
                  class="btn btn-primary btn-sm"
                  :disabled="manualCompletingOrderId === order.id"
                  @click="handleManualComplete(order)"
                >
                  {{
                    manualCompletingOrderId === order.id
                      ? t('common.loading')
                      : t('admin.commerce.actions.manualComplete')
                  }}
                </button>
              </div>
            </div>
          </article>

          <div v-if="ordersPagination.pages > 1" class="flex items-center justify-end gap-3">
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="ordersPagination.page <= 1 || ordersLoading"
              @click="changeOrdersPage(ordersPagination.page - 1)"
            >
              {{ t('orders.previous') }}
            </button>
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ ordersPagination.page }} / {{ ordersPagination.pages }}
            </span>
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="ordersPagination.page >= ordersPagination.pages || ordersLoading"
              @click="changeOrdersPage(ordersPagination.page + 1)"
            >
              {{ t('orders.next') }}
            </button>
          </div>
        </div>
        <div v-else class="card p-10">
          <EmptyState
            :title="t('orders.emptyTitle')"
            :description="t('orders.emptyDesc')"
          />
        </div>
      </div>

      <div v-else-if="activeTab === 'wallet'" class="space-y-4">
        <div v-if="walletLoading && walletItems.length === 0" class="card p-10 text-center text-gray-500 dark:text-dark-300">
          <Icon name="refresh" size="lg" class="mx-auto animate-spin" />
          <p class="mt-3">{{ t('common.loading') }}</p>
        </div>
        <div v-else-if="walletItems.length" class="space-y-4">
          <article
            v-for="item in walletItems"
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

                <div class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-5">
                  <div>
                    <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                      {{ t('admin.commerce.fields.userId') }}
                    </p>
                    <p class="mt-1 text-sm text-gray-900 dark:text-white">
                      #{{ item.user_id }}
                    </p>
                  </div>
                  <div>
                    <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                      {{ t('wallet.balanceBefore') }}
                    </p>
                    <p class="mt-1 text-sm text-gray-900 dark:text-white">
                      {{ item.balance_before.toFixed(2) }}
                    </p>
                  </div>
                  <div>
                    <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                      {{ t('wallet.balanceAfter') }}
                    </p>
                    <p class="mt-1 text-sm text-gray-900 dark:text-white">
                      {{ item.balance_after.toFixed(2) }}
                    </p>
                  </div>
                  <div>
                    <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                      {{ t('wallet.reason') }}
                    </p>
                    <p class="mt-1 text-sm text-gray-900 dark:text-white">
                      {{ walletReasonLabel(item.reason_type) }}
                    </p>
                  </div>
                  <div>
                    <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                      {{ t('wallet.createdAt') }}
                    </p>
                    <p class="mt-1 text-sm text-gray-900 dark:text-white">
                      {{ formatDateTime(item.created_at) }}
                    </p>
                  </div>
                </div>

                <div class="mt-4 text-sm text-gray-500 dark:text-dark-400">
                  <span>{{ item.reason_detail }}</span>
                  <span v-if="item.order_no" class="ml-2 font-mono">#{{ item.order_no }}</span>
                </div>
              </div>
            </div>
          </article>

          <div v-if="walletPagination.pages > 1" class="flex items-center justify-end gap-3">
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="walletPagination.page <= 1 || walletLoading"
              @click="changeWalletPage(walletPagination.page - 1)"
            >
              {{ t('wallet.previous') }}
            </button>
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ walletPagination.page }} / {{ walletPagination.pages }}
            </span>
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="walletPagination.page >= walletPagination.pages || walletLoading"
              @click="changeWalletPage(walletPagination.page + 1)"
            >
              {{ t('wallet.next') }}
            </button>
          </div>
        </div>
        <div v-else class="card p-10">
          <EmptyState
            :title="t('wallet.emptyTitle')"
            :description="t('wallet.emptyDesc')"
          />
        </div>
      </div>

      <BaseDialog :show="showProductDialog" :title="productDialogTitle" width="wide" @close="closeProductDialog">
        <form id="commerce-product-form" class="space-y-6" @submit.prevent="submitProduct">
          <section class="space-y-4">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.commerce.sections.basicInfo') }}</h3>
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="input-label">{{ t('admin.commerce.fields.code') }}</label>
                <input v-model.trim="productForm.code" type="text" class="input font-mono" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.commerce.fields.name') }}</label>
                <input v-model.trim="productForm.name" type="text" class="input" />
              </div>
            </div>
            <div>
              <label class="input-label">{{ t('admin.commerce.fields.description') }}</label>
              <textarea v-model.trim="productForm.description" rows="3" class="input"></textarea>
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="input-label">{{ t('admin.commerce.fields.productType') }}</label>
                <Select v-model="productForm.product_type" :options="productTypeOptions" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.commerce.fields.status') }}</label>
                <Select v-model="productForm.status" :options="statusOptions" />
              </div>
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="input-label">{{ t('admin.commerce.fields.coverImage') }}</label>
                <input v-model.trim="productForm.cover_image" type="text" class="input" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.commerce.fields.sortOrder') }}</label>
                <input v-model.number="productForm.sort_order" type="number" min="1" step="1" class="input" />
              </div>
            </div>
            <div>
              <label class="input-label">{{ t('admin.commerce.fields.tags') }}</label>
              <input v-model.trim="productForm.tags_text" type="text" class="input" />
              <p class="input-hint">{{ t('admin.commerce.hints.tags') }}</p>
            </div>
            <div class="flex items-center justify-between rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
              <span class="text-sm font-medium text-gray-900 dark:text-white">{{ t('marketplace.recommended') }}</span>
              <Toggle v-model="productForm.recommended" />
            </div>
          </section>

          <section class="space-y-4">
            <div class="flex items-center justify-between">
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('marketplace.prices') }}</h3>
              <button type="button" class="btn btn-secondary btn-sm" @click="addProductPrice">
                <Icon name="plus" size="sm" class="mr-1" />
                {{ t('admin.commerce.actions.addPrice') }}
              </button>
            </div>
            <div v-if="productForm.prices.length === 0" class="rounded-2xl border border-dashed border-gray-200 p-4 text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
              {{ t('admin.commerce.empty.pricesDesc') }}
            </div>
            <div v-for="(price, index) in productForm.prices" :key="`price-${index}`" class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
              <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                <div>
                  <label class="input-label">{{ t('admin.commerce.fields.priceType') }}</label>
                  <Select v-model="price.price_type" :options="priceTypeOptions" />
                </div>
                <div>
                  <label class="input-label">{{ t('admin.commerce.fields.amount') }}</label>
                  <input v-model.number="price.amount" type="number" min="0" step="0.01" class="input" />
                </div>
                <div>
                  <label class="input-label">{{ t('admin.commerce.fields.currency') }}</label>
                  <input v-model.trim="price.currency" type="text" class="input uppercase" maxlength="8" />
                </div>
                <div>
                  <label class="input-label">{{ t('admin.commerce.fields.originalAmount') }}</label>
                  <input v-model.number="price.original_amount" type="number" min="0" step="0.01" class="input" />
                </div>
                <div>
                  <label class="input-label">{{ t('admin.commerce.fields.sortOrder') }}</label>
                  <input v-model.number="price.sort_order" type="number" min="1" step="1" class="input" />
                </div>
                <div class="flex items-end justify-between gap-3">
                  <div class="flex items-center gap-3 rounded-2xl border border-gray-200 px-4 py-3 dark:border-dark-700">
                    <span class="text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('admin.commerce.fields.enabled') }}</span>
                    <Toggle v-model="price.enabled" />
                  </div>
                  <button type="button" class="btn btn-secondary btn-sm" @click="removeProductPrice(index)">{{ t('common.delete') }}</button>
                </div>
              </div>
            </div>
          </section>

          <section class="space-y-4">
            <div class="flex items-center justify-between">
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('marketplace.entitlements') }}</h3>
              <button type="button" class="btn btn-secondary btn-sm" @click="addProductGrant">
                <Icon name="plus" size="sm" class="mr-1" />
                {{ t('admin.commerce.actions.addGrant') }}
              </button>
            </div>
            <div v-if="productForm.grants.length === 0" class="rounded-2xl border border-dashed border-gray-200 p-4 text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
              {{ t('admin.commerce.empty.grantsDesc') }}
            </div>
            <div v-for="(grant, index) in productForm.grants" :key="`grant-${index}`" class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
              <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <div>
                  <label class="input-label">{{ t('admin.commerce.fields.group') }}</label>
                  <Select v-model="grant.group_id" :options="groupOptions" searchable />
                </div>
                <div>
                  <label class="input-label">{{ t('admin.commerce.fields.grantType') }}</label>
                  <Select v-model="grant.grant_type" :options="grantTypeOptions" />
                </div>
                <div>
                  <label class="input-label">{{ t('admin.commerce.fields.validityDays') }}</label>
                  <input v-model.number="grant.validity_days" type="number" min="0" step="1" class="input" />
                </div>
                <div>
                  <label class="input-label">{{ t('admin.commerce.fields.priority') }}</label>
                  <input v-model.number="grant.priority" type="number" min="0" step="1" class="input" />
                </div>
              </div>
              <div class="mt-4 flex justify-end">
                <button type="button" class="btn btn-secondary btn-sm" @click="removeProductGrant(index)">{{ t('common.delete') }}</button>
              </div>
            </div>
          </section>

          <section class="space-y-3">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.commerce.sections.metadata') }}</h3>
            <textarea v-model="productForm.metadata_text" rows="7" class="input font-mono text-sm"></textarea>
            <p class="input-hint">{{ t('admin.commerce.hints.metadata') }}</p>
          </section>
        </form>
        <template #footer>
          <div class="flex justify-end gap-3">
            <button type="button" class="btn btn-secondary" @click="closeProductDialog">{{ t('common.cancel') }}</button>
            <button type="submit" form="commerce-product-form" class="btn btn-primary" :disabled="submittingProduct">
              {{ submittingProduct ? t('common.saving') : (productForm.id ? t('common.save') : t('common.create')) }}
            </button>
          </div>
        </template>
      </BaseDialog>

      <BaseDialog
        :show="showOrderDetailDialog"
        :title="t('admin.commerce.orderDetailTitle')"
        width="wide"
        @close="closeOrderDetailDialog"
      >
        <div v-if="selectedOrderDetail" class="space-y-6">
          <section class="space-y-4">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.commerce.orderDetailBasicInfo') }}
            </h3>
            <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('orders.orderNo') }}
                </p>
                <p class="mt-1 font-mono text-sm text-gray-900 dark:text-white">
                  {{ selectedOrderDetail.order_no }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('admin.commerce.fields.userId') }}
                </p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  #{{ selectedOrderDetail.user_id }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('admin.commerce.fields.status') }}
                </p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ orderStatusLabel(selectedOrderDetail.status) }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('admin.commerce.orderDetailPaymentStatus') }}
                </p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ paymentStatusLabel(selectedOrderDetail.payment_status) }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('admin.settings.purchase.callbackDebugLookupFieldProvider') }}
                </p>
                <p class="mt-1 font-mono text-sm text-gray-900 dark:text-white">
                  {{ selectedOrderDetail.payment_provider || '-' }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('orders.price') }}
                </p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ formatPrice(selectedOrderDetail.amount, selectedOrderDetail.currency) }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('admin.commerce.orderDetailPaidAt') }}
                </p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ selectedOrderDetail.paid_at ? formatDateTime(selectedOrderDetail.paid_at) : '-' }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('admin.commerce.orderDetailExpiredAt') }}
                </p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ selectedOrderDetail.expired_at ? formatDateTime(selectedOrderDetail.expired_at) : '-' }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('orders.createdAt') }}
                </p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ formatDateTime(selectedOrderDetail.created_at) }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('orders.completedAt') }}
                </p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ selectedOrderDetail.completed_at ? formatDateTime(selectedOrderDetail.completed_at) : '-' }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('admin.commerce.orderDetailUpdatedAt') }}
                </p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ formatDateTime(selectedOrderDetail.updated_at) }}
                </p>
              </div>
            </div>
          </section>

          <section class="space-y-4">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.commerce.orderDetailSnapshot') }}
            </h3>
            <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('admin.commerce.orderDetailProductCode') }}
                </p>
                <p class="mt-1 font-mono text-sm text-gray-900 dark:text-white">
                  {{ selectedOrderDetail.snapshot.product.code }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('admin.commerce.fields.productType') }}
                </p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ productTypeLabel(selectedOrderDetail.snapshot.product.product_type) }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('admin.commerce.orderDetailPriceType') }}
                </p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ priceTypeLabel(selectedOrderDetail.snapshot.price.price_type) }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('admin.commerce.orderDetailOriginalAmount') }}
                </p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{
                    selectedOrderDetail.snapshot.price.original_amount
                      ? formatPrice(selectedOrderDetail.snapshot.price.original_amount, selectedOrderDetail.snapshot.price.currency)
                      : '-'
                  }}
                </p>
              </div>
            </div>

            <div v-if="selectedOrderDetail.snapshot.product.description" class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
              <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                {{ t('admin.commerce.fields.description') }}
              </p>
              <p class="mt-2 text-sm text-gray-700 dark:text-dark-300">
                {{ selectedOrderDetail.snapshot.product.description }}
              </p>
            </div>

            <div v-if="selectedOrderDetail.snapshot.product.tags.length" class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
              <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                {{ t('admin.commerce.fields.tags') }}
              </p>
              <div class="mt-2 flex flex-wrap gap-2">
                <span
                  v-for="tag in selectedOrderDetail.snapshot.product.tags"
                  :key="`detail-${selectedOrderDetail.id}-${tag}`"
                  class="rounded-full bg-primary-50 px-2.5 py-1 text-xs text-primary-700 dark:bg-primary-500/10 dark:text-primary-300"
                >
                  {{ tag }}
                </span>
              </div>
            </div>

            <div class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
              <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                {{ t('orders.entitlements') }}
              </p>
              <div v-if="selectedOrderDetail.snapshot.grants.length" class="mt-2 flex flex-wrap gap-2">
                <span
                  v-for="grant in selectedOrderDetail.snapshot.grants"
                  :key="`detail-grant-${selectedOrderDetail.id}-${grant.id}-${grant.group_id}`"
                  class="rounded-full bg-amber-50 px-3 py-1.5 text-sm text-amber-700 dark:bg-amber-500/10 dark:text-amber-300"
                >
                  {{ grantLabel(grant) }}
                </span>
              </div>
              <p v-else class="mt-2 text-sm text-gray-500 dark:text-dark-400">-</p>
            </div>

            <div class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
              <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                {{ t('admin.commerce.orderDetailMetadata') }}
              </p>
              <pre class="mt-2 overflow-x-auto rounded-lg bg-gray-50 px-3 py-3 text-xs text-gray-700 dark:bg-dark-800 dark:text-gray-200"><code>{{ selectedOrderDetailMetadata }}</code></pre>
            </div>
          </section>

          <section class="space-y-4">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.commerce.orderDetailPaymentAction') }}
            </h3>
            <div
              v-if="selectedOrderDetail.payment_action"
              class="grid gap-4 md:grid-cols-2"
            >
              <div class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('admin.settings.purchase.callbackDebugLookupFieldProvider') }}
                </p>
                <p class="mt-2 text-sm text-gray-900 dark:text-white">
                  {{ selectedOrderDetail.payment_action.provider_name }} ({{ selectedOrderDetail.payment_action.provider }})
                </p>
              </div>
              <div class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
                <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                  {{ t('admin.commerce.orderDetailCheckoutUrl') }}
                </p>
                <code class="mt-2 block break-all rounded-lg bg-gray-50 px-3 py-2 text-xs text-gray-700 dark:bg-dark-800 dark:text-gray-200">
                  {{ selectedOrderDetail.payment_action.checkout_url }}
                </code>
              </div>
            </div>
            <div
              v-else
              class="rounded-2xl border border-dashed border-gray-300 p-4 text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400"
            >
              {{ t('admin.commerce.orderDetailPaymentActionEmpty') }}
            </div>
          </section>

          <section class="space-y-4">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.commerce.orderDetailTransaction') }}
            </h3>
            <div
              v-if="selectedOrderDetail.payment_transaction"
              class="space-y-4"
            >
              <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <div class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('admin.commerce.orderDetailTransactionStatus') }}
                  </p>
                  <p class="mt-2 text-sm text-gray-900 dark:text-white">
                    {{ selectedOrderDetail.payment_transaction.status }}
                  </p>
                </div>
                <div class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('admin.commerce.orderDetailProviderTradeNo') }}
                  </p>
                  <p class="mt-2 font-mono text-sm text-gray-900 dark:text-white">
                    {{ selectedOrderDetail.payment_transaction.provider_trade_no || '-' }}
                  </p>
                </div>
                <div class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('admin.commerce.orderDetailPaidAmount') }}
                  </p>
                  <p class="mt-2 text-sm text-gray-900 dark:text-white">
                    {{ selectedOrderDetail.payment_transaction.paid_amount }} {{ selectedOrderDetail.payment_transaction.paid_currency }}
                  </p>
                </div>
                <div class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('admin.commerce.orderDetailUpdatedAt') }}
                  </p>
                  <p class="mt-2 text-sm text-gray-900 dark:text-white">
                    {{ formatDateTime(selectedOrderDetail.payment_transaction.updated_at) }}
                  </p>
                </div>
              </div>

              <div class="grid gap-4 xl:grid-cols-2">
                <div class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('admin.commerce.orderDetailRequestPayload') }}
                  </p>
                  <pre class="mt-2 overflow-x-auto rounded-lg bg-gray-50 px-3 py-3 text-xs text-gray-700 dark:bg-dark-800 dark:text-gray-200"><code>{{ selectedOrderDetailRequestPayload }}</code></pre>
                </div>
                <div class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
                  <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">
                    {{ t('admin.commerce.orderDetailCallbackPayload') }}
                  </p>
                  <pre class="mt-2 overflow-x-auto rounded-lg bg-gray-50 px-3 py-3 text-xs text-gray-700 dark:bg-dark-800 dark:text-gray-200"><code>{{ selectedOrderDetailCallbackPayload }}</code></pre>
                </div>
              </div>
            </div>
            <div
              v-else
              class="rounded-2xl border border-dashed border-gray-300 p-4 text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400"
            >
              {{ t('admin.commerce.orderDetailTransactionEmpty') }}
            </div>
          </section>
        </div>
        <template #footer>
          <div class="flex justify-end">
            <button type="button" class="btn btn-secondary" @click="closeOrderDetailDialog">
              {{ t('common.close') }}
            </button>
          </div>
        </template>
      </BaseDialog>

      <BaseDialog :show="showModelDialog" :title="modelDialogTitle" width="wide" @close="closeModelDialog">
        <form id="commerce-model-form" class="space-y-6" @submit.prevent="submitModel">
          <section class="space-y-4">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.commerce.sections.basicInfo') }}</h3>
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="input-label">{{ t('admin.commerce.fields.modelKey') }}</label>
                <input v-model.trim="modelForm.model_key" type="text" class="input font-mono" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.commerce.fields.displayName') }}</label>
                <input v-model.trim="modelForm.display_name" type="text" class="input" />
              </div>
            </div>
            <div>
              <label class="input-label">{{ t('admin.commerce.fields.description') }}</label>
              <textarea v-model.trim="modelForm.description" rows="3" class="input"></textarea>
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="input-label">{{ t('marketplace.vendor') }}</label>
                <input v-model.trim="modelForm.vendor" type="text" class="input" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.commerce.fields.status') }}</label>
                <Select v-model="modelForm.status" :options="statusOptions" />
              </div>
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <label class="input-label">{{ t('admin.commerce.fields.icon') }}</label>
                <input v-model.trim="modelForm.icon" type="text" class="input" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.commerce.fields.sortOrder') }}</label>
                <input v-model.number="modelForm.sort_order" type="number" min="1" step="1" class="input" />
              </div>
            </div>
            <div>
              <label class="input-label">{{ t('admin.commerce.fields.tags') }}</label>
              <input v-model.trim="modelForm.tags_text" type="text" class="input" />
              <p class="input-hint">{{ t('admin.commerce.hints.tags') }}</p>
            </div>
            <div class="flex items-center justify-between rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
              <span class="text-sm font-medium text-gray-900 dark:text-white">{{ t('marketplace.recommended') }}</span>
              <Toggle v-model="modelForm.recommended" />
            </div>
          </section>

          <section class="space-y-4">
            <div class="flex items-center justify-between">
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.commerce.sections.linkedProducts') }}</h3>
              <button type="button" class="btn btn-secondary btn-sm" @click="addModelBinding">
                <Icon name="plus" size="sm" class="mr-1" />
                {{ t('admin.commerce.actions.addBinding') }}
              </button>
            </div>
            <div v-if="modelForm.bindings.length === 0" class="rounded-2xl border border-dashed border-gray-200 p-4 text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
              {{ t('admin.commerce.empty.bindingsDesc') }}
            </div>
            <div v-for="(binding, index) in modelForm.bindings" :key="`binding-${index}`" class="rounded-2xl border border-gray-200 p-4 dark:border-dark-700">
              <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                <div class="xl:col-span-2">
                  <label class="input-label">{{ t('admin.commerce.fields.product') }}</label>
                  <Select v-model="binding.product_id" :options="productOptions" searchable />
                  <p v-if="productOptions.length === 0" class="input-hint">{{ t('admin.commerce.hints.noProductsForBinding') }}</p>
                </div>
                <div>
                  <label class="input-label">{{ t('admin.commerce.fields.bindingType') }}</label>
                  <Select v-model="binding.binding_type" :options="bindingTypeOptions" />
                </div>
              </div>
              <div class="mt-4 flex justify-end">
                <button type="button" class="btn btn-secondary btn-sm" @click="removeModelBinding(index)">{{ t('common.delete') }}</button>
              </div>
            </div>
          </section>

          <section class="space-y-3">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.commerce.sections.metadata') }}</h3>
            <textarea v-model="modelForm.metadata_text" rows="7" class="input font-mono text-sm"></textarea>
            <p class="input-hint">{{ t('admin.commerce.hints.metadata') }}</p>
          </section>
        </form>
        <template #footer>
          <div class="flex justify-end gap-3">
            <button type="button" class="btn btn-secondary" @click="closeModelDialog">{{ t('common.cancel') }}</button>
            <button type="submit" form="commerce-model-form" class="btn btn-primary" :disabled="submittingModel">
              {{ submittingModel ? t('common.saving') : (modelForm.id ? t('common.save') : t('common.create')) }}
            </button>
          </div>
        </template>
      </BaseDialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores'
import type {
  AdminGroup,
  CommerceCatalog,
  CommerceCatalogGrant,
  CommerceCatalogModel,
  CommerceOrder,
  CommerceCatalogProduct,
  CommerceWalletLedger,
  CommerceModelUpsertRequest,
  CommerceProductUpsertRequest,
  PaginatedResponse,
  SelectOption
} from '@/types'

type CommerceTab = 'products' | 'models' | 'orders' | 'wallet'

interface ProductPriceFormItem {
  price_type: string
  amount: number | null | string
  currency: string
  original_amount: number | null | string
  enabled: boolean
  sort_order: number | null | string
}

interface ProductGrantFormItem {
  group_id: number | null | string
  grant_type: string
  validity_days: number | null | string
  priority: number | null | string
}

interface ProductFormState {
  id: number | null
  code: string
  name: string
  description: string
  product_type: string
  status: string
  cover_image: string
  tags_text: string
  sort_order: number | null | string
  recommended: boolean
  metadata_text: string
  prices: ProductPriceFormItem[]
  grants: ProductGrantFormItem[]
}

interface ModelBindingFormItem {
  product_id: number | null | string
  binding_type: string
}

interface ModelFormState {
  id: number | null
  model_key: string
  display_name: string
  description: string
  vendor: string
  icon: string
  tags_text: string
  status: string
  sort_order: number | null | string
  recommended: boolean
  metadata_text: string
  bindings: ModelBindingFormItem[]
}

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const activeTab = ref<CommerceTab>('products')
const loading = ref(false)
const loadError = ref('')
const catalog = ref<CommerceCatalog | null>(null)
const groups = ref<AdminGroup[]>([])
const ordersLoading = ref(false)
const walletLoading = ref(false)
const ordersError = ref('')
const walletError = ref('')
const manualCompletingOrderId = ref<number | null>(null)
const ordersFilterOrderNoInput = ref('')
const showOrderDetailDialog = ref(false)
const selectedOrderDetail = ref<CommerceOrder | null>(null)
const ordersPagination = ref<PaginatedResponse<CommerceOrder>>({
  items: [],
  total: 0,
  page: 1,
  page_size: 10,
  pages: 1
})
const walletPagination = ref<PaginatedResponse<CommerceWalletLedger>>({
  items: [],
  total: 0,
  page: 1,
  page_size: 10,
  pages: 1
})

const showProductDialog = ref(false)
const showModelDialog = ref(false)
const submittingProduct = ref(false)
const submittingModel = ref(false)

const productForm = ref<ProductFormState>(createEmptyProductForm())
const modelForm = ref<ModelFormState>(createEmptyModelForm())

const products = computed(() => [...(catalog.value?.products ?? [])].sort((left, right) => (left.sort_order - right.sort_order) || (left.id - right.id)))
const models = computed(() => [...(catalog.value?.models ?? [])].sort((left, right) => (left.sort_order - right.sort_order) || (left.id - right.id)))
const recommendedCount = computed(() => products.value.filter(item => item.recommended).length + models.value.filter(item => item.recommended).length)
const orderItems = computed(() => ordersPagination.value.items)
const walletItems = computed(() => walletPagination.value.items)
const normalizedOrdersFilterOrderNo = computed(() => ordersFilterOrderNoInput.value.trim())
const hasOrdersFilter = computed(() => normalizedOrdersFilterOrderNo.value.length > 0)
const selectedOrderDetailMetadata = computed(() =>
  formatMetadata(selectedOrderDetail.value?.snapshot.product.metadata ?? {})
)
const selectedOrderDetailRequestPayload = computed(() =>
  formatStructuredText(selectedOrderDetail.value?.payment_transaction?.request_payload ?? '')
)
const selectedOrderDetailCallbackPayload = computed(() =>
  formatStructuredText(selectedOrderDetail.value?.payment_transaction?.callback_payload ?? '')
)
const activeLoading = computed(() => {
  switch (activeTab.value) {
    case 'orders':
      return ordersLoading.value
    case 'wallet':
      return walletLoading.value
    default:
      return loading.value
  }
})
const activeErrorMessage = computed(() => {
  switch (activeTab.value) {
    case 'orders':
      return ordersError.value
    case 'wallet':
      return walletError.value
    default:
      return loadError.value
  }
})

const statusOptions = computed<SelectOption[]>(() => [
  { value: 'draft', label: t('admin.commerce.statuses.draft') },
  { value: 'active', label: t('admin.commerce.statuses.active') },
  { value: 'disabled', label: t('admin.commerce.statuses.disabled') }
])

const productTypeOptions = computed<SelectOption[]>(() => [
  { value: 'topup_balance', label: t('marketplace.productTypes.topupBalance') },
  { value: 'subscription_group', label: t('marketplace.productTypes.subscriptionGroup') },
  { value: 'standard_group_access', label: t('marketplace.productTypes.standardGroupAccess') },
  { value: 'combo', label: t('marketplace.productTypes.combo') }
])

const priceTypeOptions = computed<SelectOption[]>(() => [
  { value: 'one_time', label: t('marketplace.priceTypes.oneTime') },
  { value: 'monthly', label: t('marketplace.priceTypes.monthly') },
  { value: 'quarterly', label: t('marketplace.priceTypes.quarterly') },
  { value: 'yearly', label: t('marketplace.priceTypes.yearly') }
])

const grantTypeOptions = computed<SelectOption[]>(() => [
  { value: 'subscription', label: t('marketplace.grantTypes.subscription') },
  { value: 'allowed_group', label: t('marketplace.grantTypes.allowedGroup') }
])

const bindingTypeOptions = computed<SelectOption[]>(() => [
  { value: 'primary', label: t('marketplace.bindingTypes.primary') },
  { value: 'upsell', label: t('marketplace.bindingTypes.upsell') },
  { value: 'topup', label: t('marketplace.bindingTypes.topup') }
])

const groupOptions = computed<SelectOption[]>(() => {
  const optionMap = new Map<number, string>()
  for (const group of groups.value) {
    optionMap.set(group.id, group.name)
  }
  for (const product of products.value) {
    for (const grant of product.grants) {
      if (!optionMap.has(grant.group_id)) {
        optionMap.set(grant.group_id, grant.group_name || `#${grant.group_id}`)
      }
    }
  }
  return [...optionMap.entries()].sort((left, right) => left[1].localeCompare(right[1])).map(([value, label]) => ({ value, label }))
})

const productOptions = computed<SelectOption[]>(() =>
  products.value.map(product => ({ value: product.id, label: `${product.name} (${product.code})` }))
)

const productDialogTitle = computed(() =>
  productForm.value.id ? t('admin.commerce.actions.editProduct') : t('admin.commerce.actions.createProduct')
)

const modelDialogTitle = computed(() =>
  modelForm.value.id ? t('admin.commerce.actions.editModel') : t('admin.commerce.actions.createModel')
)

onMounted(() => {
  applyInitialRouteState()
  void loadCatalog()
  if (activeTab.value === 'orders') {
    void loadOrdersPage(1)
  }
  if (activeTab.value === 'wallet') {
    void loadWalletPage(1)
  }
})

function normalizeRouteQueryValue(value: unknown): string {
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

function parseRouteTab(value: string): CommerceTab {
  if (value === 'products' || value === 'models' || value === 'orders' || value === 'wallet') {
    return value
  }
  return 'products'
}

function applyInitialRouteState(): void {
  activeTab.value = parseRouteTab(normalizeRouteQueryValue(route.query.tab))
  ordersFilterOrderNoInput.value = normalizeRouteQueryValue(route.query.order_no).trim()
}

async function syncCommerceRouteQuery(tab: CommerceTab = activeTab.value): Promise<void> {
  const nextQuery = { ...route.query }

  if (tab !== 'products') {
    nextQuery.tab = tab
  } else {
    delete nextQuery.tab
  }
  if (tab === 'orders' && normalizedOrdersFilterOrderNo.value) {
    nextQuery.order_no = normalizedOrdersFilterOrderNo.value
  } else {
    delete nextQuery.order_no
  }

  await router.replace({
    path: '/admin/commerce',
    query: nextQuery
  })
}

async function loadCatalog(): Promise<void> {
  loading.value = true
  loadError.value = ''
  try {
    const [catalogResponse, groupResponse] = await Promise.all([
      adminAPI.commerce.getCatalog(),
      adminAPI.groups.getAll()
    ])
    catalog.value = catalogResponse
    groups.value = groupResponse
  } catch (error: any) {
    loadError.value = error?.message || t('admin.commerce.messages.loadFailed')
    appStore.showError(loadError.value)
  } finally {
    loading.value = false
  }
}

function switchTab(tab: CommerceTab): void {
  activeTab.value = tab
  void syncCommerceRouteQuery(tab)
  if (tab === 'orders' && orderItems.value.length === 0 && !ordersLoading.value) {
    void loadOrdersPage(1)
  }
  if (tab === 'wallet' && walletItems.value.length === 0 && !walletLoading.value) {
    void loadWalletPage(1)
  }
}

async function refreshActiveTab(): Promise<void> {
  switch (activeTab.value) {
    case 'orders':
      await loadOrdersPage(ordersPagination.value.page || 1)
      break
    case 'wallet':
      await loadWalletPage(walletPagination.value.page || 1)
      break
    default:
      await loadCatalog()
      break
  }
}

async function loadOrdersPage(page: number): Promise<void> {
  ordersLoading.value = true
  ordersError.value = ''
  try {
    ordersPagination.value = await adminAPI.commerce.listOrders({
      page,
      page_size: ordersPagination.value.page_size || 10,
      order_no: normalizedOrdersFilterOrderNo.value || undefined
    })
  } catch (error: any) {
    ordersError.value = error?.response?.data?.detail || error?.message || t('orders.loadFailed')
    appStore.showError(ordersError.value)
  } finally {
    ordersLoading.value = false
  }
}

async function changeOrdersPage(page: number): Promise<void> {
  if (page < 1 || page > ordersPagination.value.pages) {
    return
  }
  await loadOrdersPage(page)
}

async function applyOrdersFilter(): Promise<void> {
  await syncCommerceRouteQuery('orders')
  await loadOrdersPage(1)
}

async function clearOrdersFilter(): Promise<void> {
  ordersFilterOrderNoInput.value = ''
  await syncCommerceRouteQuery('orders')
  await loadOrdersPage(1)
}

function openOrderDetail(order: CommerceOrder): void {
  selectedOrderDetail.value = order
  showOrderDetailDialog.value = true
}

function closeOrderDetailDialog(): void {
  showOrderDetailDialog.value = false
  selectedOrderDetail.value = null
}

async function loadWalletPage(page: number): Promise<void> {
  walletLoading.value = true
  walletError.value = ''
  try {
    walletPagination.value = await adminAPI.commerce.getWalletLedger({
      page,
      page_size: walletPagination.value.page_size || 10
    })
  } catch (error: any) {
    walletError.value = error?.response?.data?.detail || error?.message || t('wallet.loadFailed')
    appStore.showError(walletError.value)
  } finally {
    walletLoading.value = false
  }
}

async function changeWalletPage(page: number): Promise<void> {
  if (page < 1 || page > walletPagination.value.pages) {
    return
  }
  await loadWalletPage(page)
}

function canManualComplete(order: CommerceOrder): boolean {
  return order.status === 'pending' || order.status === 'paid'
}

async function handleManualComplete(order: CommerceOrder): Promise<void> {
  if (!canManualComplete(order)) {
    return
  }
  if (!window.confirm(t('admin.commerce.messages.manualCompleteConfirm', { orderNo: order.order_no }))) {
    return
  }

  manualCompletingOrderId.value = order.id
  try {
    await adminAPI.commerce.manualCompleteOrder(order.id, {})
    appStore.showSuccess(t('admin.commerce.messages.orderCompleted'))
    await loadOrdersPage(ordersPagination.value.page || 1)
    if (walletItems.value.length > 0) {
      await loadWalletPage(walletPagination.value.page || 1)
    }
  } catch (error: any) {
    appStore.showError(error?.response?.data?.detail || error?.message || t('common.error'))
  } finally {
    manualCompletingOrderId.value = null
  }
}

function isHighlightedOrder(order: CommerceOrder): boolean {
  return hasOrdersFilter.value && order.order_no === normalizedOrdersFilterOrderNo.value
}

function tabClass(tab: CommerceTab): string[] {
  return [
    'rounded-xl px-4 py-2 text-sm font-medium transition-colors',
    activeTab.value === tab
      ? 'bg-white text-primary-600 shadow-sm dark:bg-dark-700 dark:text-primary-300'
      : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'
  ]
}

function openCreateProduct(): void {
  productForm.value = createEmptyProductForm()
  showProductDialog.value = true
}

function openEditProduct(product: CommerceCatalogProduct): void {
  productForm.value = {
    id: product.id,
    code: product.code,
    name: product.name,
    description: product.description,
    product_type: product.product_type,
    status: product.status,
    cover_image: product.cover_image,
    tags_text: product.tags.join(', '),
    sort_order: product.sort_order,
    recommended: product.recommended,
    metadata_text: formatMetadata(product.metadata),
    prices: product.prices.map(item => ({
      price_type: item.price_type,
      amount: item.amount,
      currency: item.currency,
      original_amount: item.original_amount,
      enabled: item.enabled,
      sort_order: item.sort_order
    })),
    grants: product.grants.map(item => ({
      group_id: item.group_id,
      grant_type: item.grant_type,
      validity_days: item.validity_days,
      priority: item.priority
    }))
  }
  showProductDialog.value = true
}

function closeProductDialog(): void {
  showProductDialog.value = false
  productForm.value = createEmptyProductForm()
}

function openCreateModel(): void {
  modelForm.value = createEmptyModelForm()
  showModelDialog.value = true
}

function openEditModel(model: CommerceCatalogModel): void {
  modelForm.value = {
    id: model.id,
    model_key: model.model_key,
    display_name: model.display_name,
    description: model.description,
    vendor: model.vendor,
    icon: model.icon,
    tags_text: model.tags.join(', '),
    status: model.status,
    sort_order: model.sort_order,
    recommended: model.recommended,
    metadata_text: formatMetadata(model.metadata),
    bindings: model.offers.map(offer => ({
      product_id: offer.product.id,
      binding_type: offer.binding_type
    }))
  }
  showModelDialog.value = true
}

function closeModelDialog(): void {
  showModelDialog.value = false
  modelForm.value = createEmptyModelForm()
}

function addProductPrice(): void {
  productForm.value.prices.push(createEmptyProductPrice())
}

function removeProductPrice(index: number): void {
  productForm.value.prices.splice(index, 1)
}

function addProductGrant(): void {
  productForm.value.grants.push(createEmptyProductGrant())
}

function removeProductGrant(index: number): void {
  productForm.value.grants.splice(index, 1)
}

function addModelBinding(): void {
  modelForm.value.bindings.push(createEmptyModelBinding())
}

function removeModelBinding(index: number): void {
  modelForm.value.bindings.splice(index, 1)
}

async function submitProduct(): Promise<void> {
  submittingProduct.value = true
  try {
    const payload = buildProductPayload()
    if (productForm.value.id) {
      await adminAPI.commerce.updateProduct(productForm.value.id, payload)
      appStore.showSuccess(t('admin.commerce.messages.productUpdated'))
    } else {
      await adminAPI.commerce.createProduct(payload)
      appStore.showSuccess(t('admin.commerce.messages.productCreated'))
    }
    closeProductDialog()
    await loadCatalog()
  } catch (error: any) {
    appStore.showError(error?.message || error?.response?.data?.detail || t('common.error'))
  } finally {
    submittingProduct.value = false
  }
}

async function submitModel(): Promise<void> {
  submittingModel.value = true
  try {
    const payload = buildModelPayload()
    if (modelForm.value.id) {
      await adminAPI.commerce.updateModel(modelForm.value.id, payload)
      appStore.showSuccess(t('admin.commerce.messages.modelUpdated'))
    } else {
      await adminAPI.commerce.createModel(payload)
      appStore.showSuccess(t('admin.commerce.messages.modelCreated'))
    }
    closeModelDialog()
    await loadCatalog()
  } catch (error: any) {
    appStore.showError(error?.message || error?.response?.data?.detail || t('common.error'))
  } finally {
    submittingModel.value = false
  }
}

function buildProductPayload(): CommerceProductUpsertRequest {
  const code = productForm.value.code.trim()
  const name = productForm.value.name.trim()
  if (!code) {
    throw new Error(t('admin.commerce.validation.productCodeRequired'))
  }
  if (!name) {
    throw new Error(t('admin.commerce.validation.productNameRequired'))
  }

  return {
    code,
    name,
    description: productForm.value.description.trim(),
    product_type: productForm.value.product_type,
    status: productForm.value.status,
    cover_image: productForm.value.cover_image.trim(),
    tags: parseTags(productForm.value.tags_text),
    sort_order: toNumberOrDefault(productForm.value.sort_order, 100),
    recommended: productForm.value.recommended,
    metadata: parseMetadata(productForm.value.metadata_text),
    prices: productForm.value.prices.map((item, index) => {
      const amount = toNullableNumber(item.amount)
      const currency = item.currency.trim().toUpperCase()
      if (!item.price_type.trim()) {
        throw new Error(t('admin.commerce.validation.priceTypeRequired', { index: index + 1 }))
      }
      if (amount === null) {
        throw new Error(t('admin.commerce.validation.priceAmountRequired', { index: index + 1 }))
      }
      if (!currency) {
        throw new Error(t('admin.commerce.validation.currencyRequired', { index: index + 1 }))
      }
      return {
        price_type: item.price_type.trim(),
        amount,
        currency,
        original_amount: toNullableNumber(item.original_amount),
        enabled: item.enabled,
        sort_order: toNumberOrDefault(item.sort_order, 100)
      }
    }),
    group_bindings: productForm.value.grants.map((item, index) => {
      const groupId = toNullableNumber(item.group_id)
      if (groupId === null || groupId <= 0) {
        throw new Error(t('admin.commerce.validation.groupRequired', { index: index + 1 }))
      }
      return {
        group_id: groupId,
        grant_type: item.grant_type.trim(),
        validity_days: toNumberOrDefault(item.validity_days, 0),
        priority: toNumberOrDefault(item.priority, 50)
      }
    })
  }
}

function buildModelPayload(): CommerceModelUpsertRequest {
  const modelKey = modelForm.value.model_key.trim()
  const displayName = modelForm.value.display_name.trim()
  if (!modelKey) {
    throw new Error(t('admin.commerce.validation.modelKeyRequired'))
  }
  if (!displayName) {
    throw new Error(t('admin.commerce.validation.displayNameRequired'))
  }

  return {
    model_key: modelKey,
    display_name: displayName,
    description: modelForm.value.description.trim(),
    vendor: modelForm.value.vendor.trim(),
    icon: modelForm.value.icon.trim(),
    tags: parseTags(modelForm.value.tags_text),
    status: modelForm.value.status,
    sort_order: toNumberOrDefault(modelForm.value.sort_order, 100),
    recommended: modelForm.value.recommended,
    metadata: parseMetadata(modelForm.value.metadata_text),
    product_bindings: modelForm.value.bindings.map((item, index) => {
      const productId = toNullableNumber(item.product_id)
      if (productId === null || productId <= 0) {
        throw new Error(t('admin.commerce.validation.productRequired', { index: index + 1 }))
      }
      return {
        product_id: productId,
        binding_type: item.binding_type.trim()
      }
    })
  }
}

function createEmptyProductForm(): ProductFormState {
  return {
    id: null,
    code: '',
    name: '',
    description: '',
    product_type: 'subscription_group',
    status: 'active',
    cover_image: '',
    tags_text: '',
    sort_order: 100,
    recommended: false,
    metadata_text: '{}',
    prices: [createEmptyProductPrice()],
    grants: []
  }
}

function createEmptyProductPrice(): ProductPriceFormItem {
  return {
    price_type: 'monthly',
    amount: null,
    currency: 'CNY',
    original_amount: null,
    enabled: true,
    sort_order: 100
  }
}

function createEmptyProductGrant(): ProductGrantFormItem {
  return {
    group_id: null,
    grant_type: 'subscription',
    validity_days: 30,
    priority: 50
  }
}

function createEmptyModelForm(): ModelFormState {
  return {
    id: null,
    model_key: '',
    display_name: '',
    description: '',
    vendor: '',
    icon: '',
    tags_text: '',
    status: 'active',
    sort_order: 100,
    recommended: false,
    metadata_text: '{}',
    bindings: []
  }
}

function createEmptyModelBinding(): ModelBindingFormItem {
  return {
    product_id: null,
    binding_type: 'primary'
  }
}

function parseTags(input: string): string[] {
  const result: string[] = []
  const seen = new Set<string>()
  for (const rawItem of input.split(/[\n,，]/)) {
    const item = rawItem.trim()
    if (!item || seen.has(item)) {
      continue
    }
    seen.add(item)
    result.push(item)
  }
  return result
}

function parseMetadata(input: string): Record<string, unknown> {
  const content = input.trim()
  if (!content) {
    return {}
  }
  try {
    const value = JSON.parse(content) as unknown
    if (!value || Array.isArray(value) || typeof value !== 'object') {
      throw new Error(t('admin.commerce.validation.metadataObjectOnly'))
    }
    return value as Record<string, unknown>
  } catch (error) {
    if (error instanceof Error && error.message === t('admin.commerce.validation.metadataObjectOnly')) {
      throw error
    }
    throw new Error(t('admin.commerce.validation.metadataInvalid'))
  }
}

function formatMetadata(metadata: Record<string, unknown>): string {
  return JSON.stringify(metadata ?? {}, null, 2)
}

function formatStructuredText(value: string): string {
  const content = value.trim()
  if (!content) {
    return '-'
  }

  try {
    return JSON.stringify(JSON.parse(content), null, 2)
  } catch {
    return content
  }
}

function toNullableNumber(value: unknown): number | null {
  if (value === null || value === undefined || value === '') {
    return null
  }
  const numeric = Number(value)
  return Number.isFinite(numeric) ? numeric : null
}

function toNumberOrDefault(value: unknown, defaultValue: number): number {
  const numeric = toNullableNumber(value)
  return numeric === null ? defaultValue : numeric
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

function walletDirectionLabel(direction: string): string {
  return direction === 'debit' ? t('wallet.directions.debit') : t('wallet.directions.credit')
}

function walletReasonLabel(reasonType: string): string {
  if (reasonType === 'commerce_order_topup') {
    return t('wallet.reasonTypes.orderTopup')
  }
  return reasonType
}

function signedAmount(direction: string, amount: number): string {
  const prefix = direction === 'debit' ? '-' : '+'
  return `${prefix}${amount.toFixed(2)}`
}

function directionBadgeClass(direction: string): string {
  return direction === 'debit'
    ? 'rounded-full bg-red-50 px-2.5 py-1 text-xs font-medium text-red-700 dark:bg-red-500/10 dark:text-red-300'
    : 'rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'
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

function statusLabel(status: string): string {
  switch (status) {
    case 'draft':
      return t('admin.commerce.statuses.draft')
    case 'active':
      return t('admin.commerce.statuses.active')
    case 'disabled':
      return t('admin.commerce.statuses.disabled')
    default:
      return status
  }
}

function statusBadgeClass(status: string): string {
  switch (status) {
    case 'active':
      return 'badge-success'
    case 'disabled':
      return 'badge-danger'
    case 'draft':
      return 'badge-warning'
    default:
      return 'badge-gray'
  }
}
</script>
