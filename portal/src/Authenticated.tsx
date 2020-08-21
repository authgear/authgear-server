import React, { useEffect } from "react";
import { graphql, QueryRenderer } from "react-relay";
import authgear from "@authgear/web";
import { AuthenticatedQueryResponse } from "./__generated__/AuthenticatedQuery.graphql";
import { environment } from "./relay";
import ShowError from "./ShowError";
import ShowLoading from "./ShowLoading";

const query = graphql`
  query AuthenticatedQuery {
    viewer {
      id
    }
  }
`;

type ShowQueryResultProps = AuthenticatedQueryResponse & {
  children?: React.ReactElement;
};

const ShowQueryResult: React.FC<ShowQueryResultProps> = function ShowQueryResult(
  props: ShowQueryResultProps
) {
  const { viewer } = props;
  const redirectURI = window.location.origin + "/";

  useEffect(() => {
    if (viewer == null) {
      // Normally we should call endAuthorization after being redirected back to here.
      // But we know that we are first party app and are using response_type=none so
      // we can skip that.
      authgear
        .startAuthorization({
          redirectURI,
          prompt: "login",
        })
        .catch((err) => {
          console.error(err);
        });
    }
  }, [viewer, redirectURI]);

  if (viewer != null) {
    return props.children ?? null;
  }

  return null;
};

interface Empty {}

interface Props {
  children?: React.ReactElement;
}

// CAVEAT: <Authenticated><Route path="/foobar/:id"/></Authenticated> will cause useParams to return empty object :(
const Authenticated: React.FC<Props> = function Authenticated(ownProps: Props) {
  return (
    <QueryRenderer<{ variables: Empty; response: AuthenticatedQueryResponse }>
      environment={environment}
      query={query}
      variables={{}}
      render={({ error, props }) => {
        if (error != null) {
          return <ShowError error={error} />;
        }
        if (props == null) {
          return <ShowLoading />;
        }
        return <ShowQueryResult {...props} {...ownProps} />;
      }}
    />
  );
};

export default Authenticated;
