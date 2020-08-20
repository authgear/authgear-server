import {
  Environment,
  Network,
  RecordSource,
  Store,
  Variables,
  RequestParameters,
  ObservableFromValue,
  GraphQLResponse,
} from "relay-runtime";

function fetchQuery(
  request: RequestParameters,
  variables: Variables
): ObservableFromValue<GraphQLResponse> {
  return fetch("/api/graphql", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      query: request.text,
      variables,
    }),
  }).then(async (response) => {
    return response.json();
  });
}

const source = new RecordSource();
const store = new Store(source);
const network = Network.create(fetchQuery);

const environment = new Environment({
  network,
  store,
});

export { environment };
