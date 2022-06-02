const { Reporter } = require("@parcel/plugin");
const path = require("path");

module.exports = new Reporter({
  async report({ event, options }) {
    if (event.type != "buildSuccess") {
      return;
    }

    const manifest = {};

    const bundles = event.bundleGraph.getBundles();
    for (const bundle of bundles) {
      const assetName = bundle.displayName.replace(".[hash].", ".");
      manifest[assetName] = path.relative(
        bundle.target.distDir,
        bundle.filePath
      );
    }

    const targetPath = "../pkg/lib/webparcel/parcel_gen.go";
    const tpl = `package webparcel

func init() {
    ParcelAssetMap = map[string]string${JSON.stringify(manifest)}
}`;
    await options.outputFS.writeFile(targetPath, tpl);
    console.log(`📄 Wrote bundle manifest to: ${targetPath}`);
  },
});
