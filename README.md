<a href="https://www.authgear.com/?utm_source=github&utm_medium=readme&utm_campaign=logo"><img src="https://raw.githubusercontent.com/authgear/.github/main/profile/authgear-logo-github.svg" width="400" alt="Authgear Logo">
</a>

<h3>
  <a href="https://docs.authgear.com/?utm_source=github&utm_medium=readme&utm_campaign=top_links">📘 Docs</a>
  | <a href="https://www.authgear.com/?utm_source=github&utm_medium=readme&utm_campaign=top_links">☁️ SaaS Cloud</a>
  | <a href="https://demo.authgear.com/?utm_source=github&utm_medium=readme&utm_campaign=top_links">✨ Demo</a>
  | <a href="https://discord.gg/Kdn5vcYwAS">💬 Discord</a>
</h3>

[![checks](https://img.shields.io/github/check-runs/authgear/authgear-server/main)](https://github.com/authgear/authgear-server/actions?query=branch%3Amain)
[![release](https://img.shields.io/github/v/release/authgear/authgear-server)](https://github.com/authgear/authgear-server/releases)
[![cloud](https://img.shields.io/badge/cloud-available-green)](https://demo.authgear.com/?utm_source=github&utm_medium=readme&utm_campaign=badge)
[![discord](https://img.shields.io/discord/918079010917982220?label=discord&labelColor=5865F2
)](https://discord.gg/Kdn5vcYwAS)


# Authgear: Open source alternative to Auth0 / Clerk / Firebase Auth

Authgear is an open-source extensible turnkey solution for all of your consumer authentication needs. Authgear gets you started in 5 minutes with developer-friendly SDKs and a comprehensive portal.

Available for self-hosting and on [Authgear Cloud](https://www.authgear.com/?utm_source=github&utm_medium=readme&utm_campaign=intro).

With a wide range of out-of-the-box features, it's perfect for SaaS product apps and multi-apps ecosystem, such as:

- Passwordless login: Magic-link/OTP with Email, SMS, WhatsApp
- Passkeys
- Pre-built signup/login page
- Pre-built user account settings page
- Biometric Login on iOS and Android
- 2FA: TOTP (Google Authenticator, Authy), SMS, Email
- Integration with analytics, CDP, and drip campaigns
- Enterprise Security: Audit logs, Bruteforce Protection, Bot Production, Rate Limits
- Modern authentication and authorization protocols and SSO: OIDC/OAuth 2.0/SAML
- B2B Enterprise Connections: ADFS, LDAP
- and more...

**Why developers use Authgear**

- **Authgear Portal**: Web-interface for user management and setting up authentication/authorization configurations
- **Admin API**: Powerful GraphQL API to manage resources and all things authentication
- **End-user Experience**: Beautiful and tailorable out-of-the-box pre-built authentication flows
- **Enterprise-grade security**: MFA, SSO, RBAC, and audit logs.

Contact us: <br>
[![schedule a demo](https://img.shields.io/badge/schedule%20a%20demo-0b63e9)](https://www.authgear.com/schedule-demo/?utm_source=github&utm_medium=readme&utm_campaign=contact_us)


## Who is using Authgear

We're grateful to the companies listed below for their ongoing support and significant impact on our community. If you want to join the list, email us at hello@authgear.com!

<table>
    <tr>
    <td>
    <a href="https://www.bupa.com.hk/">
    <img src="https://raw.githubusercontent.com/authgear/.github/main/meta/adopters/bupa.png" alt="Bupa"></a></td>
    <td><a href="https://www.cimic.com.au/"><img src="https://raw.githubusercontent.com/authgear/.github/main/meta/adopters/cimic-group.png" alt="Cimic Group"></a></td>
    <td><a href="https://www.hkland.com"><img src="https://raw.githubusercontent.com/authgear/.github/main/meta/adopters/hongkong-land.png" alt="Hongkong Land"></a></td>
    <td><a href="https://www.k11.com/"><img src="https://raw.githubusercontent.com/authgear/.github/main/meta/adopters/k11.png" alt="K11"></a></td>
    <td><a href="https://www.mtr.com.hk/"><img src="https://raw.githubusercontent.com/authgear/.github/main/meta/adopters/mtr.png" alt="MTR"></a></td>
    </tr>
</table>

## Features and Components

The repo `authgear-server` includes the following components of Authgear:

- Authgear server (the core service)
- Portal (a web-interface for managing configurations in Authgear projects)
- AuthUI  (a customizable User Interface (UI) for login, user registration, and profile settings pages)
- Admin API (provides a GraphQL interface for developers to interact with services and data on Authgear)

This repo is the open-source project that powers Authgear's authentication-as-a-service solution. It includes the code for the server, AuthUI, the Portal, and Admin API. You can use it to set up your own self-hosted instance of Authgear service.

### Authgear SDK

In addition to Authgear Server, we provide SDKs that developers can use to integrate Authgear into their apps.

These SDKs exist as standalone projects under the following repositories:

- [JavaScript/React native/Capacitor](https://github.com/authgear/authgear-sdk-js)
- [iOS](https://github.com/authgear/authgear-sdk-ios)
- [Android](https://github.com/authgear/authgear-sdk-android)
- [Flutter](https://github.com/authgear/authgear-sdk-flutter)
- [Xamarin](https://github.com/authgear/authgear-sdk-xamarin)

## Documentation and getting started
The easiest way to start is to sign up at [authgear.com](https://authgear.com) for a free account.

[✅ Quick Start Guide](https://docs.authgear.com/get-started/start-building)

Our Quick Start Guide includes tutorials and code examples for popular programming languages, tools, and frameworks like JavaScript, Go, PHP, Next.js, Laravel, Spring, and more.

For more details about getting started with using Authgear, check out the official documentation site at https://docs.authgear.com.

Also, you can take a look at our [example projects repos](https://github.com/orgs/authgear/repositories?language=&q=example&sort=&type=all) that demonstrate how to use Authgear.

## Installation and setup

The Authgear Server project allows developers to set up their own instance of Authgear.

We've provided detailed instructions on how to set up a self-hosted instance of Authgear here: https://docs.authgear.com/deploy-on-your-cloud/local

[Helm Chart](https://docs.authgear.com/deploy-on-your-cloud/helm) is the recommended way to deploy Authgear on Kubernetes for production usage

## How to contribute
Please refer to [CONTRIBUTING.md](CONTRIBUTING.md) if you need instructions on contributing to the development of Authgear Server.

## Contributors

Currently there are 42 contributors for this repository. Feel free to contribute!

<a href="https://github.com/authgear/authgear-server/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=authgear/authgear-server" />
</a>

<small>Made with [contrib.rocks](https://contrib.rocks).</small>

## Credits

- Free email provider domains list provided by: https://gist.github.com/tbrianjones/5992856/
- This product includes GeoLite2 data created by MaxMind, available from https://www.maxmind.com

---

Part of [Skymarkers](https://www.skymakers.digital/). We 😻 open-source.
