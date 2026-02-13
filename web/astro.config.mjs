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
    build: {
      chunkSizeWarningLimit: 1000,
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (id.includes("node_modules")) {
              if (
                id.includes("react") ||
                id.includes("react-dom") ||
                id.includes("react-router-dom")
              ) {
                return "vendor-react";
              }
              if (id.includes("lucide-react")) {
                return "vendor-lucide";
              }
              if (id.includes("emoji-picker-react")) {
                return "vendor-emoji-picker";
              }
              if (id.includes("date-fns")) {
                return "vendor-date-fns";
              }
              return "vendor";
            }
          },
        },
      },
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
