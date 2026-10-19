import {
  AuthenticationFlowConfig,
  AuthenticationFlowsConfig,
  PortalAPIAppConfig,
} from "../../../../types";
import { DEFAULT_FLOW_NAME, FlowType, flowListKey } from "../model";
import { CreatableType } from "./rows";
import { copyAs } from "./actions";

export type NewFlows = { type: FlowType; flow: AuthenticationFlowConfig }[];

export interface ProjectMethods {
  identifications: string[];
  primary: string[];
  secondary: string[];
}

export function projectMethods(config: PortalAPIAppConfig): ProjectMethods {
  const identities = config.authentication?.identities ?? [];
  const loginIDs = identities.includes("login_id")
    ? (config.identity?.login_id?.keys ?? []).map((k) => k.type)
    : [];
  return {
    identifications: [
      ...loginIDs,
      ...identities.filter((i) => i === "oauth" || i === "passkey"),
    ],
    primary: (config.authentication?.primary_authenticators ?? []).map(
      (a) => `primary_${a}`
    ),
    secondary: (config.authentication?.secondary_authenticators ?? []).map(
      (a) => `secondary_${a}`
    ),
  };
}

// passkey only logs in, so a bundle's passkey option names no signup flow.
const LOGIN_ONLY = new Set(["passkey"]);
// Identities that sign in without a separate authenticate step.
const NO_AUTHENTICATOR = new Set(["oauth"]);

function bundle(
  name: string,
  identification: string,
  signup: AuthenticationFlowConfig,
  login: AuthenticationFlowConfig
): NewFlows {
  return [
    {
      type: "signup_login",
      flow: {
        name,
        steps: [
          {
            type: "identify",
            one_of: [
              LOGIN_ONLY.has(identification)
                ? { identification, login_flow: name }
                : { identification, login_flow: name, signup_flow: name },
            ],
          },
        ],
      },
    },
    { type: "signup", flow: signup },
    { type: "login", flow: login },
  ];
}

// The identification a from-scratch flow starts with: the first login ID,
// else social login. null when the project has neither.
function scratchIdentification(methods: ProjectMethods): string | null {
  return (
    methods.identifications.find(
      (i) => !LOGIN_ONLY.has(i) && !NO_AUTHENTICATOR.has(i)
    ) ??
    methods.identifications.find((i) => NO_AUTHENTICATOR.has(i)) ??
    null
  );
}

export function scratchMissing(methods: ProjectMethods): boolean {
  return scratchIdentification(methods) == null;
}

// Only call when scratchMissing is false.
export function scratchFlows(
  type: CreatableType,
  name: string,
  methods: ProjectMethods
): NewFlows {
  const identification = scratchIdentification(methods);
  if (identification == null) {
    throw new Error("scratchFlows: the project has no login method");
  }
  const authentication = NO_AUTHENTICATOR.has(identification)
    ? undefined
    : methods.primary[0];
  const signup: AuthenticationFlowConfig = {
    name,
    steps: [{ type: "identify", one_of: [{ identification }] }],
  };
  const login: AuthenticationFlowConfig = {
    name,
    steps: [
      { type: "identify", one_of: [{ identification }] },
      ...(authentication != null
        ? [{ type: "authenticate", one_of: [{ authentication }] }]
        : []),
    ],
  };
  if (type === "signup") return [{ type, flow: signup }];
  if (type === "login") return [{ type, flow: login }];
  return bundle(name, identification, signup, login);
}

export type StarterID =
  | "current"
  | "email_password"
  | "email_code"
  | "phone_code"
  | "email_password_2fa";
export const STARTER_IDS: readonly StarterID[] = [
  "current",
  "email_password",
  "email_code",
  "phone_code",
  "email_password_2fa",
];

export type MissingMethod =
  | "email_login"
  | "phone_login"
  | "password"
  | "email_code"
  | "sms_code"
  | "totp";

const needs: Record<
  StarterID,
  { identification?: string; primary?: string; secondary?: string }
