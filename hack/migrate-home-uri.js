#!/usr/bin/env node

const fs = require("fs");
const yaml = require("js-yaml");

const json = fs.readFileSync(process.stdin.fd, "utf-8");
const r = JSON.parse(json);

if (!(r.metadata && r.metadata.labels && r.metadata.labels["authgear.com/app-id"])) {
    process.exit(2);
}

const config = yaml.safeLoad(Buffer.from(r.data["authgear.yaml"], "base64"));

if (!config.ui || !config.ui.home_uri) {
    process.exit(2);
}

const a = config.ui.home_uri;
config.ui.default_client_uri = a;
config.ui.default_redirect_uri = a;
config.ui.default_post_logout_redirect_uri = a;
delete config.ui.home_uri;
r.data["authgear.yaml"] = Buffer.from(yaml.safeDump(config)).toString("base64");

console.log(JSON.stringify(r));
