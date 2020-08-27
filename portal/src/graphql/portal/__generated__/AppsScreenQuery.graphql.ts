/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from "relay-runtime";
export type AppsScreenQueryVariables = {};
export type AppsScreenQueryResponse = {
    readonly apps: {
        readonly edges: ReadonlyArray<{
            readonly node: {
                readonly id: string;
            } | null;
        } | null> | null;
    } | null;
};
export type AppsScreenQuery = {
    readonly response: AppsScreenQueryResponse;
    readonly variables: AppsScreenQueryVariables;
};



/*
query AppsScreenQuery {
  apps {
    edges {
      node {
        id
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
    "concreteType": "AppConnection",
    "kind": "LinkedField",
    "name": "apps",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "AppEdge",
        "kind": "LinkedField",
        "name": "edges",
        "plural": true,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "App",
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
    "name": "AppsScreenQuery",
    "selections": (v0/*: any*/),
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "AppsScreenQuery",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "6fe0d6d01f3dcbd7297d68e582ee10b3",
    "id": null,
    "metadata": {},
    "name": "AppsScreenQuery",
    "operationKind": "query",
    "text": "query AppsScreenQuery {\n  apps {\n    edges {\n      node {\n        id\n      }\n    }\n  }\n}\n"
  }
};
})();
(node as any).hash = 'dfafca341141bddf41ff3398872cd951';
export default node;
