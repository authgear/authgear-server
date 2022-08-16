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
    - `ag.event.*`
      - Type: Custom Event
      - Event name: `ag.event.*`
      - Use regex matching: true
      - Fires on: All Custom Events
    - `ag.lifecycle.identified`
      - Type: Custom Event
      - Event name: `ag.lifecycle.identified`
      - Use regex matching: true
      - Fires on: All Custom Events
  - Built-In Variables
    - Select all
  - User-Defined Variables
    - `gtm.element.dataset`
      - Type: Data Layer Variables
    - `app_id`
      - Type: Data Layer Variables
    - `event_data`
      - Type: Data Layer Variables
  - Tags
    - `Mixpanel Initialize`
      - Type: Custom HTML
      - HTML: Use the following HTML script tag snippet, replace the mixpanel init script based on the [doc](https://developer.mixpanel.com/docs/javascript-quickstart#installation-option-2-html) and replace the `YOUR_PROJECT_TOKEN`.
        ```html
        <script type="text/javascript">
          // REPLACE WITH MIXPANEL INIT SCRIPT

          window._agMixpanelIdentifyIfNeeded = function (mixpanel) {
            if (window._agUserData == null) {
              return;
            }
            try {
              var distinctID = mixpanel.get_distinct_id();
              if (distinctID === window._agUserData.user_id) {
                return;
              }
              mixpanel.reset();
              mixpanel.identify(window._agUserData.user_id);
              if (window._agUserData.email) {
                mixpanel.people.set({ $email: window._agUserData.email });
              }
            } catch (_) {}
          };

          // REPLACE THE MIXPANEL PROJECT TOKEN
          mixpanel.init("YOUR_PROJECT_TOKEN", {
            debug: false,
            loaded: function (mixpanel) {
              window._agMixpanelIdentifyIfNeeded(mixpanel);
            },
          });
        </script>
        ```
      - Trigger: All Pages
    - `Mixpanel Track Click`
      - Type: Custom HTML
      - HTML:
        ```html
        <script type="text/javascript">
          function toUnderscore(s) {
            return s.replace(/\.?([A-Z])/g, function (x,y){return "_" + y.toLowerCase()}).replace(/^_/, "");
          }
          var AUTHGEAR_EVENT = "authgearEvent";
          var AUTHGEAR_EVENT_DATA = "authgearEventData";
          var dataset = {{gtm.element.dataset}};
          if (AUTHGEAR_EVENT in dataset) {
            var event = dataset.authgearEvent;
            var eventData = {};
            Object.keys(dataset).map(function(k) {
              if (k.startsWith(AUTHGEAR_EVENT_DATA)) {
                var eventDataKey = toUnderscore(k.replace(AUTHGEAR_EVENT_DATA, ""));
                eventData[eventDataKey] = dataset[k];
              }
            });
            mixpanel.track(event, eventData);
          }
        </script>
        ```
      - Trigger: `Click With data-authgear-event`
    - `Mixpanel Track Custom Event`
      - Type: Custom HTML
      - HTML:
        ```html
        <script type="text/javascript">
          var appID = {{app_id}};
          var eventData = {{event_data}};
          mixpanel.track({{Event}}, Object.assign({app_id: appID}, eventData));
        </script>
        ```
      - Trigger: `ag.event.*`
    - `Mixpanel Identify`
      - Type: Custom HTML
      - HTML:
        ```html
        <script type="text/javascript">
          var event = {{Event}};
          if (event === "ag.lifecycle.identified") {
            window._agUserData = {{event_data}};
            window._agMixpanelIdentifyIfNeeded(mixpanel);
          }
        </script>
        ```
      - Trigger: `ag.lifecycle.identified`
