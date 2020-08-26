/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from "relay-runtime";
export type AppScreenQueryVariables = {
    id: string;
};
export type AppScreenQueryResponse = {
    readonly node: {
        readonly id?: string;
        readonly appConfig?: unknown;
        readonly secretConfig?: unknown;
    } | null;
};
export type AppScreenQuery = {
    readonly response: AppScreenQueryResponse;
    readonly variables: AppScreenQueryVariables;
};



/*
query AppScreenQuery(
  $id: ID!
) {
  node(id: $id) {
    __typename
    ... on App {
      id
      appConfig
      secretConfig
    }
    id
  }
}
*/

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "id"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "id"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "appConfig",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "secretConfig",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "AppScreenQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/)
            ],
            "type": "App",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AppScreenQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/)
            ],
            "type": "App",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "4e6ce509a5f6209460a7a79a9194bc3d",
    "id": null,
    "metadata": {},
    "name": "AppScreenQuery",
    "operationKind": "query",
    "text": "query AppScreenQuery(\n  $id: ID!\n) {\n  node(id: $id) {\n    __typename\n    ... on App {\n      id\n      appConfig\n      secretConfig\n    }\n    id\n  }\n}\n"
  }
};
})();
(node as any).hash = '58193eaf0b66d536dd0dc61d993692ba';
export default node;
