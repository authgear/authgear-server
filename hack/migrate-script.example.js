#!/usr/bin/env node

const fs = require("fs");
const yaml = require("js-yaml");

const json = fs.readFileSync(process.stdin.fd, "utf-8");
const r = JSON.parse(json);

if (!(r.metadata && r.metadata.labels && r.metadata.labels["authgear.com/app-id"])) {
    process.exit(2);
}

const config = yaml.safeLoad(Buffer.from(r.data["authgear.secrets.yaml"], "base64"));
for (const item of config.secrets) {
    if (item.key === "oauth") {
        item.key = "oidc";
    }
}
r.data["authgear.secrets.yaml"] = Buffer.from(yaml.safeDump(config)).toString("base64");

console.log(JSON.stringify(r));
