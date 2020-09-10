# Templates

Authgear serves web pages and send email and SMS messages. Templates allow the developer to provide localization or even customize the default ones.

  * [Template](#template)
    * [Template Type](#template-type)
    * [Template Language Tag](#template-language-tag)
  * [Template Resolution](#template-resolution)
  * [Component Templates](#component-templates)
  * [Localization of the text of the template](#localization-of-the-text-of-the-template)
  * [localize](#localize)
  * [Translation file](#translation-file)
    * [Translation Resolution](#translation-resolution)
  * [Available templates](#available-templates)
    * [otp_message](#otp_message)

## Template

Each template must have a type, optionally a key and language tag.

### Template Type

Each template must have a type. The list of types is predefined. Here are some examples:

```
forgot_password_email.html
forgot_password_email.txt

user_verification_message.html
user_verification_message.txt
```

### Template Language Tag

Each template may optionally have a language tag. The language tag is specified in [BCP47](https://tools.ietf.org/html/bcp47).

## Template Resolution

To resolve a template, the input is the template type and user preferred languages. The type is determined by the feature while the user preferred languages is provided by the user.

All templates have default value so template resolution always succeed.

The templates are first resolved by matching the type, and then select the best language according to the user preferred languages.

## Component Templates

Some template may depend on other templates which are included during rendering. This enables customizing a particular component of a template. The dependency is expressed by a whitelist that is hard-coded by the Authgear developer. It can be assumed there is no dependency cycle.

For example, `auth_ui_login.html` depend on `auth_ui_header.html` to provide the header. If the developer just wants to customize the header, they do not need to provide customized templates for ALL pages. They just need to provide `auth_ui_header.html`.

## Localization of the text of the template

In addition to the template language tag, sometimes it is preferred to localize the text of the template rather the whole template.

For example, `auth_ui_login.html` defines the HTML structure and is used for all languages. What the developer wants to localize is the text.

Each translation key is parsed as template so you can just use `{{ template }}` to refer it. This allows HTML in the translation.

```html
<input type="password" placeholder="{{ template "enter.password" }}">
<!-- <input type="password placeholder="Enter Password"> -->
```

```html
<p>{{ template "email.sent" (makemap "email" .email "name" .name) }}</p>
<!-- <p>Hi John, an email has been sent to john.doe@example.com</p -->
```

## Translation file

The translation file is a template itself. It is simply a flat JSON object with string keys and string values. The value is in ICU MessageFormat. Not all ICU MessageFormat arguments are supported. The supported are `select`, `plural` and `selectordinal`.

Here is an example of the translation file.

```json5
{
  "email.sent": "Hi {1}, an email has been sent to {0}"
}
```

### Translation Resolution

Translation resolution is different from template resolution. Template resolution is file-based while translation resolution is key-based.

For example,

```json5
// The zh variant of auth_ui_translation.json
{
  "enter.password": "輸入密碼",
  "enter.email": "輸入電郵地址"
}
```

```json5
// The zh-Hant-HK variant of auth_ui_translation.json
{
  "enter.password": "入你嘅密碼"
}
```

And the user preferred languages is `["zh-Hant-HK"]`.

`"enter.password"` resolves to `"入你嘅密碼"` and `"enter.email"` resolves to `"輸入電郵地址"`.

## Available templates

> TODO: WIP need update

### `otp_message`

One-time-password message. Used for authentication and user verification.

- Template types:
    - `otp_message_email.txt`
    - `otp_message_email.html`
    - `otp_message_sms.txt`
- Context:
    - `AppName`: the display name of the app.
    - `Email`: The recipient email of the message; empty if not sending an email message.
    - `Phone`: The recipient phone number of the message; empty if not sending an SMS message.
    - `LoginID`: The login ID of the identity
        - `LoginID.Key`: Login ID key
        - `LoginID.Value`: Login ID
    - `Code`: The one-time-password.
    - `Host`: Host of authgear, usually used for [Web OTP](https://web.dev/web-otp/) API.
    - `Origin`: The origin page of the OTP message, can be `signup`/`login`/`settings`.
    - `Operation`: The operation triggering the OTP message, can be `primary_auth`/`secondary_auth`/`verify`.
