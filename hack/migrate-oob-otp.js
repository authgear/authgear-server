#!/usr/bin/env node

const fs = require("fs");
const yaml = require("js-yaml");

const json = fs.readFileSync(process.stdin.fd, "utf-8");
const r = JSON.parse(json);

if (!(r.metadata && r.metadata.labels && r.metadata.labels["authgear.com/app-id"])) {
    process.exit(2);
}

const config = yaml.safeLoad(Buffer.from(r.data["authgear.yaml"], "base64"));

if (!config.authentication) {
  process.exit(2);
}

var updated = false;

if (config.authentication.primary_authenticators) {
  const idx = config.authentication.primary_authenticators.indexOf('oob_otp');
  if (idx != -1) {
    config.authentication.primary_authenticators.splice(idx, 1, 'oob_otp_email', 'oob_otp_sms');
    updated = true;
  }
}

if (config.authentication.secondary_authenticators) {
  const idx = config.authentication.secondary_authenticators.indexOf('oob_otp');
  if (idx != -1) {
    config.authentication.secondary_authenticators.splice(idx, 1, 'oob_otp_sms');
    updated = true;
  }
}

if (!updated) {
  process.exit(2);
}

r.data["authgear.yaml"] = Buffer.from(yaml.safeDump(config)).toString("base64");

console.log(JSON.stringify(r));
