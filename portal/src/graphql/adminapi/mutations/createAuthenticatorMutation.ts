import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  CreateAuthenticatorMutationMutation,
  CreateAuthenticatorMutationDocument,
  CreateAuthenticatorMutationMutationVariables,
} from "./createAuthenticatorMutation.generated";
import { AuthenticatorKind, AuthenticatorType } from "../globalTypes.generated";

interface AuthenticatorDefinitionOOBOTPEmail {
  type: AuthenticatorType.OobOtpEmail;
  kind: AuthenticatorKind;
  email: string;
}

interface AuthenticatorDefinitionOOBOTPSMS {
  type: AuthenticatorType.OobOtpSms;
  kind: AuthenticatorKind;
  phone: string;
}

interface AuthenticatorDefinitionPassword {
  type: AuthenticatorType.Password;
  kind: AuthenticatorKind;
  password: string;
}

type AuthenticatorDefinition =
  | AuthenticatorDefinitionOOBOTPEmail
  | AuthenticatorDefinitionOOBOTPSMS
  | AuthenticatorDefinitionPassword;

interface Authenticator {
  id: string;
}

export type CreateAuthenticatorFunction = (
  definition: AuthenticatorDefinition
) => Promise<Authenticator | undefined>;

export function useCreateAuthenticatorMutation(userID: string): {
  createAuthenticator: CreateAuthenticatorFunction;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    CreateAuthenticatorMutationMutation,
    CreateAuthenticatorMutationMutationVariables
  >(CreateAuthenticatorMutationDocument);

  const createAuthenticator = useCallback(
    async (definitionParam: AuthenticatorDefinition) => {
      const definition: CreateAuthenticatorMutationMutationVariables["definition"] =
        {
          kind: definitionParam.kind,
          type: definitionParam.type,
        };
      switch (definitionParam.type) {
        case AuthenticatorType.OobOtpEmail:
          definition.oobOtpEmail = {
            email: definitionParam.email,
          };
          break;
        case AuthenticatorType.OobOtpSms:
          definition.oobOtpSMS = {
            phone: definitionParam.phone,
          };
          break;
        case AuthenticatorType.Password:
          definition.password = {
            password: definitionParam.password,
          };
          break;
        default:
          throw new Error("unknown AuthenticatorDefinition type");
      }

      const result = await mutationFunction({
        variables: {
          userID,
          definition,
        },
      });

      return result.data?.createAuthenticator.authenticator;
    },
    [mutationFunction, userID]
  );
  return { createAuthenticator, error, loading };
}
