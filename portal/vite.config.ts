import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react";
import { parse } from "node-html-parser";

const plugin: Plugin = {
  name: "vite-plugin-authgear-portal:build",
  apply: "build",
  transformIndexHtml: {
    order: "post",
    handler: (htmlString) => {
      const html = parse(htmlString);

      const scripts = html.querySelectorAll("script");
      const styles = html.querySelectorAll("style");
      const stylesheetLinks = html.querySelectorAll("link[rel=stylesheet]");

      const elements = [...scripts, ...styles, ...stylesheetLinks];
      for (const e of elements) {
        e.setAttribute("nonce", "{{ $.CSPNonce }}");
      }

      return html.toString();
    },
  },
};

function viteAuthgearPortal() {
  return [plugin];
}

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
