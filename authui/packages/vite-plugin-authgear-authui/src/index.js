import htmlParser from "node-html-parser";
import path from "path";
import fs from "fs/promises";

const templateNameByAssetName = {
  "build.html": "__generated_asset.html",
  "build-authflowv2.html": "authflowv2/__generated_asset.html",
};

/**
 * @param {string} s
 */
function removeLeadingSlash(s) {
  return s.replace(/^\/+/, "");
}

/**
 * @param {string} filePath
 * @return {string}
 */
function nameWithoutHash(filePath) {
  const textArray = filePath.split(".");
  if (textArray.length === 3) {
    return textArray[0] + "." + textArray[2];
  }
  return filePath;
}

/**
 * @param {Record<string, string>} attributes
 * @returns {string}
 */
function stringifyHTMLAttributes(attributes) {
  const attributesString = Object.entries(attributes).map(
    ([key, value]) => `${key}="${value}"`
  );
  return attributesString.join(" ");
}

/**
 * @param {{type: "css" | "js" | "modulepreload", name: string, attributes: Record<string, string>}[]} elements
 * @returns {string}
 */
function elementsToHTMLString(elements) {
  const textArray = [];
  for (const element of elements) {
    if (element.type === "css") {
      const htmlLine = `<link rel="stylesheet" href="{{ call $.GeneratedStaticAssetURL "${element.name}" }}">`;
      if (element.name === "tailwind-dark-theme.css") {
        textArray.push(`{{ if $.DarkThemeEnabled }}`);
        textArray.push(htmlLine);
        textArray.push(`{{ end }}`);
      } else {
        textArray.push(htmlLine);
      }
    }
    if (element.type === "modulepreload") {
      const htmlLine = `<link rel="modulepreload" href="{{ call $.GeneratedStaticAssetURL "${element.name}" }}">`;
      textArray.push(htmlLine);
    }
    if (element.type === "js") {
      const attributes = Object.fromEntries(
        Object.entries(element.attributes).filter(
          ([key]) => !["src", "crossorigin"].includes(key)
        )
      );
      const attributesString = stringifyHTMLAttributes(attributes);
      const htmlLine = `<script ${attributesString} nonce="{{ $.CSPNonce }}" src="{{ call $.GeneratedStaticAssetURL "${element.name}" }}"></script>`;
      textArray.push(htmlLine);
    }
  }
  return textArray.join("\n");
}

/**
 * @param {string} targetPath
 * @param {Record<string, string>} manifest
 */
async function writeManifest(targetPath, manifest) {
  await fs.writeFile(targetPath, JSON.stringify(manifest));
}

/**
 * @param {string} targetPath
 * @param {string} templateName
 * @param {{type: "css" | "js" | "modulepreload", name: string, attributes: Record<string, string>}[]} elements
 */
async function writeHTMLTemplate(targetPath, templateName, elements) {
  const tpl = [
    `{{ define "${templateName}" }}`,
    elementsToHTMLString(elements),
    "{{ end }}",
  ].join("\n");
  await fs.writeFile(targetPath, tpl);
}

// TODO: Support hot reload without reloading webpage
/** @returns {import("vite").Plugin} */
function servePlugin({ input }) {
  let config;

  return {
    name: "vite-plugin-authgear-authui:serve",
    apply: "serve",

    configResolved(_config) {
      config = _config;
    },
  };
}

/** @returns {import('vite').Plugin} */
function buildPlugin({ input }) {
  const manifest = {};
  let config;

  return {
    name: "vite-plugin-authgear-authui:build",
    apply: "build",

    configResolved(_config) {
      config = _config;
    },

    async writeBundle(_options, bundles) {
      for (const [bundleName, bundleInfo] of Object.entries(bundles)) {
        const assetName = nameWithoutHash(bundleName);
        const assetBaseName = path.basename(assetName);
        if (Object.keys(templateNameByAssetName).includes(assetBaseName)) {
          const elements = [];
          const root = htmlParser.parse(bundleInfo.source);
          const head = root.getElementsByTagName("head")[0];

          for (const node of head.childNodes) {
            if (node.tagName === "LINK") {
              const hashedName = removeLeadingSlash(node.getAttribute("href"));
              const key = nameWithoutHash(hashedName);
              if (node.getAttribute("rel") === "stylesheet") {
                elements.push({
                  type: "css",
                  name: key,
                });
              }
              if (node.getAttribute("rel") === "modulepreload") {
                elements.push({
                  type: "modulepreload",
                  name: key,
                });
              }
            }
            if (node.tagName === "SCRIPT") {
              let hashedName = removeLeadingSlash(node.getAttribute("src"));
              // When we want to keep the type of script to "classic" instead of "module",
              // Vite will not modify the script tag line and keep using the ".ts" for the "src" attribute.
              // ref https://github.com/vitejs/vite/blob/7d24b5f56697f6ec6e6facbe8601d3f993b764c8/packages/vite/src/node/plugins/html.ts#L450
              hashedName = hashedName.replace(".ts", ".js");
              const key = nameWithoutHash(hashedName);
              const attributes = node.attributes;
              elements.push({
                type: "js",
                attributes: attributes,
                name: key,
              });
            }
          }

          // Generate html template(s)
          /** @type {string} */
          const templateName = templateNameByAssetName[assetBaseName];
          const targetHTMLTemplatePath = `../resources/authgear/templates/en/web/${templateName}`;
          await writeHTMLTemplate(
            targetHTMLTemplatePath,
            templateName,
            elements
          );
          config.logger.info(
            `📄 Wrote bundle HTML to: ${targetHTMLTemplatePath}`
          );
        }
        manifest[assetBaseName] = bundleName;
      }

      // Generate manifest file
      const targetManifestPath =
        "../resources/authgear/generated/manifest.json";
      await writeManifest(targetManifestPath, manifest);
      config.logger.info(`📄 Wrote bundle manifest to: ${targetManifestPath}`);
    },

    config(_config) {
      return {
        build: {
          watch: _config.build.watch && {
            include: "src/**",
          },
          emptyOutDir: true,
          sourcemap: true,
          cssCodeSplit: true,
          // Avoid image assets being inlined into css files
          assetsInlineLimit: 0,
          assetsDir: "",
          rollupOptions: {
            input: input,
            output: {
              format: "module",
              manualChunks: (id) => {
                if (id.includes("node_modules")) {
                  return "vendor";
                }
                // To avoid css files being bundled in 1 file
                if (id.endsWith(".css")) {
                  return path.basename(id);
                }
                return null;
              },
              hashCharacters: "hex",
              assetFileNames: () => "[name].[hash][extname]",
              chunkFileNames: () => "[name].[hash].js",
              entryFileNames: () => "[name].[hash].js",
              sourcemapFileNames: () => "[name].[chunkhash].map.js",
            },
          },
        },
      };
    },
  };
}

function viteAuthgearAuthUI(options) {
  return [servePlugin(options), buildPlugin(options)];
}
export { viteAuthgearAuthUI };
