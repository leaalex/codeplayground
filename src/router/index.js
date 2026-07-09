import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import LoginView from '../views/LoginView.vue'
import FilesView from '../views/FilesView.vue'
import CreateFileView from '../views/CreateFileView.vue'
import PlaygroundView from '../views/PlaygroundView.vue'
import UsersView from '../views/UsersView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
    { path: '/', redirect: '/files' },
    { path: '/files', name: 'files', component: FilesView },
    { path: '/files/new', name: 'createFile', component: CreateFileView },
    { path: '/playground/:id', name: 'playground', component: PlaygroundView },
    { path: '/users', name: 'users', component: UsersView, meta: { adminOnly: true } },
  ],
})

function safeRedirectPath(raw) {
  if (typeof raw !== 'string' || !raw.startsWith('/') || raw.startsWith('//')) {
    return '/files'
  }
  if (raw === '/login' || raw.startsWith('/login?')) {
    return '/files'
  }
  return raw
}

router.beforeEach((to, _from, next) => {
  const { isAuthenticated, isAdmin } = useAuth()
  if (!to.meta.public && !isAuthenticated.value) {
    next({ name: 'login', query: { redirect: to.fullPath } })
  } else if (to.name === 'login' && isAuthenticated.value) {
    next(safeRedirectPath(to.query.redirect))
  } else if (to.meta.adminOnly && !isAdmin.value) {
    next({ name: 'files' })
  } else {
    next()
  }
})

export default router
