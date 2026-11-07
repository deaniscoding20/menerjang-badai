import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
export default defineConfig({
  plugins: [react()], base: '/',
  build: { assetsDir: 'assets', sourcemap: false },
  server: { proxy: { '/api': 'http://127.0.0.1:8080', '/auth': 'http://127.0.0.1:8080' } },
  test: { environment: 'jsdom', setupFiles: ['./src/test/setup.ts'], include: ['src/**/*.test.{ts,tsx}'], restoreMocks: true },
});
