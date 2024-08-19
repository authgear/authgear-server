- [Get Users in Admin API](#get-users-in-admin-api)
  * [Introduction](#introduction)
  * [Use Cases](#use-cases)
  * [Design](#design)
  * [Error Response](#error-response)⏎

# Get Users in Admin API

## Introduction

This document describes the feature for adding three immediately consistent get user APIs for different search criteria. Current `searchKeyword` in list user query is searching data in Elastic Search so it can only return cached result. 

Auth0 reference:
- https://auth0.com/docs/manage-users/user-search/retrieve-users-with-get-users-by-email-endpoint

## Use Cases

- An administrator needs to search users with real time consistency.
- An administrator needs to export list of users data to a file.

## Design

Get user queries can search users for different auth method. Parameters are all case-sensitive.
- When there are accounts linking `create_new_user`, then an email can be shared by multiple users, thus return an array of `User` for `getUsersByStandardAttribute`.
- `attributeName` must be `email`, `phone_number` or `preferred_username`.

```graphql
type Query {
  # attributeName must be `email`, `phone_number` or `preferred_username`.
  getUsersByStandardAttribute(attributeName: String!, attributeValue: String!): [User!]!
  getUserByLoginID(loginIDKey: String!, loginIDValue: String!): User
  getUserByOAuth(oauthProviderAlias: String!, oauthProviderUserID: String!): User
}
```

- If user is not found, just return null or an empty array (for `getUsersByStandardAttribute`).
- In `getUsersByStandardAttribute`, the `attributeValue` is normalized according to configuration before use. Therefore, it could be invalid.

## Error Response

|Description|Name|Reason|Info|
|---|---|---|---|
|Invalid argument provided.|`Invalid`|`GetUsersInvalidArgument`|-|

|Possible error message of `GetUsersInvalidArgument`|When|
|---|---|
|`attributeName must be email, phone_number or preferred_name`|When `attributeName` is not `email`, `phone_number` or `preferred_name`.|
|`invalid attributeValue`|When `attributeValue` cannot be normalized according to its type.|
|`invalid Login ID key`|When `loginIDKey` is not a configured login ID key in the project.|
|`invalid OAuth provider alias`|When `oauthProviderAlias` is not an alias of a configured OAuth provider in the project.|
