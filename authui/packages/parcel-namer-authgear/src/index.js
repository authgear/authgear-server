const { Namer } = require("@parcel/plugin");
const path = require("path");

const NAMING_LIST = [
  { asset: "src/authgear.ts", newName: "authgear", type: "js" },
  { asset: "src/tailwind.css", newName: "tailwind", type: "css" },

  {
    asset: "intl-tel-input/build/js/intlTelInput.min.js",
    newName: "intlTelInput",
    type: "js",
  },
  {
    asset: "intl-tel-input/build/js/utils.js",
    newName: "intlTelInputUtils",
    type: "js",
  },
  {
    asset: "intl-tel-input/build/css/intlTelInput.min.css",
    newName: "intlTelInput",
    type: "css",
  },

  {
    asset: "@tabler/icons/iconfont/tabler-icons.min.css",
    newName: "tabler-icons",
    type: "css",
  },

  {
    asset: "cropperjs/dist/cropper.min.js",
    newName: "cropper",
    type: "js",
  },
  {
    asset: "cropperjs/dist/cropper.min.css",
    newName: "cropper",
    type: "css",
  },

  {
    asset: "@hotwired/stimulus/dist/stimulus.js",
    newName: "stimulus",
    type: "js",
  },
  {
    asset: "@hotwired/turbo/dist/turbo.es2017-esm.js",
    newName: "turbo",
    type: "js",
  },
  {
    asset: "zxcvbn/lib/main.js",
    newName: "zxcvbn",
    type: "js",
  },
  {
    asset: "axios/index.js",
    newName: "axios",
    type: "js",
  },
  {},
];

module.exports = new Namer({
  name({ bundle }) {
    const asset = bundle.getMainEntry() ?? bundle.getEntryAssets()[0];
    if (!asset) {
      return null;
    }

    const assetPath = asset.filePath;
    const nameFromSuper = path.basename(assetPath, path.extname(assetPath));

    for (const item of NAMING_LIST) {
      if (assetPath.includes(item.asset) && item.type === bundle.type) {
        let name = item.newName;
        if (bundle.type == "js") {
          if (bundle.env.outputFormat == "esmodule") {
            name = item.newName + "-module";
          }
          if (bundle.env.outputFormat == "global") {
            name = item.newName + "-classic";
          }
        }
        return name + "." + bundle.hashReference + "." + item.type;
      }
    }

    return null;
  },
});
