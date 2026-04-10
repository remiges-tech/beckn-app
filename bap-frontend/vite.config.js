import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    // Dev-server proxy: forwards /api/* to the BAP application.
    // In production (Docker) nginx.conf handles the same routing.
    proxy: {
      '/api': {
        target: 'http://localhost:8083',
        changeOrigin: true,
      }
    }
  }
})
