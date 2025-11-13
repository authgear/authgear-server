### https://authgear.com/claims/user/has_primary_password

#### Why do we need it

The frontend needs a way to determine whether to display a 'Change Password' button.

#### Competitor Reviews

See https://linear.app/authgear/issue/DEV-3027#comment-16467205

#### Decisions

- We considered exposing an array of authenticator objects in blocking event payloads (e.g., in the `oidc.id_token.pre_create` event) to allow app developers to determine if the user has a primary password authenticator within the hook. However, this exposes more information about the user, which might not be necessary.


- Ultimately, we decided to add a simple boolean flag for this specific use case to minimize the exposed information. This approach is also easier for app developers to use.
