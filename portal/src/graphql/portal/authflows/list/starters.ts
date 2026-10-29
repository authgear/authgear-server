import {
  AuthenticationFlowConfig,
  AuthenticationFlowsConfig,
  PortalAPIAppConfig,
} from "../../../../types";
import { DEFAULT_FLOW_NAME, FlowType, flowListKey } from "../model";
import { CreatableType } from "./rows";

// What New flow creates: one flow, in an array the editor takes as drafts.
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

// passkey only logs in, so its signup_login option names no signup flow.
const LOGIN_ONLY = new Set(["passkey"]);
// Identities that sign in without a separate authenticate step.
const NO_AUTHENTICATOR = new Set(["oauth"]);

// A signup_login flow: the identify step, handing off to the default signup
// and login flows (custom or generated), which the editor can repoint.
function signupLogin(name: string, identification: string): NewFlows {
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
                ? { identification, login_flow: DEFAULT_FLOW_NAME }
                : {
                    identification,
                    login_flow: DEFAULT_FLOW_NAME,
                    signup_flow: DEFAULT_FLOW_NAME,
                  },
            ],
          },
        ],
      },
    },
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
  return signupLogin(name, identification);
}

export type StarterID =
  | "current"
  | "otp"
  | "password"
  | "verified_email_phone"
  | "social"
  | "password_2fa";
export const STARTER_IDS: readonly StarterID[] = [
  "current",
  "otp",
  "password",
  "verified_email_phone",
  "social",
  "password_2fa",
];

export type MissingMethod =
  | "email_login"
  | "phone_login"
  | "email_or_phone"
  | "otp_login"
  | "password"
  | "email_code"
  | "sms_code"
  | "social"
  | "totp";

type Step = NonNullable<AuthenticationFlowConfig["steps"]>[number];

const LOGIN_IDS = ["email", "phone"] as const;
type LoginID = (typeof LOGIN_IDS)[number];
const OTP_OF: Record<LoginID, string> = {
  email: "primary_oob_otp_email",
  phone: "primary_oob_otp_sms",
};
const PASSWORD = "primary_password";
const TOTP = "secondary_totp";

// The step every template identifies in; code steps send to what it took.
const IDENTIFY = "identify";

// The email and phone login IDs the project has, in that order.
function loginIDs(m: ProjectMethods): LoginID[] {
  return LOGIN_IDS.filter((id) => m.identifications.includes(id));
}

function hasOTP(m: ProjectMethods, id: LoginID): boolean {
  return m.primary.includes(OTP_OF[id]);
}

export function starterMissing(
  id: StarterID,
  m: ProjectMethods
): MissingMethod[] {
  const ids = loginIDs(m);
  const out: MissingMethod[] = [];
  switch (id) {
    case "current":
      break;
    case "otp":
      if (!ids.some((i) => hasOTP(m, i))) out.push("otp_login");
      break;
    case "password":
    case "password_2fa":
      if (ids.length === 0) out.push("email_or_phone");
      if (!m.primary.includes(PASSWORD)) out.push("password");
      if (id === "password_2fa" && !m.secondary.includes(TOTP))
        out.push("totp");
      break;
    case "verified_email_phone":
      if (!ids.includes("email")) out.push("email_login");
      if (!ids.includes("phone")) out.push("phone_login");
      if (!hasOTP(m, "email")) out.push("email_code");
      if (!hasOTP(m, "phone")) out.push("sms_code");
      break;
    case "social":
      if (!m.identifications.includes("oauth")) out.push("social");
      break;
  }
  return out;
}

const otpStep = (type: string, id: LoginID, target = IDENTIFY): Step => ({
  type,
  one_of: [{ authentication: OTP_OF[id], target_step: target }],
});
const passwordStep = (type: string): Step => ({
  type,
  one_of: [{ authentication: PASSWORD }],
});
const verify = (target = IDENTIFY): Step => ({
  type: "verify",
  target_step: target,
});
const identify = (
  options: { identification: string; steps?: Step[] }[],
  name = IDENTIFY
): Step => ({ type: "identify", name, one_of: options });

// Signing up with a login ID and a password: verify it, then set one.
function passwordSignup(): Step[] {
  return [verify(), passwordStep("create_authenticator")];
}
// Logging in with a password, or an OTP code where the project has it.
function passwordLogin(m: ProjectMethods, id: LoginID): Step {
  return {
    type: "authenticate",
    one_of: [
      { authentication: PASSWORD },
      ...(hasOTP(m, id)
        ? [{ authentication: OTP_OF[id], target_step: IDENTIFY }]
        : []),
    ],
  };
}
const totpSteps: Step[] = [
  { type: "create_authenticator", one_of: [{ authentication: TOTP }] },
  { type: "view_recovery_code" },
];
const totpLogin: Step = {
  type: "authenticate",
  one_of: [{ authentication: TOTP }, { authentication: "recovery_code" }],
};

