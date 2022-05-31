import {Namer} from '@parcel/plugin';
import {Config} from "./Config";
import {PluginLogger} from '@parcel/logger';
import crypto from "crypto";
import {md5FromFilePath} from "@parcel/utils";

const CONFIG = Symbol.for('parcel-plugin-config');

// noinspection JSUnusedGlobalSymbols
export default new Namer({
    config: Config | undefined,
    delegate: null,

    async name(opts: { bundle: Bundle, bundleGraph: {}, options: {}, logger: PluginLogger }) {
        const config = this.ensureConfig(opts.options.projectRoot, opts.options.packageManager, opts.logger);
        if (!config) {
            return null;
        }

        const disable = config.developmentDisable && opts.options.mode === 'development';
        let nameFromSuper = null
        
        const asset = opts.bundle.getMainEntry() ?? opts.bundle.getEntryAssets()[0];
        if (asset) {
            const assetPath = asset.filePath;
            nameFromSuper = assetPath.substring(assetPath.lastIndexOf("/") + 1)

            const i = nameFromSuper.lastIndexOf(".")
            if (nameFromSuper.substring(i) === ".ts") {
                nameFromSuper = nameFromSuper.substring(0, i) + ".js"
            }
        }
        if (nameFromSuper != null && !disable) {
            return this.rewrite(opts.bundle, opts.bundleGraph, opts.options, nameFromSuper, opts.logger);
        }
        return nameFromSuper;
    },

    ensureConfig(projectRoot: string, packageManager: {}, logger: PluginLogger) {
        if (!this.config) {
            const config = new Config();
            config.loadFromPackageFolder(projectRoot, logger);
            if (!config.chain) {
                throw Error('No chain namer has been found in project. Set package.json#parcel-namer-rewrite:chain to set a delegate namer ("@parcel/namer-default" by default)');
            }

            const delegatePackage = packageManager.load(config.chain, projectRoot);
            if (!delegatePackage) {
                throw Error(`'Package ${config.delegate}' is not available. Set package.json#parcel-namer-rewrite:chain to set a delegate namer ("@parcel/namer-default" by default)`);
            }

            const delegate = delegatePackage.default[CONFIG];
            if (!delegate) {
                throw Error(`Package '${config.delegate}' has been found, but it's not a namer. Set package.json#parcel-namer-rewrite:chain to set a delegate namer ("@parcel/namer-default" by default)`);
            }

            this.delegate = delegate;
            this.config = config;
        }
        return this.config;
    },

    async rewrite(bundle: { id: string }, bundleGraph: {}, options: {}, superName: string, logger) {
        const rule = this.config.selectRule(superName);
        if (!rule) {
            return superName;
        }

        let bundleHash = '';

        if (options.mode !== 'development' || this.config.developmentHashing) {
            if (this.config.useParcelHash) {
                bundleHash = bundle.hashReference;
            } else {
                let assets = [];
                bundle.traverseAssets((asset) => assets.push(asset));

                let hash = crypto.createHash('md5');
                for (let i = 0; i < assets.length; ++i) {
                    const asset = assets[i];
                    if (asset.filePath) {
                        const fileHash = await md5FromFilePath(asset.fs, asset.filePath);
                        hash.update([asset.filePath, fileHash].join(':'));
                    }
                }

                bundleHash = hash.digest('hex').substr(0, 6);
            }
        }

        // if we need hashing - remove bundle hash placeholder
        if (bundleHash && bundle.hashReference) {
            superName = superName.replace("." + bundle.hashReference, "")
        }

        const rewrite = superName
            .replace(rule.test, rule.to)
            .replace(/{(.?)hash(.?)}/, bundleHash.length > 0 ? `$1${bundleHash}$2` : '');

        if (this.config.silent !== true)
            logger.info({
                message: `Rewrite ${superName} -> ${rewrite}`
            });

        return rewrite;
    }
});
