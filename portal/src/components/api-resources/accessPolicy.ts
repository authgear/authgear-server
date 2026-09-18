import {
  AccessPolicy,
  AccessPolicyInput,
} from "../../graphql/adminapi/globalTypes.generated";

// One key per client category, see docs/specs/api-resource.md "Access Policy".
type AccessPolicyKey =
  | "allowStaticFirstPartyClientAccess"
  | "allowStaticThirdPartyClientAccess"
  | "allowDynamicFirstPartyClientAccess"
  | "allowDynamicThirdPartyClientAccess";

// The keys the portal offers. The static first-party key is labelled "Client
// applications", after the page those clients are created on. Static
// third-party clients are left out for now: the third_party_app type cannot
// be created there, so the key would have nothing to point at. A value set
// through the Admin API is still shown read-only in the scope list.
export type VisibleAccessPolicyKey = Exclude<
  AccessPolicyKey,
  "allowStaticThirdPartyClientAccess"
>;

// Display order wherever the categories are listed.
export const VISIBLE_ACCESS_POLICY_KEYS: readonly VisibleAccessPolicyKey[] = [
  "allowStaticFirstPartyClientAccess",
  "allowDynamicFirstPartyClientAccess",
  "allowDynamicThirdPartyClientAccess",
];

export type AccessPolicyState = Record<AccessPolicyKey, boolean>;

export const CLOSED_ACCESS_POLICY: AccessPolicyState = {
  allowStaticFirstPartyClientAccess: false,
  allowStaticThirdPartyClientAccess: false,
  allowDynamicFirstPartyClientAccess: false,
  allowDynamicThirdPartyClientAccess: false,
};

export const accessPolicyLabelIDs: Record<VisibleAccessPolicyKey, string> = {
  allowStaticFirstPartyClientAccess: "AccessPolicy.static-first-party.label",
  allowDynamicFirstPartyClientAccess: "AccessPolicy.dynamic-first-party.label",
  allowDynamicThirdPartyClientAccess: "AccessPolicy.dynamic-third-party.label",
};

export const accessPolicyDescriptionIDs: Record<
  VisibleAccessPolicyKey,
  string
> = {
  allowStaticFirstPartyClientAccess:
    "AccessPolicy.static-first-party.description",
  allowDynamicFirstPartyClientAccess:
    "AccessPolicy.dynamic-first-party.description",
  allowDynamicThirdPartyClientAccess:
    "AccessPolicy.dynamic-third-party.description",
};

export const accessPolicyShortLabelIDs: Record<VisibleAccessPolicyKey, string> =
  {
    allowStaticFirstPartyClientAccess: "AccessPolicy.static-first-party.short",
    allowDynamicFirstPartyClientAccess:
      "AccessPolicy.dynamic-first-party.short",
    allowDynamicThirdPartyClientAccess:
      "AccessPolicy.dynamic-third-party.short",
  };

export function accessPolicyStateFromAccessPolicy(
  policy: AccessPolicy
): AccessPolicyState {
  return {
    allowStaticFirstPartyClientAccess: policy.allowStaticFirstPartyClientAccess,
    allowStaticThirdPartyClientAccess: policy.allowStaticThirdPartyClientAccess,
    allowDynamicFirstPartyClientAccess:
      policy.allowDynamicFirstPartyClientAccess,
    allowDynamicThirdPartyClientAccess:
      policy.allowDynamicThirdPartyClientAccess,
  };
}

export function withAccessPolicyKey(
  state: AccessPolicyState,
  key: AccessPolicyKey,
  value: boolean
): AccessPolicyState {
  const next = { ...state };
  next[key] = value;
  return next;
}

// An input that changes one key and leaves the others as they are, which is
// what the Admin API does with omitted fields on update.
export function accessPolicyInputForKey(
  key: AccessPolicyKey,
  value: boolean
): AccessPolicyInput {
  const input: AccessPolicyInput = {};
  input[key] = value;
  return input;
}
