/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from "relay-runtime";
export type UsersScreenQueryVariables = {};
export type UsersScreenQueryResponse = {
    readonly users: {
        readonly edges: ReadonlyArray<{
            readonly node: {
                readonly id: string;
                readonly createdAt: unknown;
            } | null;
        } | null> | null;
    } | null;
};
export type UsersScreenQuery = {
    readonly response: UsersScreenQueryResponse;
    readonly variables: UsersScreenQueryVariables;
};



/*
query UsersScreenQuery {
  users {
    edges {
      node {
        id
        createdAt
      }
    }
  }
}
*/

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "UserConnection",
    "kind": "LinkedField",
    "name": "users",
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
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "UsersScreenQuery",
    "selections": (v0/*: any*/),
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "UsersScreenQuery",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "52b8b27fab19b9abf95c0966cc66f0f8",
    "id": null,
    "metadata": {},
    "name": "UsersScreenQuery",
    "operationKind": "query",
    "text": "query UsersScreenQuery {\n  users {\n    edges {\n      node {\n        id\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();
(node as any).hash = 'e7716396ca7db1a1b827e260585805dd';
export default node;
