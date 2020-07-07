# Templates

Authgear serves web pages and send email and SMS messages. Templates allow the developer to provide localization or even customize the default ones.

  * [Template](#template)
    * [Template Type](#template-type)
    * [Template Key](#template-key)
    * [Template Language Tag](#template-language-tag)
  * [Template Resolution](#template-resolution)
  * [Component Templates](#component-templates)
  * [Localization of the text of the template](#localization-of-the-text-of-the-template)
  * [localize](#localize)
  * [Translation file](#translation-file)
    * [Translation Resolution](#translation-resolution)

## Template

Each template must have a type, optionally a key and a language tag.

### Template Type

Each template must have a type. The list of types are predefined. Here are some examples

```
forgot_password_email.html
forgot_password_email.txt

user_verification_message.html
user_verification_message.txt
```

### Template Key

Some template may require a key. The key is used differentiate different instances of the same type of the template. For example, the verification message template of email message should be different from that of SMS message.

### Template Language Tag

Each template may optionally have a language tag. The language tag is specified in [BCP47](https://tools.ietf.org/html/bcp47).

## Template Resolution

To resolve a template, the input is the template type, optionally the template key and finally the user preferred languages. The type and key is determined by the feature while the user preferred languages is provided by the user.

All templates have default value so template resolution always succeed.

The templates are first resolved by matching the type and the key. And then select the best language according to the user preferred languages.

## Component Templates

Some template may depend on other templates which are included during rendering. This enables customizing a particular component of a template. The dependency is expressed by a whitelist that is hard-coded by the Authgear developer. It can be assumed there is no dependency cycle.

For example, `auth_ui_login.html` depend on `auth_ui_header.html` and `auth_ui_footer.html` to provide the header and footer. If the developer just wants to customize the header, they do not need to provide customized templates for ALL pages. They just need to provide `auth_ui_header.html`.

## Localization of the text of the template

In addition to the template language tag, sometimes it is preferred to localize the text of the template rather the whole template.

For example, `auth_ui_login.html` defines the HTML structure and is used for all languages. What the developer wants to localize is the text.

## localize

A special function named `localize` can be used to format a localized string.

```html
<input type="password" placeholder="{{ localize "enter.password" }}">
<!-- <input type="password placeholder="Enter Password"> -->
```

```html
<p>{{ localize "email.sent" .email .name }}</p>
<!-- <p>Hi John, an email has been sent to john.doe@example.com</p -->
```

`localize` takes a translation key, followed any arguments required by that translation key. If the key is not found, the key itself is returned.

## Translation file

The translation file is a template itself. It is simply a flat JSON object with string keys and string values. The value is in ICU MessageFormat. Not all ICU MessageFormat arguments are supported. The supported are `select`, `plural` and `selectordinal`.

Here is an example of the translation file.

```json
{
  "email.sent": "Hi {1}, an email has been sent to {0}"
}
```

### Translation Resolution

Translation resolution is different from template resolution. Template resolution is file-based while translation resolution is key-based.

For example,

```json
// The zh variant of auth_ui_translation.json
{
  "enter.password": "輸入密碼",
  "enter.email": "輸入電郵地址"
}
```

```json
// The zh-Hant-HK variant of auth_ui_translation.json
{
  "enter.password": "入你嘅密碼"
}
```

And the user preferred languages is `["zh-Hant-HK"]`.

`"enter.password"` resolves to `"入你嘅密碼"` and `"enter.email"` resolves to `"輸入電郵地址"`.
