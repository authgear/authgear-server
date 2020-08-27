/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from "relay-runtime";
export type AuthenticatedQueryVariables = {};
export type AuthenticatedQueryResponse = {
    readonly viewer: {
        readonly id: string;
    } | null;
};
export type AuthenticatedQuery = {
    readonly response: AuthenticatedQueryResponse;
    readonly variables: AuthenticatedQueryVariables;
};



/*
query AuthenticatedQuery {
  viewer {
    id
  }
}
*/

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "User",
    "kind": "LinkedField",
    "name": "viewer",
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
];
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "AuthenticatedQuery",
    "selections": (v0/*: any*/),
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "AuthenticatedQuery",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "f733183efba3dcb614f30d2cfff08ba1",
    "id": null,
    "metadata": {},
    "name": "AuthenticatedQuery",
    "operationKind": "query",
    "text": "query AuthenticatedQuery {\n  viewer {\n    id\n  }\n}\n"
  }
};
})();
(node as any).hash = '2e4de737e9bd88ffcb12e77fb1138060';
export default node;
