# Auth UI

### Material Icons subset

When adding material icons, run the following command in project root to generate the subset of material icons:

```sh
make generate-material-icons
```

After that, commit the changes in `scripts/python/material-icons.txt`, `authui/src/authflowv2/icons/material-symbols-outlined-subset.ttf` and `authui/src/authflowv2/icons/material-symbols-outlined-subset.woff2`.

### @metamask/jazzicon

In oursky/authgear-server, the connection to the URL `https://codeload.github.com/MetaMask/jazzicon/tar.gz/4fe23bbbe5088e128cb24082972e28d87e76d156` fails very often.
Therefore, we download the tarball, and use it directly.
We also updated to the latest commit.
See https://github.com/MetaMask/jazzicon/compare/d923914fda6a8795f74c2e66134f73cd72070667..4fe23bbbe5088e128cb24082972e28d87e76d156 for the changes.
The changes DO NOT cover runtime behavior changes, so it is safe to upgrade.
