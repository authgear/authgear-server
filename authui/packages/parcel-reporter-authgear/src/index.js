const { Reporter } = require("@parcel/plugin");
const { parse } = require("node-html-parser");
const path = require("path");

module.exports = new Reporter({
  async report({ event, options }) {
    if (event.type != "buildSuccess") {
      return;
    }

    const manifest = {};
    const elements = [];

    const bundles = event.bundleGraph.getBundles();
    for (const bundle of bundles) {
      const assetName = bundle.displayName.replace(".[hash].", ".");
      if (assetName === "build.html") {
        const htmlFile = await options.outputFS.readFile(bundle.filePath);
        const htmlString = htmlFile.toString();
        const root = parse(htmlString);
        const head = root.getElementsByTagName("head")[0];

        for (const node of head.childNodes) {
          if (
            node.tagName === "LINK" &&
            node.getAttribute("rel") === "stylesheet"
          ) {
            const hashedName = node.getAttribute("href").substring(1);
            const key = nameWithoutHash(hashedName);
            elements.push({
              type: "css",
              name: key,
            });
          }
          if (node.tagName === "SCRIPT") {
            const hashedName = node.getAttribute("src").substring(1);
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
      manifest[assetName] = path.relative(
        bundle.target.distDir,
        bundle.filePath
      );
    }

    const targetManifestPath =
      "../resources/authgear/static/generated/manifest.json";
    await options.outputFS.writeFile(
      targetManifestPath,
      JSON.stringify(manifest)
    );
    console.log(`📄 Wrote bundle manifest to: ${targetManifestPath}`);

    const targetHTMLPath =
      "../resources/authgear/templates/en/web/__generated_asset.html";
    const tpl = `{{ define "__generated_asset.html" }}
${elementsTohtmlString(elements).join("\n")}
{{ end }}`;
    await options.outputFS.writeFile(targetHTMLPath, tpl);
    console.log(`📄 Wrote bundle HTML to: ${targetHTMLPath}`);
  },
});

function nameWithoutHash(path) {
  const textArray = path.split(".");
  if (textArray.length === 3) {
    return textArray[0] + "." + textArray[2];
  }
  return path;
}

function stringifyHTMLAttributes(attributes) {
  const result = [];
  for (const [key, value] of Object.entries(attributes)) {
    if (key === "src") {
      continue;
    }
    result.push(`${key}="${value}"`);
  }
  return result.join(" ");
}

function elementsTohtmlString(elements) {
  const textArray = [];
  for (const element of elements) {
    if (element.type === "css") {
      if (element.name === "tailwind-dark-theme.css") {
        textArray.push(`{{ if $.DarkThemeEnabled }}`);
        textArray.push(
          `<link rel="stylesheet" href="{{ call $.GeneratedStaticAssetURL "${element.name}" }}">`
        );
        textArray.push(`{{ end }}`);
      } else {
        textArray.push(
          `<link rel="stylesheet" href="{{ call $.GeneratedStaticAssetURL "${element.name}" }}">`
        );
      }
    }
    if (element.type === "js") {
      textArray.push(
        `<script ${stringifyHTMLAttributes(
          element.attributes
        )} src="{{ call $.GeneratedStaticAssetURL "${
          element.name
        }" }}"></script>`
      );
    }
  }
  return textArray;
}
