This plugin allows to rewrite bundle file names.

## Why?

To get more control over bundle file names. 
You may find another useful plugin - https://github.com/ol-loginov/parcel-reporter-entries to
 get report about actual bundle file names. 

## Install and usage


Install the package first:
```
npm install --save-dev parcel-namer-rewrite
```

And edit `.parcelrc` file to add new namer (put it before or instead default one):
```
/* .parcelrc */
{
  "extends": "@parcel/config-default",
  "namers": [ "parcel-namer-rewrite" ]
}
```


## Configuration

Plugin takes all config from package.json file. Example of config is below:

```json
{
    "name": "...",
    "version": "...",
    "description": "",

  
    "parcel-namer-rewrite": {
        "chain": "@parcel/namer-default",
                
        "rules": {
            "styles/(.*).css": "$1.{hash}.css",
            "scripts/TestMePlease.js": "Dunno.js"
        }
    }
}
```
p.s. Line with "chain" is optional, you may omit that line, and default namer will be used to get actual names

This example:
1) Rewrites all .css bundles in "styles" folder of distDir directly to distDir and adds hashes to names.
2) Moves "scripts/TestMePlease.js" to "Dunno.js"

### How to write rules

"rules" is the object with rules: "key" is RegExp pattern to search in file names, "value" is replacement
 strings (you may use RegExp references here)


### Using hash in file names

File name `"$1.{hash}.css"` adds hash before extension. If hash is empty (in development mode, for example), then you get double dot in file name:
```
"file.{hash}.css" -> "file..css"
```

But there is a solution - just put dot inside brackets. For example: `"file{.hash}.css"` or `"file{hash.}.css"`. In this case for empty hash strings all inside brackets will be removed. And you may use any single character instead of dot: `"file{~hash}.css"`, `"file{-hash}.css"` are valid replacements.

In development mode hashes are blank for file names. You may force hashing in development mode by setting "developmentHashing" to `true`.
Like this:

```json5
{
    "parcel-namer-rewrite": {
        "developmentHashing": true,
                
        "rules": {
            /* some rules  */
        }
    }
}
```

### Hide logs

This plugin will log every renamed file. To disable this feature, add this options to `package.json` :

```json5
{
    "parcel-namer-rewrite": {
        // ...
        "silent" : true
    }
}
```


### Disable in development mode

You may want to disable the namer in development mode. To disable this feature, add this options to `package.json` :

```json5
{
    "parcel-namer-rewrite": {
        // ...
        "developmentDisable" : true
    }
}
```


