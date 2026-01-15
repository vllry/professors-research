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
        // Default to localhost for non-Docker dev/test. Docker Compose sets VITE_API_URL and/or API_HOST.
        target:
          process.env.VITE_API_URL ||
          `http://${process.env.API_HOST || 'localhost'}:${process.env.PORT || '8080'}`,
        changeOrigin: true,
      },
    },
  },
})

