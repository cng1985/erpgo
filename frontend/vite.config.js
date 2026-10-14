import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { readdir, rm } from 'node:fs/promises'
import { join } from 'node:path'

function cleanEmbeddedOutput() {
  let outputDir
  return {
    name: 'clean-embedded-output',
    apply: 'build',
    configResolved(config) {
      outputDir = config.build.outDir
    },
    async buildStart() {
      let entries
      try {
        entries = await readdir(outputDir)
      } catch (error) {
        if (error.code === 'ENOENT') return
        throw error
      }
      await Promise.all(
        entries
          .filter((entry) => entry !== '.keep')
          .map((entry) => rm(join(outputDir, entry), { recursive: true, force: true })),
      )
    },
  }
}

export default defineConfig({
  plugins: [vue(), cleanEmbeddedOutput()],
  build: {
    outDir: '../backend/internal/web/dist',
    emptyOutDir: false,
  },
  server: {
    proxy: {
      '/api': process.env.VITE_API_TARGET || 'http://127.0.0.1:8080',
    },
  },
})
