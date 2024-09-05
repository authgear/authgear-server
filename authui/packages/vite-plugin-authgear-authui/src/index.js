import htmlParser from "node-html-parser";
import path from "path";
import fs from "fs/promises";

const templateBase = "../resources/authgear/templates/en/web/";
const templateNameByAssetName = {
  "build.html": "__generated_asset.html",
  "build-authflowv2.html": "authflowv2/__generated_asset.html",
};

const devServerBase = "/_vite";

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
 * @param {string} elementsString
 */
async function writeHTMLTemplate(targetPath, templateName, elementsString) {
  const tpl = [
    `{{ define "${templateName}" }}`,
    elementsString,
    "{{ end }}",
  ].join("\n");
  await fs.writeFile(targetPath, tpl);
}

/** @returns {import("vite").Plugin} */
function servePlugin({ input }) {
  let config;

  /**
   * @param {string} entrtPoint
   * @param {string} moduleName
   * @returns {string}
   */
  function rewriteModuleNameByEntryPoint(entrtPoint, moduleName) {
    // Do not rewrite the module name if it is a URL
    if (moduleName.includes("//")) {
      return moduleName;
    }
    const entryPointDir = path.dirname(path.join(config.base, entrtPoint));
    const newModuleName = path.join(entryPointDir, moduleName);
    return newModuleName;
  }

  /**
   * @param {string} filePath
   */
  async function buildHTMLTemplateFromEntryPoint(filePath) {
    const templateName = templateNameByAssetName[path.basename(filePath)];
    if (templateName == null) {
      config.logger.error(`No template name found for ${filePath}`);
      return;
    }

    const source = await fs.readFile(filePath);
    const root = htmlParser.parse(source);
    const head = root.getElementsByTagName("head")[0];

    const nodes = [];
    for (const node of head.childNodes) {
      if (node.attributes == null) {
        continue;
      }

      // Rewrite node attributes by condition
      // If it is <script>, only rewrite the "src"
      // If it is <link>, only rewrite the "href"
      for (const [key, value] of Object.entries(node.attributes)) {
        if (
          (node.tagName === "SCRIPT" && key === "src") ||
          (node.tagName === "LINK" &&
            node.getAttribute("rel") === "stylesheet" &&
            key === "href")
        ) {
          const moduleName = rewriteModuleNameByEntryPoint(filePath, value);
          node.setAttribute(key, moduleName);
          node.setAttribute("nonce", "{{ $.CSPNonce }}");
        }
      }

      nodes.push(node);
    }

    const elementsStringList = nodes.map((node) => node.toString());

    // Inject vite client for HMR
    // ref https://vitejs.dev/guide/backend-integration.html
    const viteClientSrc = path.join(config.base, "/@vite/client");
    const elementsString = [
      `<script type="module" nonce="{{ $.CSPNonce }}" src="${viteClientSrc}"></script>`,
      ...elementsStringList,
    ].join("\n");

    const targetHTMLTemplatePath = path.join(templateBase, templateName);
    await writeHTMLTemplate(
      targetHTMLTemplatePath,
      templateName,
      elementsString
    );
    config.logger.info(`📄 Wrote HTML to: ${targetHTMLTemplatePath}`);
  }

  /**
   * @param {string} filePath
   */
  async function buildHTMLTemplateIfNeeded(filePath) {
    const relativeFilePath = path.relative(config.root, filePath);
    if (path.extname(relativeFilePath) === ".html") {
      await buildHTMLTemplateFromEntryPoint(relativeFilePath);
    }
  }

  return {
    name: "vite-plugin-authgear-authui:serve",
    apply: "serve",

    configResolved(_config) {
      config = _config;
    },

    config() {
      return {
        base: devServerBase,
      };
    },

    async handleHotUpdate({ file }) {
      await buildHTMLTemplateIfNeeded(file);
    },

    async buildStart(_options) {
      const entryPoints = Object.values(input);
      for (const entryPoint of entryPoints) {
        await buildHTMLTemplateIfNeeded(entryPoint);
      }
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
          const targetHTMLTemplatePath = path.join(templateBase, templateName);
          await writeHTMLTemplate(
            targetHTMLTemplatePath,
            templateName,
            elementsToHTMLString(elements)
          );
          config.logger.info(
            `📄 Wrote bundle HTML to: ${targetHTMLTemplatePath}`
          );
        }
        manifest[assetBaseName] = bundleName;
      }

      // Generate manifest file
      const targetManifestPath = path.join(
        path.relative(config.root, config.build.outDir),
        "/manifest.json"
      );
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
            // Workaround for building bundles with non-deterministic filenames
            // Active issue: https://github.com/vitejs/vite/issues/13672
            // Workaround from https://github.com/vitejs/vite/issues/10506#issuecomment-1367718113
            maxParallelFileOps: 1,
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
