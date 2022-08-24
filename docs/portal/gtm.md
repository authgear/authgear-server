# Portal Google Tag Manager Setup

This document describes how to setup portal tracking to send data to Mixpanel via Google Tag Manager.

## Setup

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
    - `History Change`
      - Type: History Change
      - Fries on: All History Changes
  - Built-In Variables
    - Select all
  - User-Defined Variables
    - `gtm.element.dataset`
      - Type: Data Layer Variables
    - `app_context`
      - Type: Data Layer Variables
    - `event_data`
      - Type: Data Layer Variables
  - Tags
    - `Mixpanel Initialize`
      - Type: Custom HTML
      - HTML: Use mixpanel installation script based on the [doc](https://developer.mixpanel.com/docs/javascript-quickstart#installation-option-2-html). DO NOT need to call `mixpanel.init` in this stage.
      - Trigger: Initialization - All Pages
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
          var event = {{Event}};
          if (event.startsWith("ag.event.")) {
            var appContext = {{app_context}};
            var eventData = {{event_data}};
            mixpanel.track({{Event}}, Object.assign({}, appContext, eventData));
          }
        </script>
        ```
      - Trigger: `ag.event.*`
    - `Mixpanel Identify`
      - Type: Custom HTML
      - HTML: Use the following HTML script tag snippet and replace YOUR_PROJECT_TOKEN.
        ```html
        <script type="text/javascript">
          var event = {{Event}};
          if (event === "ag.lifecycle.identified") {
            var agUserData = {{event_data}};

            // REPLACE THE MIXPANEL PROJECT TOKEN
            mixpanel.init("YOUR_PROJECT_TOKEN", {
              debug: false,
              cookie_domain: window.location.hostname,
              loaded: function (mixpanel) {
                var distinctID = mixpanel.get_distinct_id();
                if (distinctID !== agUserData.user_id) {
                  mixpanel.reset();
                  mixpanel.identify(agUserData.user_id);
                  if (agUserData.email) {
                    mixpanel.people.set({ $email: agUserData.email });
                  }
                }
                mixpanel.track("Pageview", { page_path: {{Page Path}} });
              },
            });
          }
        </script>
        ```
      - Trigger: `ag.lifecycle.identified`
    - `Mixpanel Track Pageview`
      - Type: Custom HTML
      - HTML:
        ```html
        <script type="text/javascript">
          var event = {{Event}};
          if (event === "gtm.historyChange" && mixpanel.track) {
            mixpanel.track("Pageview", { page_path: {{Page Path}} });
          }
        </script>
        ```
      - Trigger: History Change

## Implementation Details

Initial page view is tracked after `mixpanel.identify` to ensure the events are
associated with the correct user.