// A template's signup and login flows. Only call when starterMissing is
// empty.
function starterPair(
  id: Exclude<StarterID, "current">,
  name: string,
  m: ProjectMethods
): { signup: AuthenticationFlowConfig; login: AuthenticationFlowConfig } {
  const ids = loginIDs(m);
  switch (id) {
    case "otp": {
      const otpIDs = ids.filter((i) => hasOTP(m, i));
      return {
        signup: {
          name,
          steps: [
            identify(
              otpIDs.map((i) => ({
                identification: i,
                steps: [otpStep("create_authenticator", i)],
              }))
            ),
          ],
        },
        login: {
          name,
          steps: [
            identify(
              otpIDs.map((i) => ({
                identification: i,
                steps: [otpStep("authenticate", i)],
              }))
            ),
          ],
        },
      };
    }
    case "password":
    case "password_2fa": {
      const twoFA = id === "password_2fa";
      return {
        signup: {
          name,
          steps: [
            identify(
              ids.map((i) => ({
                identification: i,
                steps: [...passwordSignup(), ...(twoFA ? totpSteps : [])],
              }))
            ),
          ],
        },
        login: {
          name,
          steps: [
            identify(
              ids.map((i) => ({
                identification: i,
                steps: [passwordLogin(m, i), ...(twoFA ? [totpLogin] : [])],
              }))
            ),
          ],
        },
      };
    }
    case "verified_email_phone": {
      const setup = (i: LoginID): Step =>
        identify(
          [
            {
              identification: i,
              steps: [
                verify(`setup_${i}`),
                otpStep("create_authenticator", i, `setup_${i}`),
              ],
            },
          ],
          `setup_${i}`
        );
      return {
        signup: { name, steps: [setup("email"), setup("phone")] },
        login: {
          name,
          steps: [
            identify(
              (["email", "phone"] as const).map((i) => ({
                identification: i,
                steps: [otpStep("authenticate", i)],
              }))
            ),
          ],
        },
      };
    }
    case "social": {
      // Email with a password beside social login, when the project has both.
      const email = ids.includes("email") && m.primary.includes(PASSWORD);
      return {
        signup: {
          name,
          steps: [
            identify([
              { identification: "oauth" },
              ...(email
                ? [{ identification: "email", steps: passwordSignup() }]
                : []),
            ]),
          ],
        },
        login: {
          name,
          steps: [
            identify([
              { identification: "oauth" },
              ...(email
                ? [
                    {
                      identification: "email",
                      steps: [passwordLogin(m, "email")],
                    },
                  ]
                : []),
            ]),
          ],
        },
      };
    }
  }
}

// A signup & login template makes its signup and login flows too, under the
// same name, and hands off to them.
export function starterFlows(
  id: StarterID,
  type: CreatableType,
  name: string,
  generated: AuthenticationFlowsConfig,
  m: ProjectMethods
): NewFlows {
  if (id === "current") {
    // Hand-offs keep naming the default flows.
    return (generated[flowListKey(type)] ?? [])
      .filter((f) => f.name === DEFAULT_FLOW_NAME)
      .map((flow) => ({ type, flow: { ...flow, name } }));
  }
  const { signup, login } = starterPair(id, name, m);
  if (type === "signup") return [{ type, flow: signup }];
  if (type === "login") return [{ type, flow: login }];
  const identifications = (login.steps?.[0].one_of ?? []).map(
    (o) => o.identification ?? ""
  );
  return [
    {
      type: "signup_login",
      flow: {
        name,
        steps: [
          {
            type: "identify",
            one_of: identifications.map((identification) => ({
              identification,
              login_flow: name,
              signup_flow: name,
            })),
          },
        ],
      },
    },
    { type: "signup", flow: signup },
    { type: "login", flow: login },
  ];
}

// What a template's card draws: the steps of signing up and of logging in,
// each a choice of what the user gives.
export type StripItem =
  | "email"
  | "phone"
  | "email_or_phone"
  | "otp"
  | "verify"
  | "password"
  | "social"
  | "authenticator";
export type Strip = StripItem[][];

export const STARTER_STRIPS: Record<
  Exclude<StarterID, "current">,
  { signup: Strip; login: Strip }
> = {
  otp: {
    signup: [["email_or_phone"], ["otp"]],
    login: [["email_or_phone"], ["otp"]],
  },
  password: {
    signup: [["email_or_phone"], ["verify"], ["password"]],
    login: [["email_or_phone"], ["password", "otp"]],
  },
  verified_email_phone: {
    signup: [["email"], ["verify"], ["phone"], ["verify"]],
    login: [["email_or_phone"], ["otp"]],
  },
  social: {
    signup: [["social", "email"], ["password"]],
    login: [["social", "email"], ["password"]],
  },
  password_2fa: {
    signup: [["email_or_phone"], ["password"], ["authenticator"]],
    login: [["email_or_phone"], ["password"], ["authenticator"]],
  },
};
