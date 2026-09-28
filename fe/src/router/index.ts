import { createRouter, createWebHistory } from 'vue-router'
import { useAdminStore } from '@/stores/admin'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
    {
      path: '/p/:slug',
      name: 'product',
      component: () => import('@/views/ProductView.vue'),
    },
    {
      path: '/admin/login',
      name: 'admin-login',
      component: () => import('@/views/AdminLoginView.vue'),
    },
    {
      path: '/admin/produtos',
      name: 'admin-products',
      component: () => import('@/views/AdminProductsView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin',
      redirect: { name: 'admin-products' },
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: { name: 'home' },
    },
  ],
})

router.beforeEach((to) => {
  const store = useAdminStore()
  if (to.meta.requiresAuth && !store.isAuthenticated) {
    return { name: 'admin-login' }
  }
  if (to.name === 'admin-login' && store.isAuthenticated) {
    return { name: 'admin-products' }
  }
  return true
})

export default router
