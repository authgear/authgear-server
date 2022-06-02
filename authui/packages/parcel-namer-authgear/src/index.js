const { Namer } = require("@parcel/plugin");
const path = require("path");

const NAMING_LIST = [
  { asset: "authgear", newName: "authgear", type: "js" },
  { asset: "tailwind", newName: "tailwind", type: "css" },
];

module.exports = new Namer({
  name({ bundle }) {
    const asset = bundle.getMainEntry() ?? bundle.getEntryAssets()[0];
    const assetPath = asset.filePath;
    const nameFromSuper = path.basename(assetPath, path.extname(assetPath));

    for (const item of NAMING_LIST) {
      if (item.asset === nameFromSuper && item.type === bundle.type) {
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
