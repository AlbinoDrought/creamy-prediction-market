import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { guest: true },
    },
    {
      path: '/home',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
      meta: { requiresAuth: true, playerOnly: true },
    },
    {
      path: '/predictions/:id',
      name: 'prediction',
      component: () => import('@/views/PredictionView.vue'),
      meta: { requiresAuth: true, playerOnly: true },
    },
    {
      path: '/leaderboard',
      name: 'leaderboard',
      component: () => import('@/views/LeaderboardView.vue'),
      meta: { requiresAuth: true, playerOnly: true },
    },
    {
      path: '/my-bets',
      name: 'my-bets',
      component: () => import('@/views/MyBetsView.vue'),
      meta: { requiresAuth: true, playerOnly: true },
    },
    {
      path: '/admin',
      component: () => import('@/views/admin/AdminLayout.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
      children: [
        {
          path: '',
          name: 'admin-dashboard',
          component: () => import('@/views/admin/DashboardView.vue'),
        },
        {
          path: 'predictions/new',
          name: 'admin-prediction-new',
          component: () => import('@/views/admin/PredictionForm.vue'),
        },
        {
          path: 'predictions/:id',
          name: 'admin-prediction-edit',
          component: () => import('@/views/admin/PredictionForm.vue'),
        },
        {
          path: 'users',
          name: 'admin-users',
          component: () => import('@/views/admin/UsersView.vue'),
        },
      ],
    },
  ],
})

router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()

  // Try to restore session if we have a token but no user
  if (authStore.token && !authStore.user) {
    await authStore.fetchUser()
  }

  const isAuthenticated = authStore.isAuthenticated
  const isAdmin = authStore.isAdmin

  // Guest-only routes (login page)
  if (to.meta.guest && isAuthenticated) {
    // Redirect admins to admin dashboard, players to home
    return next({ name: isAdmin ? 'admin-dashboard' : 'home' })
  }

  // Protected routes
  if (to.meta.requiresAuth && !isAuthenticated) {
    return next({ name: 'login' })
  }

  // Admin routes - only admins can access
  if (to.meta.requiresAdmin && !isAdmin) {
    return next({ name: 'home' })
  }

  // Player-only routes - admins cannot access
  if (to.meta.playerOnly && isAdmin) {
    return next({ name: 'admin-dashboard' })
  }

  next()
})

export default router
