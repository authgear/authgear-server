/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from "relay-runtime";
import { FragmentRefs } from "relay-runtime";
export type UsersList_users = {
    readonly users: {
        readonly edges: ReadonlyArray<{
            readonly node: {
                readonly id: string;
                readonly createdAt: unknown;
            } | null;
        } | null> | null;
        readonly pageInfo: {
            readonly hasNextPage: boolean;
            readonly endCursor: string | null;
        };
        readonly totalCount: number | null;
    } | null;
    readonly " $refType": "UsersList_users";
};
export type UsersList_users$data = UsersList_users;
export type UsersList_users$key = {
    readonly " $data"?: UsersList_users$data;
    readonly " $fragmentRefs": FragmentRefs<"UsersList_users">;
};



const node: ReaderFragment = {
  "argumentDefinitions": [
    {
      "kind": "RootArgument",
      "name": "cursor"
    },
    {
      "kind": "RootArgument",
      "name": "pageSize"
    }
  ],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      {
        "count": "pageSize",
        "cursor": "cursor",
        "direction": "forward",
        "path": [
          "users"
        ]
      }
    ]
  },
  "name": "UsersList_users",
  "selections": [
    {
      "alias": "users",
      "args": null,
      "concreteType": "UserConnection",
      "kind": "LinkedField",
      "name": "__UsersList_users_connection",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "UserEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "User",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "id",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "createdAt",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "__typename",
                  "storageKey": null
                }
              ],
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "cursor",
              "storageKey": null
            }
          ],
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "concreteType": "PageInfo",
          "kind": "LinkedField",
          "name": "pageInfo",
          "plural": false,
          "selections": [
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "hasNextPage",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "endCursor",
              "storageKey": null
            }
          ],
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "totalCount",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Query",
  "abstractKey": null
};
(node as any).hash = '369b7c8922757526dfbb1a4c81dfea5e';
export default node;
