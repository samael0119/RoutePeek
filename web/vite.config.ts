import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  base: '/static/dist/',
  plugins: [vue()],
  build: {
    outDir: '../cmd/server/static/dist',
    emptyOutDir: true
  },
  test: {
    environment: 'node'
  }
});
