# Auth UI

### Material Icons subset

When adding material icons, run the following command in project root to generate the subset of material icons:

```sh
make generate-material-icons
```

After that, commit the changes in `scripts/python/material-icons.txt`, `authui/src/authflowv2/icons/material-symbols-outlined-subset.ttf` and `authui/src/authflowv2/icons/material-symbols-outlined-subset.woff2`.

### ./tarballs/

As a workaround, we can put tarballs in ./tarballs to install dependency from a tarball.
This usage is not encouraged.
