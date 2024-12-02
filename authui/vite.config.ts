import { defineConfig } from "vite";
import { viteAuthgearAuthUI } from "vite-plugin-authgear-authui";

export default defineConfig(() => ({
  plugins: [
    viteAuthgearAuthUI({
      input: {
        v1: "./src/build.html",
        v2: "./src/build-authflowv2.html",
        colorscheme: "./src/colorscheme.ts",
      },
    }),
  ],
  experimental: {
    renderBuiltUrl: () => ({ relative: true }),
  },
  server: {
    host: "0.0.0.0",
    port: 5173,
  },
}));
