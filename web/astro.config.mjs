// @ts-check
import { defineConfig } from "astro/config";
import react from "@astrojs/react";
import tailwind from "@astrojs/tailwind";
import node from "@astrojs/node";

const API_PORT = process.env.API_PORT || 8081;

const isDev = process.env.NODE_ENV === "development";

// https://astro.build/config
export default defineConfig({
  adapter: node({ mode: "standalone" }),
  integrations: [react(), tailwind()],
  security: {
    checkOrigin: false,
  },
  vite: {
    ssr: {
      noExternal: isDev ? /^(?!react|react-dom|react-router-dom|cookie)/ : true,
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
