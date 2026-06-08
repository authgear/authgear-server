import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react";
import { parse } from "node-html-parser";
import crypto from "crypto";
import fs from "fs/promises";
import path from "path";

// Adds CSP nonces to all inline scripts/styles in index.html.
const noncePlugin: Plugin = {
  name: "vite-plugin-authgear-portal:nonce",
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

      const head = html.querySelector("head")!;
      for (const e of elements) {
        if (e.getAttribute("data-order") === "last") {
          head.removeChild(e);
          head.appendChild(e);
        }
      }

      return html.toString();
    },
  },
};

// Adds SRI integrity attributes and an import map to index.html.
// Must run in writeBundle (after files are written) because Vite applies
// additional transforms to some chunks after transformIndexHtml runs,
// so chunk.code there does not match the final file on disk.
const sriPlugin: Plugin = {
  name: "vite-plugin-authgear-portal:sri",
  apply: "build",
  async writeBundle(options, bundle) {
    const outDir = options.dir!;

    // Compute hashes from the final bundle content (matches files on disk).
    const sriMap: Record<string, string> = {};
    const importIntegrity: Record<string, string> = {};
    for (const [name, chunk] of Object.entries(bundle)) {
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
        if (chunk.type === "chunk") {
          importIntegrity["/" + name] = `sha384-${hash}`;
        }
      }
    }

    // Read index.html (already written by Vite, with nonces from noncePlugin).
    const indexHtmlPath = path.join(outDir, "index.html");
    const source = await fs.readFile(indexHtmlPath, "utf-8");
    const html = parse(source);
    const head = html.querySelector("head")!;

    // Inject import map as the first child of <head> so it precedes all
    // <script type="module"> tags. Covers dynamically imported chunks
    // (lazy-loaded routes/components) not listed in the HTML directly.
    // Browser support: Chrome 126+, Firefox 127+, Safari 17.4+.
    head.insertAdjacentHTML(
      "afterbegin",
      `<script type="importmap" nonce="{{ $.CSPNonce }}">${JSON.stringify({ integrity: importIntegrity })}</script>`
    );

    // Add integrity + crossorigin to statically referenced scripts and styles.
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

    await fs.writeFile(indexHtmlPath, html.toString());
  },
};

function viteAuthgearPortal() {
  return [noncePlugin, sriPlugin];
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
