#!/usr/bin/env node

const fs = require("fs");
const yaml = require("js-yaml");

const json = fs.readFileSync(process.stdin.fd, "utf-8");
const r = JSON.parse(json);

if (!(r.metadata && r.metadata.labels && r.metadata.labels["authgear.com/app-id"])) {
    process.exit(2);
}

const config = yaml.safeLoad(Buffer.from(r.data["authgear.yaml"], "base64"));

if (!config.localization || !config.localization.fallback_language) {
    process.exit(2);
}
if (config.localization.fallback_language === "en") {
    process.exit(2);
}

config.localization.supported_languages = [
    "en",
    config.localization.fallback_language,
];
r.data["authgear.yaml"] = Buffer.from(yaml.safeDump(config)).toString("base64");

console.log(JSON.stringify(r));
