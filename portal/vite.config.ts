import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { viteAuthgearPortal } from "vite-plugin-authgear-portal";

export default defineConfig(() => ({
  plugins: [react(), viteAuthgearPortal()],
  // The index.html is under the "./src" directory
  root: "./src",
  server: {
    host: "0.0.0.0",
    port: 1234,
  },
  build: {
    outDir: "../dist",
    emptyOutDir: true,
    sourcemap: true,
    cssCodeSplit: true,
    manifest: true,
    // Avoid image assets being inlined into css files
    assetsInlineLimit: 0,
    assetsDir: "shared-assets",
    rollupOptions: {
      // Workaround for building bundles with non-deterministic filenames
      // Active issue: https://github.com/vitejs/vite/issues/13672
      // Workaround from https://github.com/vitejs/vite/issues/10506#issuecomment-1367718113
      maxParallelFileOps: 1,
    },
  },
}));
