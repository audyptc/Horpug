import { defineConfig, loadEnv } from 'vite'
import react, { reactCompilerPreset } from '@vitejs/plugin-react'
import babel from '@rolldown/plugin-babel'
import tailwindcss from '@tailwindcss/vite'
import path from 'node:path'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  // Empty VITE_API_URL (the production default, see frontend/Dockerfile) means
  // "same origin" — nginx proxies /api to the backend there. In dev there's no
  // nginx, so we proxy /api and /health to the backend ourselves and keep the
  // frontend code calling the same relative paths in both environments.
  const apiProxyTarget = env.VITE_API_URL || 'http://localhost:3009'

  return {
    resolve: {
      alias: {
        '@': path.resolve(import.meta.dirname, './src'),
      },
    },
    plugins: [
      tailwindcss(),
      react(),
      babel({ presets: [reactCompilerPreset()] })
    ],
    server: {
      proxy: {
        '/api': { target: apiProxyTarget, changeOrigin: true },
        '/health': { target: apiProxyTarget, changeOrigin: true },
      },
    },
  }
})
