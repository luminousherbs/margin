// @ts-check
import { defineConfig } from "astro/config";
import react from "@astrojs/react";
import tailwind from "@astrojs/tailwind";
import node from "@astrojs/node";

// https://astro.build/config
export default defineConfig({
  adapter: node({ mode: "standalone" }),
  integrations: [react(), tailwind()],
  vite: {
    ssr: {
      noExternal: true,
      external: ["@resvg/resvg-js"],
    },
    build: {
      chunkSizeWarningLimit: 1000,
    },
    server: {
      proxy: {
        "/api": {
          target: "http://localhost:8080",
          changeOrigin: true,
        },
        "/auth": {
          target: "http://localhost:8080",
          changeOrigin: true,
        },
      },
    },
  },
});
