import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { viteSingleFile } from "vite-plugin-singlefile";
import { parse } from "node-html-parser";

function nonce() {
  return {
    name: "vite-plugin-authgear-portal-graphiql:build",
    apply: "build" as const,
    transformIndexHtml: {
      order: "post" as const,
      handler: (htmlString: string) => {
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
}

export default defineConfig(() => ({
  plugins: [react(), viteSingleFile(), nonce()],
  root: "./src",
  // We do not configure `server` as this project is not supposed to be run in dev mode.
  build: {
    // outDir is relative to root.
    outDir: "../dist",
    emptyOutDir: true,
    // We do not need sourcemap because we do not support serving them.
    sourcemap: false,
    // We do not need CSS code split because we generate an inlined HTML file.
    cssCodeSplit: false,
    // We do not need manifest because we generate an inlined HTML file.
    manifest: false,
  },
}));
