## 2.3.0 (2020-4-02)

### Features

- New auth ui support
- Update Skygear Auth as OpenID Provider
- Support app and gears endpoint in gateway

### Breaking Changes

- Web hooks do not resolve relative path for hook URLs #1260
- Disable auth API if not enabled, auth API is disabled by default
- Update AppConfiguration

### Other notes

- Update facebook RP to use v6.0 Graph API

## 2.2.0 (2020-2-28)

### Features

- Support static asset routing #1201

### Bug Fixes

- Fix compound authz policy short-circuit behavior #1218
- Fix Cache-Control header format
- Fix Ignore query and fragment in validating allowed callback URLs in SSO #1211

### Other notes

- Update sample yaml

## 2.1.0 

- Initial release