> = {
  current: {},
  email_password: { identification: "email", primary: "primary_password" },
  email_code: { identification: "email", primary: "primary_oob_otp_email" },
  phone_code: { identification: "phone", primary: "primary_oob_otp_sms" },
  email_password_2fa: {
    identification: "email",
    primary: "primary_password",
    secondary: "secondary_totp",
  },
};

const missingOf: Record<string, MissingMethod> = {
  email: "email_login",
  phone: "phone_login",
  primary_password: "password",
  primary_oob_otp_email: "email_code",
  primary_oob_otp_sms: "sms_code",
  secondary_totp: "totp",
};

export function starterMissing(
  id: StarterID,
  methods: ProjectMethods
): MissingMethod[] {
  const n = needs[id];
  const out: MissingMethod[] = [];
  if (
    n.identification != null &&
    !methods.identifications.includes(n.identification)
  )
    out.push(missingOf[n.identification]);
  if (n.primary != null && !methods.primary.includes(n.primary))
    out.push(missingOf[n.primary]);
  if (n.secondary != null && !methods.secondary.includes(n.secondary))
    out.push(missingOf[n.secondary]);
  return out;
}

// The signup and login flows of a starter.
function starterPair(
  id: Exclude<StarterID, "current">,
  name: string
): {
  identification: string;
  signup: AuthenticationFlowConfig;
  login: AuthenticationFlowConfig;
} {
  const identify = (
    identification: string,
    steps: AuthenticationFlowConfig["steps"]
  ) => ({
    type: "identify",
    name: "identify",
    one_of: [{ identification, steps }],
  });
  switch (id) {
    case "email_password":
    case "email_password_2fa": {
      const twoFA = id === "email_password_2fa";
      return {
        identification: "email",
        signup: {
          name,
          steps: [
            identify("email", [
              { type: "verify", target_step: "identify" },
              {
                type: "create_authenticator",
                one_of: [{ authentication: "primary_password" }],
              },
              ...(twoFA
                ? [
                    {
                      type: "create_authenticator",
                      one_of: [{ authentication: "secondary_totp" }],
                    },
                    { type: "view_recovery_code" },
                  ]
                : []),
            ]),
          ],
        },
        login: {
          name,
          steps: [
            identify("email", [
              {
                type: "authenticate",
                one_of: [{ authentication: "primary_password" }],
              },
              ...(twoFA
                ? [
                    {
                      type: "authenticate",
                      one_of: [
                        { authentication: "secondary_totp" },
                        { authentication: "recovery_code" },
                      ],
                    },
                  ]
                : []),
            ]),
          ],
        },
      };
    }
    case "email_code":
    case "phone_code": {
      const identification = id === "email_code" ? "email" : "phone";
      const authentication =
        id === "email_code" ? "primary_oob_otp_email" : "primary_oob_otp_sms";
      const codeStep = (type: string) => ({
        type,
        one_of: [{ authentication, target_step: "identify" }],
      });
      return {
        identification,
        signup: {
          name,
          steps: [identify(identification, [codeStep("create_authenticator")])],
        },
        login: {
          name,
          steps: [identify(identification, [codeStep("authenticate")])],
        },
      };
    }
  }
}

export function starterFlows(
  id: StarterID,
  type: CreatableType,
  name: string,
  generated: AuthenticationFlowsConfig
): NewFlows {
  if (id === "current") {
    const types: FlowType[] =
      type === "signup_login" ? ["signup_login", "signup", "login"] : [type];
    const sources = types.flatMap((t) =>
      (generated[flowListKey(t)] ?? [])
        .filter((f) => f.name === DEFAULT_FLOW_NAME)
        .map((flow) => ({ type: t, flow }))
    );
    return copyAs(sources, name);
  }
  const { identification, signup, login } = starterPair(id, name);
  if (type === "signup") return [{ type, flow: signup }];
  if (type === "login") return [{ type, flow: login }];
  return bundle(name, identification, signup, login);
}

export function stepSummary(flow: AuthenticationFlowConfig): string[] {
  const walk = (steps: AuthenticationFlowConfig["steps"]): string[] =>
    (steps ?? []).flatMap((s) => [s.type, ...walk(s.one_of?.[0]?.steps)]);
  return walk(flow.steps);
}
