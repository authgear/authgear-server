# Setup Whatsapp Cloud API

## Prerequisite

You need the following to setup whatsapp cloud api in authgear:

- A facebook business
- A facebook app
- A whatsapp business account with a verified phone number

## Configure secrets

```yaml
- data:
    phone_number_id: "YOUR_PHONE_NUMBER_ID"
    access_token: YOUR_ACCESS_TOKEN
    authentication_template:
      type: copy_code_button
      copy_code_button:
        name: YOUR_TEMPLATE_NAME
        languages:
          - en # Add other languages
    webhook:
      verify_token: YOUR_VERIFY_TOKEN
      app_secret: YOUR_APP_SECRET
  key: whatsapp.cloud-api
```

You should be able to find YOUR_PHONE_NUMBER_ID in your facebook app dashboard, in "WhatsApp" -> "API Setup".

YOUR_ACCESS_TOKEN should be access token of a system user. Read https://developers.facebook.com/docs/whatsapp/business-management-api/get-started#system-user-access-tokens

YOUR_TEMPLATE_NAME should be a template in "WhatsApp Manager" -> "Message templates". The template category MUST be "Authentication". Create templates in same name for all supported languages. You should list all supported languages in `languages`.

YOUR_VERIFY_TOKEN can be any random string. You can generate one with `openssl rand -hex 16`.

YOUR_APP_SECRET can be found in "App Settings" -> "Basic".

## Configure webhook

- Ensure your facebook app is "Live" (Not "Development").
- In the "WhatsApp" tab, click "Configuration". You should see a "Webhook" section.
- In the "Webhook" section, fill in "Callback URL". It should be `{AUTHGEAR_ENDPOINT}/whatsapp/webhook`.
- In the "Webhook" section, fill in "Verify token". It should be `YOUR_VERIFY_TOKEN` you've configured in the secrets.
- In "Webhook fields", subscribe to "messages".

### Subscribe to the whatsapp business account

This step is required to let your facebook app to receive notifications from your whatsapp business account.

Note, you must be logged in to an account with access to both your facebook app and whatsapp business account to perform these steps:

- Go to https://developers.facebook.com/tools/explorer
- In "Meta App", select your facebook app
- In User or Page, select "User Token"
- In "Add a permission", select:
  - whatsapp_business_management
  - whatsapp_business_messaging
- Press Generate Access Token
- Then, type `{WABA_ID}/subscribed_apps` as the request path. Replace `{WABA_ID}` with your whatsapp business account ID.
- Change the request method to `POST`.
- Press "Submit"
- Change the request method to `GET`, submit and confirm a new item is created. You should an item like this:

```json
{
  "whatsapp_business_api_data": {
    "link": "https://www.facebook.com/games/?app_id=1234567890987654",
    "name": "YOUR APP NAME",
    "id": "1234567890987654"
  }
},
```
