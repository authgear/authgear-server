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
      - HTML: [Initialization code with mixpanel.init call](https://developer.mixpanel.com/docs/javascript-quickstart#installation-option-2-html).
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
      - Trigger: `Click With data-authgear-event`
