import React, { useEffect } from "react";
import { gql, useQuery } from "@apollo/client";
import authgear from "@authgear/web";
import { AuthenticatedQuery } from "./__generated__/AuthenticatedQuery";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";

const query = gql`
  query AuthenticatedQuery {
    viewer {
      id
    }
  }
`;

interface ShowQueryResultProps extends AuthenticatedQuery {
  children?: React.ReactElement;
}

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

interface Props {
  children?: React.ReactElement;
}

// CAVEAT: <Authenticated><Route path="/foobar/:id"/></Authenticated> will cause useParams to return empty object :(
const Authenticated: React.FC<Props> = function Authenticated(ownProps: Props) {
  const { loading, error, data, refetch } = useQuery<AuthenticatedQuery>(query);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return <ShowQueryResult viewer={data?.viewer ?? null} {...ownProps} />;
};

export default Authenticated;
