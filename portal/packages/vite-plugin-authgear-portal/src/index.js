import { parse } from "node-html-parser";

const plugin = {
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

export { viteAuthgearPortal };
