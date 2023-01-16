# "Back to App" redirection

This documeent describes configuration of "Back to App" redirection.

- [Configuration](#configuration)
- [Screens with "Back to App"](#screens-with-back-to-app)
  - [Magic Link verification](#magic-link-verification)
  - [Reset password](#reset-password)
  - ["Return to where you were" screen](#return-to-where-you-were-screen)


## Configuration

In certain cases (described below), Authgear may display a "Back to App" link for navigating back to client's app.

To enable, include following section in `authgear.yaml`:

```yaml
ui:
  back_to_app:
    uri: http://portal.localhost:8000/
```

To only display link on mobile, update `authgear.yaml` as follow:

```yaml
ui:
  back_to_app:
    uri: http://portal.localhost:8000/
    display_mode: mobile
```

Desktop user will be prompted to close the window instead.

## Screens with "Back to App"

### Magic Link verification

When an application initiates Magic Link authentication, it sends a link that initiates the passwordless verification process.

This process can happen in another device / session, but only the original session will be granted access and can proceeed to next step.

The other device will reach deadend screen, by default end-user will be prompted to close the window.

### Reset password

Both forgot password and reset password reaches deadend screen upon success. By default end-user will only see a success message.

### "Return to where you were" screen

This screen is specifically for links that should not be opened on current device. By default end-user will be prompted to return to app manually.
