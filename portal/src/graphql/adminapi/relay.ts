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

export function makeEnvironment(graphqlOpaqueAppID: string): Environment {
  function fetchQuery(
    request: RequestParameters,
    variables: Variables
  ): ObservableFromValue<GraphQLResponse> {
    return fetch(
      `/api/apps/${encodeURIComponent(graphqlOpaqueAppID)}/graphql`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          query: request.text,
          variables,
        }),
      }
    ).then(async (response) => {
      // Discard data when errors is present
      // See https://github.com/facebook/relay/issues/1913
      const { data, errors, ...rest } = await response.json();
      if (errors != null) {
        return {
          ...rest,
          errors,
        };
      }
      return {
        ...rest,
        data,
      };
    });
  }

  const source = new RecordSource();
  const store = new Store(source);
  const network = Network.create(fetchQuery);

  const environment = new Environment({
    network,
    store,
  });

  return environment;
}
