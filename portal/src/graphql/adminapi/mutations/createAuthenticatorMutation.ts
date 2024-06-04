import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  CreateAuthenticatorMutationMutation,
  CreateAuthenticatorMutationDocument,
  CreateAuthenticatorMutationMutationVariables,
} from "./createAuthenticatorMutation.generated";
import { AuthenticatorKind } from "../globalTypes.generated";

interface AuthenticatorDefinitionOOBOTPEmail {
  type: "oob_otp_email";
  kind: AuthenticatorKind;
  email: string;
}

interface AuthenticatorDefinitionOOBOTPSMS {
  type: "oob_otp_sms";
  kind: AuthenticatorKind;
  phone: string;
}

interface AuthenticatorDefinitionPassword {
  type: "oob_otp_password";
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
        {};
      switch (definitionParam.type) {
        case "oob_otp_email":
          definition.oobOtpEmail = {
            kind: definitionParam.kind,
            email: definitionParam.email,
          };
          break;
        case "oob_otp_sms":
          definition.oobOtpSMS = {
            kind: definitionParam.kind,
            phone: definitionParam.phone,
          };
          break;
        case "oob_otp_password":
          definition.password = {
            kind: definitionParam.kind,
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
