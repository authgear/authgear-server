import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react";
import { parse } from "node-html-parser";
import crypto from "crypto";

const plugin: Plugin = {
  name: "vite-plugin-authgear-portal:build",
  apply: "build",
  transformIndexHtml: {
    order: "post",
    handler: (htmlString, ctx) => {
      // Build SRI hash map from bundle: "/path/to/asset" -> "sha384-<base64>"
      // ctx.bundle is available for order:"post" handlers during build.
      const sriMap: Record<string, string> = {};
      if (ctx.bundle != null) {
        for (const [name, chunk] of Object.entries(ctx.bundle)) {
          let content: string | Uint8Array | undefined;
          if (chunk.type === "chunk") {
            content = chunk.code;
          } else if (chunk.type === "asset") {
            content = chunk.source as string | Uint8Array;
          }
          if (content != null) {
            const bytes =
              typeof content === "string" ? Buffer.from(content) : content;
            const hash = crypto
              .createHash("sha384")
              .update(bytes)
              .digest("base64");
            sriMap["/" + name] = `sha384-${hash}`;
          }
        }
      }

      const html = parse(htmlString);

      const scripts = html.querySelectorAll("script");
      const styles = html.querySelectorAll("style");
      const stylesheetLinks = html.querySelectorAll("link[rel=stylesheet]");

      const elements = [...scripts, ...styles, ...stylesheetLinks];
      for (const e of elements) {
        e.setAttribute("nonce", "{{ $.CSPNonce }}");
      }

      const head = html.querySelector("head")!;
      for (const e of elements) {
        if (e.getAttribute("data-order") === "last") {
          head.removeChild(e);
          head.appendChild(e);
        }
      }

      for (const el of html.querySelectorAll("script[src]")) {
        const src = el.getAttribute("src");
        if (src != null && sriMap[src] != null) {
          el.setAttribute("integrity", sriMap[src]);
          el.setAttribute("crossorigin", "anonymous");
        }
      }
      for (const el of html.querySelectorAll("link[rel=stylesheet]")) {
        const href = el.getAttribute("href");
        if (href != null && sriMap[href] != null) {
          el.setAttribute("integrity", sriMap[href]);
          el.setAttribute("crossorigin", "anonymous");
        }
      }
      for (const el of html.querySelectorAll("link[rel=modulepreload]")) {
        const href = el.getAttribute("href");
        if (href != null && sriMap[href] != null) {
          el.setAttribute("integrity", sriMap[href]);
          el.setAttribute("crossorigin", "anonymous");
        }
      }

      return html.toString();
    },
  },
};

function viteAuthgearPortal() {
  return [plugin];
}

export default defineConfig({
  plugins: [react(), viteAuthgearPortal()],
  // The index.html is under the "./src" directory
  root: "./src",
  server: {
    host: "0.0.0.0",
    port: 1234,
    hmr: {
      port: 51234,
    },
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
});
