# Portal Google Tag Manager Setup

This document describes how to setup portal tracking to send data to Mixpanel via Google Tag Manager.

1. Create Mixpanel project and obtain project token from settings page
1. Create following resources in Google Tag Manager
  - Triggers
    - `Click With data-authgear-event`
      - Type: Click - All Elements
      - Fires on: Some Clicks
      - Conditions:
        - Click Elements, matches CSS selector, `[data-authgear-event]`
    - `Authgear Custom Event`
      - Type: Custom Event
      - Event name: `authgear.*`
      - Use regex matching: true
      - Fires on: All Custom Events
  - User-Defined Variables
    - `gtm.element.dataset.authgearEvent`
      - Type: Data Layer Variables
    - `gtm.element.dataset.authgearEventValue1`
      - Type: Data Layer Variables
    - `gtm.element.dataset.authgearEventAppId`
      - Type: Data Layer Variables
    - `appID`
      - Type: Data Layer Variables
    - `value1`
      - Type: Data Layer Variables
  - Tags
    - `Mixpanel Initialize`
      - Type: Custom HTML
      - HTML: [Initialization code with mixpanel.init call](https://developer.mixpanel.com/docs/javascript-quickstart#installation-option-2-html).
      - Trigger: All Pages
    - `Mixpanel Track Click`
      - Type: Custom HTML
      - HTML:
          ```
          <script type="text/javascript">
            var event = {{gtm.element.dataset.authgearEvent}};
            var value1 = {{gtm.element.dataset.authgearEventValue1}};
            mixpanel.track(event, {value1: value1});
          </script>
          ```
      - Trigger: `Click With data-authgear-event`
    - `Mixpanel Track Custom Event`
      - Type: Custom HTML
      - HTML:
          ```
          <script type="text/javascript">
            var value1 = {{value1}};
            var appID = {{appID}};
            mixpanel.track(event, {appID: appID, value1: value1});
          </script>
          ```
      - Trigger: `Click With data-authgear-event`
