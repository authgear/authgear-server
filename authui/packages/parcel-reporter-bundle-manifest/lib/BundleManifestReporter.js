var $5J5yU$parcelplugin = require("@parcel/plugin");
var $5J5yU$path = require("path");

function $parcel$interopDefault(a) {
  return a && a.__esModule ? a.default : a;
}
function $parcel$defineInteropFlag(a) {
  Object.defineProperty(a, '__esModule', {value: true, configurable: true});
}
function $parcel$export(e, n, v, s) {
  Object.defineProperty(e, n, {get: v, set: s, enumerable: true, configurable: true});
}

$parcel$defineInteropFlag(module.exports);

$parcel$export(module.exports, "default", () => $fd2c454dbd295085$export$2e2bcd8739ae039);
$parcel$export(module.exports, "MANIFEST_FILENAME", () => $fd2c454dbd295085$export$942f2ec118ed46e4);


const $fd2c454dbd295085$var$normalisePath = (p)=>{
    return p.replace(/[\\/]+/g, "/");
};
var $fd2c454dbd295085$export$2e2bcd8739ae039 = new $5J5yU$parcelplugin.Reporter({
    // TODO: Add type definition for Reporter
    async report ({ event: event , options: options  }) {
        if (event.type !== "buildSuccess") return;
        const bundlesByTarget = new Map();
        for (const bundle of event.bundleGraph.getBundles())if (!bundle.isInline) {
            let bundles = bundlesByTarget.get(bundle.target.distDir);
            if (!bundles) {
                bundles = [];
                bundlesByTarget.set(bundle.target.distDir, bundles);
            }
            bundles.push(bundle);
        }
        const assetNames = [];
        for (const [targetDir, bundles] of bundlesByTarget){
            const manifest = {
            };
            for (const bundle of bundles){
                var ref;
                /**
         * Use main entry first as the key of the manifest, and fallback to the first asset of the bundle if main entry doesn't exist.
         *
         * Some bundle doesn't have a main entry (`bundle.getMainEntry()`); e.g. CSS bundle that's the result of CSS files imported from JS.
         *
         * The bundle could have multiple assets; e.g. multiple CSS files combined into one bundle,
         * so we only choose the first one to avoid multiple bundle in the manifest.
         *
         * We cannot use the bundled file name without hash as a key because there' might be only hash; e.g. styles.css -> asdfjkl.css.
         */ const asset = (ref = bundle.getMainEntry()) !== null && ref !== void 0 ? ref : bundle.getEntryAssets()[0];
                if (asset) {
                    const assetPath = asset.filePath;
                    const entryRoot = event.bundleGraph.getEntryRoot(bundle.target);
                    var assetName = $fd2c454dbd295085$var$normalisePath(($parcel$interopDefault($5J5yU$path)).relative(entryRoot, assetPath));
                    if (assetNames.includes(assetName)) {
                        const i = assetName.lastIndexOf(".");
                        assetName = assetName.substring(0, i) + "-modern" + assetName.substring(i);
                    }
                    assetNames.push(assetName);
                    const bundleUrl = $fd2c454dbd295085$var$normalisePath(`${bundle.target.publicUrl}/${($parcel$interopDefault($5J5yU$path)).relative(bundle.target.distDir, bundle.filePath)}`);
                    manifest[assetName] = bundleUrl;
                }
            }
            const targetPath = `${targetDir}/${$fd2c454dbd295085$export$942f2ec118ed46e4}`;
            await options.outputFS.writeFile(targetPath, JSON.stringify(manifest));
            console.log(`📄 Wrote bundle manifest to: ${targetPath}`);
        }
    }
});
const $fd2c454dbd295085$export$942f2ec118ed46e4 = "parcel-manifest.json";


//# sourceMappingURL=BundleManifestReporter.js.map
