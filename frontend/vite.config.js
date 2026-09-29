import adapter from '@sveltejs/adapter-auto';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [sveltekit()],
  server: {
    port: 5173,
    proxy: {
      '/api':   { target: 'http://localhost:8080', changeOrigin: true },
      '/cover': { target: 'http://localhost:8080', changeOrigin: true },
      '/book':  { target: 'http://localhost:8080', changeOrigin: true }
    }
  }
});
