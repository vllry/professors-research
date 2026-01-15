import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  server: {
    host: '0.0.0.0',
    // Allow requests from Docker container hostnames
    allowedHosts: ['web', 'localhost', '127.0.0.1'],
    proxy: {
      '/api': {
        target: process.env.VITE_API_URL || `http://${process.env.API_HOST || 'api'}:${process.env.PORT || '8080'}`,
        changeOrigin: true,
      },
    },
  },
})

