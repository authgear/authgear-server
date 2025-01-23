import { defineConfig, type Plugin, type ResolvedConfig } from "vite";

import htmlParser, { HTMLElement } from "node-html-parser";
import path from "path";
import fs from "fs/promises";

const fontFileExtensions = [".eot", ".otf", ".ttf", "woff", ".woff2"];

const templateBase = "../resources/authgear/templates/en/web/";
const templateNameByAssetName: Record<string, string> = {
  "build.html": "__generated_asset.html",
  "build-authflowv2.html": "authflowv2/__generated_asset.html",
};

const devServerBase = "/_vite";

function nameWithoutHash(filePath: string): string {
  const originalFilePath = filePath;

  const exts = [];

  while (true) {
    const ext = path.extname(filePath);
    if (ext === "") {
      break;
    }

    filePath = filePath.slice(0, -ext.length);
    exts.push(ext);
  }
  // [.map, .js] => [.js, .map]
  exts.reverse();

  // filePath is now ended with the hash, or it is not an asset at all.
  // If it is not an asset, for example, index.html, just return the originalFilePath.
  const match = /-[-_A-Za-z0-9]{8}$/.exec(filePath);
  if (match == null) {
    return originalFilePath;
  }

  const withoutHash = filePath.slice(0, match.index);
  const ext = exts.join("");

  return withoutHash + ext;
}

function stringifyHTMLAttributes(attributes: Record<string, string>): string {
  const attributesString = Object.entries(attributes).map(
    ([key, value]) => `${key}="${value}"`
  );
  return attributesString.join(" ");
}

type OurHTMLElement =
  | OurHTMLElementCSS
  | OurHTMLElementJS
  | OurHTMLElementFontPreload
  | OurHTMLElementModulePreload;

interface OurHTMLElementCSS {
  type: "css";
  name: string;
}

interface OurHTMLElementJS {
  type: "js";
  name: string;
  attributes: Record<string, string>;
}

interface OurHTMLElementFontPreload {
  type: "modulepreload";
  name: string;
}

interface OurHTMLElementModulePreload {
  type: "fontpreload";
  name: string;
}

function elementsToHTMLString(elements: OurHTMLElement[]): string {
  // We want turbo to perform full reload if any script or css files changed.
  // Add data-turbo-track="reload" to all css and script elements.
  const textArray = [];
  for (const element of elements) {
    if (element.type === "css") {
      const htmlLine = `<link nonce="{{ $.CSPNonce }}" rel="stylesheet" href="{{ call $.GeneratedStaticAssetURL "${element.name}" }}" data-turbo-track="reload">`;
      if (element.name === "tailwind-dark-theme.css") {
        textArray.push(`{{ if $.DarkThemeEnabled }}`);
        textArray.push(htmlLine);
        textArray.push(`{{ end }}`);
      } else {
        textArray.push(htmlLine);
      }
    }
    if (element.type === "fontpreload") {
      const htmlLine = `<link rel="preload" as="font" crossorigin="anonymous" href="{{ call $.GeneratedStaticAssetURL "${element.name}" }}">`;
      textArray.push(htmlLine);
    }
    if (element.type === "modulepreload") {
      const htmlLine = `<link rel="modulepreload" href="{{ call $.GeneratedStaticAssetURL "${element.name}" }}" data-turbo-track="reload">`;
      textArray.push(htmlLine);
    }
    if (element.type === "js") {
      const attributes = Object.fromEntries(
        Object.entries(element.attributes).filter(
          ([key]) => !["src", "crossorigin"].includes(key)
        )
      );
      const attributesString = stringifyHTMLAttributes(attributes);
      const htmlLine = `<script ${attributesString} nonce="{{ $.CSPNonce }}" src="{{ call $.GeneratedStaticAssetURL "${element.name}" }}" data-turbo-track="reload"></script>`;
      textArray.push(htmlLine);
    }
  }
  return textArray.join("\n");
}

async function writeManifest(
  targetPath: string,
  manifest: Record<string, string>
) {
  await fs.writeFile(targetPath, JSON.stringify(manifest));
}

async function writeHTMLTemplate(
  targetPath: string,
  templateName: string,
  elementsString: string
) {
  const tpl = [
    `{{ define "${templateName}" }}`,
    elementsString,
    "{{ end }}",
  ].join("\n");
  await fs.writeFile(targetPath, tpl);
}

interface AuthgearAuthUIPluginOptions {
  input: {
    v1: string;
    v2: string;
    colorscheme: string;
  };
}

