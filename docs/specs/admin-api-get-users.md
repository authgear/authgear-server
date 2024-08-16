# Get Users in Admin API

- [Get Users in Admin API](#get-users-in-admin-api)
  - [Introduction](#introduction)
  - [Use Cases](#use-cases)
  - [Design](#design)
    - [Get User queries](#get-user-queries)
      - [Error Response](#error-response)

## Introduction

This document describes the feature for adding three immediately consistent get user APIs for different search criteria. Current `searchKeyword` in list user query is searching data in Elastic Search so it can only return cached result. 

Auth0 reference:
- https://auth0.com/docs/manage-users/user-search/retrieve-users-with-get-users-by-email-endpoint

## Use Cases

- An administrator needs to search users with real time consistency.
- An administrator needs to export list of users data to a file.

## Design

### Get User queries

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

#### Error Response

|Description|Name|Reason|Info|
|---|---|---|---|
|Invalid argument provided.|`Invalid`|`GetUsersInvalidArgument`|Will return detail on any invalid input|

Possible `GetUsersInvalidArgumentType` error message:
- `INVALID_ATTRIBUTE_NAME`: attributeName must be email, phone_number or preferred_name
- `INVALID_LOGIN_ID_KEY`: Invalid Login ID key
- `INVALID_OAUTH_PROVIDER_ALIAS`: Invalid OAuth provider alias

If user is not found, just return null or empty array(for `getUsersByStandardAttribute`). No error in this case.

If it is `attributeName`: "email" `attributeValue`: "nonsense", then just return empty array.
