#!/bin/sh

set -x

# This is how we self-host the runtime assets of FluentUI.
#
# We utilize the public directory feature of Vite.
# See https://v2.vitejs.dev/guide/assets.html#the-public-directory
# This feature merely copies the files in the public directory to the root of the outDir.
# Since we use FileServer to serve asset, assets have to be put in the asset directory.
# We automate this copy process with a NPM postinstall script. (This script)
#
# Finally we tell fluentui to load the assets from the portal backend, instead of from the default CDN.
# This is done with window.FabricConfig.

# In docker build, the postinstall runs before the src are copied.
# So we run mkdir -p to ensure the directory exist.
mkdir -p ./src/public/shared-assets/
# In case you wonder why we do not just use shell expansion here,
# if ./src/public/shared-assets is really empty, sh DOES NOT expand, and take '*' literally.
# Since we do not have such a file, the command will fail.
find ./src/public/shared-assets -name 'fabric-icons-*.woff' -print -exec rm '{}' \;
# When window.FabricConfig.iconBaseUrl is set, it loads the font directly in the directory.
# So we just copy the fonts to outDir.
cp -R ./node_modules/@fluentui/font-icons-mdl2/fonts/. ./src/public/shared-assets/.

# When window.FabricConfig.fontBaseUrl is set, it loads the font with a certain structure.
# The original URL is https://static2.sharepointonline.com/files/fabric/assets/fonts/segoeui-westeuropean/segoeui-bold.woff2
# When fontBaseUrl is set, the URL looks like https://origin/shared-assets/fonts/segoeui-westeuropean/segoeui-bold.woff2

# FluentUI actually has support for many fonts.
# For the full list of the fonts it may load at runtime, see ./node_modules/@fluentui/react/dist/css/fabric.css
# Since our site is lang=en, it will ever load "Segoe UI Web (West European)"
# So we just download and copy them.
# Since this process has to be done once only, the following commands are commented out.
# wget https://static2.sharepointonline.com/files/fabric/assets/fonts/segoeui-westeuropean/segoeui-light.woff2 -O ./src/public/shared-assets/fonts/segoeui-westeuropean/segoeui-light.woff2
# wget https://static2.sharepointonline.com/files/fabric/assets/fonts/segoeui-westeuropean/segoeui-light.woff -O ./src/public/shared-assets/fonts/segoeui-westeuropean/segoeui-light.woff
# wget https://static2.sharepointonline.com/files/fabric/assets/fonts/segoeui-westeuropean/segoeui-semilight.woff2 -O ./src/public/shared-assets/fonts/segoeui-westeuropean/segoeui-semilight.woff2
# wget https://static2.sharepointonline.com/files/fabric/assets/fonts/segoeui-westeuropean/segoeui-semilight.woff -O ./src/public/shared-assets/fonts/segoeui-westeuropean/segoeui-semilight.woff
# wget https://static2.sharepointonline.com/files/fabric/assets/fonts/segoeui-westeuropean/segoeui-regular.woff2 -O ./src/public/shared-assets/fonts/segoeui-westeuropean/segoeui-regular.woff2
# wget https://static2.sharepointonline.com/files/fabric/assets/fonts/segoeui-westeuropean/segoeui-regular.woff -O ./src/public/shared-assets/fonts/segoeui-westeuropean/segoeui-regular.woff
# wget https://static2.sharepointonline.com/files/fabric/assets/fonts/segoeui-westeuropean/segoeui-semibold.woff2 -O ./src/public/shared-assets/fonts/segoeui-westeuropean/segoeui-semibold.woff2
# wget https://static2.sharepointonline.com/files/fabric/assets/fonts/segoeui-westeuropean/segoeui-semibold.woff -O ./src/public/shared-assets/fonts/segoeui-westeuropean/segoeui-semibold.woff
# wget https://static2.sharepointonline.com/files/fabric/assets/fonts/segoeui-westeuropean/segoeui-bold.woff2 -O ./src/public/shared-assets/fonts/segoeui-westeuropean/segoeui-bold.woff2
# wget https://static2.sharepointonline.com/files/fabric/assets/fonts/segoeui-westeuropean/segoeui-bold.woff -O ./src/public/shared-assets/fonts/segoeui-westeuropean/segoeui-bold.woff
