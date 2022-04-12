import { gql } from "@apollo/client";

export default gql`
  query ScreenNavQuery($id: ID!) {
    node(id: $id) {
      __typename
      ... on App {
        id
        effectiveFeatureConfig
        planName
        tutorialStatus {
          appID
          data
        }
      }
    }
  }
`;