function servePlugin({ input }: AuthgearAuthUIPluginOptions): Plugin {
  let config: ResolvedConfig;

  function rewriteModuleNameByEntryPoint(
    entrtPoint: string,
    moduleName: string
  ): string {
    // Do not rewrite the module name if it is a URL
    if (moduleName.includes("//")) {
      return moduleName;
    }
    const entryPointDir = path.dirname(path.join(config.base, entrtPoint));
    const newModuleName = path.join(entryPointDir, moduleName);
    return newModuleName;
  }

  async function buildHTMLTemplateFromEntryPoint(filePath: string) {
    const templateName = templateNameByAssetName[path.basename(filePath)];
    if (templateName == null) {
      config.logger.error(`No template name found for ${filePath}`);
      return;
    }

    const source = await fs.readFile(filePath, { encoding: "utf8" });
    const root = htmlParser.parse(source);
    const head = root.getElementsByTagName("head")[0];

    const nodes: HTMLElement[] = [];
    for (const node of head.childNodes) {
      if (node instanceof HTMLElement) {
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

  async function buildHTMLTemplateIfNeeded(filePath: string): Promise<void> {
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

function buildPlugin({ input }: AuthgearAuthUIPluginOptions): Plugin {
  const manifest: Record<string, string> = {};
  let config: ResolvedConfig;

  return {
    name: "vite-plugin-authgear-authui:build",
    apply: "build",

    configResolved(_config) {
      config = _config;
    },

    async writeBundle(_options, bundles) {
      // Handle font preload
      const fontPreloadElements: OurHTMLElement[] = [];
      for (const [bundleName, _bundleInfo] of Object.entries(bundles)) {
        const assetName = nameWithoutHash(bundleName);
        const assetBaseName = path.basename(assetName);
        const assetExt = path.extname(assetBaseName);
        if (fontFileExtensions.includes(assetExt)) {
          fontPreloadElements.push({
            type: "fontpreload",
            name: assetBaseName,
          });
        }
      }

      // Handle other bundles
      for (const [bundleName, bundleInfo] of Object.entries(bundles)) {
        const assetName = nameWithoutHash(bundleName);
        const assetBaseName = path.basename(assetName);
        if (Object.keys(templateNameByAssetName).includes(assetBaseName)) {
          if (bundleInfo.type !== "asset") {
            throw new Error(
              `expected ${bundleName} to be an asset but it was ${bundleInfo.type}`
            );
          }
          if (typeof bundleInfo.source !== "string") {
            throw new Error(`expected ${bundleName}.source to be a string`);
          }
          const root = htmlParser.parse(bundleInfo.source);
          const elements: OurHTMLElement[] = [...fontPreloadElements];
          const head = root.getElementsByTagName("head")[0];

          for (const node of head.childNodes) {
            if (node instanceof HTMLElement) {
              if (node.tagName === "LINK") {
                const href = node.getAttribute("href");
                if (href != null) {
                  const key = nameWithoutHash(path.basename(href));
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
              }
              if (node.tagName === "SCRIPT") {
                const src = node.getAttribute("src");
                if (src != null) {
                  let hashedName = path.basename(src);
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
            }
          }

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

        // Exclude territories.js from manifest.json
        // It is because territories.json is not unique just based on basename.
        // If we include it, then manifest.json will become unstable, causing
        // the reproducible build checking to fail.
        let shouldIncludeInManifest = true;
        if (/territories-[-_A-Za-z0-9]+\.js(\.map)?$/.test(bundleName)) {
          shouldIncludeInManifest = false;
        }
        if (shouldIncludeInManifest) {
          manifest[assetBaseName] = bundleName;
        }
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
          watch: _config.build?.watch && {
            include: "src/**",
          },
          emptyOutDir: true,
          sourcemap: true,
          cssCodeSplit: true,
          // Avoid image assets being inlined into css files
          assetsInlineLimit: 0,
          assetsDir: "shared-assets",
          rollupOptions: {
            // Workaround for building bundles with non-deterministic filenames
            // Active issue: https://github.com/vitejs/vite/issues/13672
            // Workaround from https://github.com/vitejs/vite/issues/10506#issuecomment-1367718113
            maxParallelFileOps: 1,
            input: input,
            output: {
              format: "module",
              manualChunks: (id) => {
                // Keep cldr data separate.
                if (id.includes("node_modules/cldr-localenames-full/")) {
                  return null;
                }

                // Other node_modules assets should be packed into vendor.js
                if (id.includes("node_modules")) {
                  return "vendor";
                }

                // To avoid css files being bundled in 1 file
                if (id.endsWith(".css")) {
                  return path.basename(id);
                }

                return null;
              },
            },
          },
        },
      };
    },
  };
}

function viteAuthgearAuthUI(options: AuthgearAuthUIPluginOptions) {
  return [servePlugin(options), buildPlugin(options)];
}

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
