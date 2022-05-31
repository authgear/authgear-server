# parcel-reporter-bundle-manifest

This will add `parcel-manifest.json` to the target dir. Example:

```json
{
  "index.html": "/index.html",
  "layout.css": "/layout.f955332a.css",
  "editor.css": "/editor.e1160f52.css",
  "editor.tsx": "/editor.9099e93e.js"
}
```

## Installation

```sh
npm install --save-dev parcel-reporter-bundle-manifest
```

## Usage

Add `parcel-reporter-bundle-manifest` to `.parcelrc` in `reporters`.

```json
{
  "extends": "@parcel/config-default",
  "reporters": ["...", "parcel-reporter-bundle-manifest"]
}
```

## More info:

- https://github.com/parcel-bundler/parcel#parcelrcreporters
- https://github.com/parcel-bundler/parcel#reporters

## Development

### Releasing

1. Bump the version in `package.json`
2. Push to the `main` branch in GitHub
3. Create a release for that version

## Acknowledgement

This plugin behave similarly to https://github.com/mugi-uno/parcel-plugin-bundle-manifest.

## License

parcel-reporter-bundle-manifest © [Autify Engineers](https://github.com/autifyhq). Released under the [MIT License](LICENSE).<br/>
Authored and maintained by [Autify Engineers](https://github.com/autifyhq) with help from [contributors](https://github.com/autifyhq/parcel-reporter-bundle-manifest/graphs/contributors).
