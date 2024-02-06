# Auth UI

### Material Icons subset

When adding material icons, run the following command in project root to generate the subset of material icons:

```sh
make generate-material-icons
```

After that, commit the changes in `authui/src/authflowv2/icons/material-symbols-outlined-subset.ttf` and `authui/src/authflowv2/icons/material-symbols-outlined-subset.woff2`.

If new icon cannot be auto-parsed i.e. font file not changed, add the icon name to `scripts/python/material-icons.txt` and run the command again.
