import { HookEvent } from "https://deno.land/x/authgear_deno_hook@v1.3.0/mod.ts";

// Replace the endpoint if necessary.
const POSTHOG_ENDPOINT = "https://eu.posthog.com";
// Fill in your project API key.
const POSTHOG_API_KEY = "";

interface Client {
  client_id: string;
  x_application_type: string;
}

function findNewClients(oldClients: Client[], newClients: Client[]): Client[] {
  const oldClientIDSet = new Set<string>();
  for (const c of oldClients) {
    oldClientIDSet.add(c.client_id);
  }

  const newClientIDSet = new Set<string>();
  for (const c of newClients) {
    newClientIDSet.add(c.client_id);
  }

  const diff = new Set<string>();
  for (const client_id of newClientIDSet) {
    if (!oldClientIDSet.has(client_id)) {
      diff.add(client_id);
    }
  }

  const output: Client[] = [];
  for (const c of newClients) {
    if (diff.has(c.client_id)) {
      output.push(c);
    }
  }

  return output;
}

async function identify(options: {
  user_id: string;
  ip: string | undefined;
  email: string | undefined;
}): Promise<void> {
  // https://posthog.com/docs/api/post-only-endpoints
  const url = new URL("/capture", POSTHOG_ENDPOINT);

  const $set: any = {};
  if (options.email != null) {
    $set.email = options.email;
  }

  const properties: any = {
    $set,
  };
  if (options.ip != null) {
    properties["$ip"] = options.ip;
  } else {
    properties["$geoip_disable"] = true;
  }

  let body: any = {
    api_key: POSTHOG_API_KEY,
    // https://posthog.com/docs/api/post-only-endpoints#identify
    event: "$identify",
    distinct_id: options.user_id,
    properties,
  };
  const resp = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  const text = await resp.text();
  if (resp.status !== 200) {
    throw new Error(text);
  }
}

async function groupIdentify(options: {
  project_id: string;
  ip: string | undefined;
}): Promise<void> {
  // https://posthog.com/docs/api/post-only-endpoints
  const url = new URL("/capture", POSTHOG_ENDPOINT);

  const properties: any = {
    $group_type: "project",
    $group_key: options.project_id,
  };
  if (options.ip != null) {
    properties["$ip"] = options.ip;
  } else {
    properties["$geoip_disable"] = true;
  }

  let body: any = {
    api_key: POSTHOG_API_KEY,
    // https://posthog.com/docs/api/post-only-endpoints#group-identify
    event: "$groupidentify",
    distinct_id: "groups_setup_id",
    properties,
  };
  const resp = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  const text = await resp.text();
  if (resp.status !== 200) {
    throw new Error(text);
  }
}

async function capture(options: {
  user_id: string;
  event: string;
  ip: string | undefined;
  additionalProperties?: Record<string, unknown>;
}): Promise<void> {
  // https://posthog.com/docs/api/post-only-endpoints
  const url = new URL("/capture", POSTHOG_ENDPOINT);

  let properties: any = {};
  if (options.ip != null) {
    properties["$ip"] = options.ip;
  } else {
    properties["$geoip_disable"] = true;
  }
  if (options.additionalProperties != null) {
    properties = {
      ...properties,
      ...options.additionalProperties,
    };
  }

  const body: any = {
    api_key: POSTHOG_API_KEY,
    event: options.event,
    distinct_id: options.user_id,
    properties,
  };

  const resp = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  const text = await resp.text();
  if (resp.status !== 200) {
    throw new Error(text);
  }
}

// e is of type any because the type definition does not cover project events.
export default async function (e: any): Promise<void> {
  switch (e.type) {
    case "user.created": {
      const user_id = e.context.user_id;
      const ip = e.context.ip_address;
      const email = e.payload.user?.standard_attributes?.email;
      // We must call identify here so that even tracking is blocked in the browser,
      // we can still identify the user.
      await identify({
        ip,
        user_id,
        email,
      });
      await capture({
        ip,
        user_id,
        event: "signedUp",
      });
      break;
    }
    case "project.app.created": {
      const ip = e.context.ip_address;
      const actor_user_id = e.context.audit_context.actor_user_id;
      const projectID = e.context.app_id;
      await groupIdentify({
        project_id: projectID,
        ip,
      });
      await capture({
        ip,
        user_id: actor_user_id,
        event: "createdProject",
        additionalProperties: {
          project: projectID,
        },
      });
      break;
    }
    case "project.app.updated": {
      const ip = e.context.ip_address;
      const actor_user_id = e.context.audit_context.actor_user_id;
      const projectID = e.context.app_id;
      const { app_config_old, app_config_new } = e.payload;
      const clients_old: Client[] = app_config_old.oauth?.clients ?? [];
      const clients_new: Client[] = app_config_new.oauth?.clients ?? [];
      const newClients = findNewClients(clients_old, clients_new);
      for (const c of newClients) {
        await groupIdentify({
          project_id: projectID,
          ip,
        });
        await capture({
          ip,
          user_id: actor_user_id,
          event: "createdApplication",
          additionalProperties: {
            projectID,
            applicationType: c.x_application_type,
            $groups: {
              project: projectID,
            },
          },
        });
      }
      break;
    }
    case "project.collaborator.invitation.accepted": {
      const ip = e.context.ip_address;
      const actor_user_id = e.context.audit_context.actor_user_id;
      const projectID = e.context.app_id;
      await groupIdentify({
        project_id: projectID,
        ip,
      });
      await capture({
        ip,
        user_id: actor_user_id,
        event: "acceptedInvite",
        additionalProperties: {
          projectID,
          $groups: {
            project: projectID,
          },
        },
      });
      break;
    }
  }
}
