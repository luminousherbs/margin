// @ts-check
import { defineConfig } from "astro/config";
import react from "@astrojs/react";
import tailwind from "@astrojs/tailwind";
import node from "@astrojs/node";
import { fileURLToPath } from "url";
import { dirname, resolve } from "path";

const __dirname = dirname(fileURLToPath(import.meta.url));

const API_PORT = process.env.API_PORT || 8081;

// https://astro.build/config
export default defineConfig({
  output: "server",
  adapter: node({ mode: "standalone" }),
  integrations: [react(), tailwind()],
  security: { checkOrigin: false },
  prefetch: {
    prefetchAll: true,
    defaultStrategy: "viewport",
  },
  vite: {
    resolve: {
      alias: {
        "@": resolve(__dirname, "src"),
      },
    },
    ssr: {
      external: ["@resvg/resvg-js"],
    },
    build: {
      commonjsOptions: {
        transformMixedEsModules: true,
      },
      chunkSizeWarningLimit: 1000,
    },
    server: {
      proxy: {
        "/api": {
          target: `http://127.0.0.1:${API_PORT}`,
          changeOrigin: true,
        },
        "/auth": {
          target: `http://127.0.0.1:${API_PORT}`,
          changeOrigin: true,
        },
      },
    },
  },
});
