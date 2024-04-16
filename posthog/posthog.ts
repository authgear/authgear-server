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

async function groupIdentify(project_id: string): Promise<void> {
  // https://posthog.com/docs/api/post-only-endpoints
  const url = new URL("/capture", POSTHOG_ENDPOINT);
  let body: any = {
    api_key: POSTHOG_API_KEY,
    // https://posthog.com/docs/api/post-only-endpoints#group-identify
    event: "$groupidentify",
    distinct_id: "groups_setup_id",
    properties: {
      $group_type: "project",
      $group_key: project_id,
    },
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

async function capture(
  distinct_id: string,
  event: string,
  properties?: Record<string, unknown>,
): Promise<void> {
  // https://posthog.com/docs/api/post-only-endpoints
  const url = new URL("/capture", POSTHOG_ENDPOINT);
  let body: any = {
    api_key: POSTHOG_API_KEY,
    event,
    distinct_id,
  };
  if (properties != null) {
    body.properties = properties;
  }
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

export default async function (e: HookEvent): Promise<void> {
  switch (e.type) {
    case "user.created": {
      const user_id = e.context.user_id;
      await capture(user_id, "signedUp");
      break;
    }
    case "project.app.created": {
      const actor_user_id = e.context.audit_context.actor_user_id;
      const projectID = e.context.app_id;
      await groupIdentify(projectID);
      await capture(actor_user_id, "createdProject", {
        projectID,
        $groups: {
          project: projectID,
        },
      });
      break;
    }
    case "project.app.updated": {
      const actor_user_id = e.context.audit_context.actor_user_id;
      const projectID = e.context.app_id;
      const { app_config_old, app_config_new } = e.payload;
      const clients_old: Client[] = app_config_old.oauth?.clients ?? [];
      const clients_new: Client[] = app_config_new.oauth?.clients ?? [];
      const newClients = findNewClients(clients_old, clients_new);
      for (const c of newClients) {
        await groupIdentify(projectID);
        await capture(actor_user_id, "createdApplication", {
          projectID,
          applicationType: c.x_application_type,
          $groups: {
            project: projectID,
          },
        });
      }
      break;
    }
    case "project.collaborator.invitation.accepted": {
      const actor_user_id = e.context.audit_context.actor_user_id;
      const projectID = e.context.app_id;
      await groupIdentify(projectID);
      await capture(actor_user_id, "acceptedInvite", {
        projectID,
        $groups: {
          project: projectID,
        },
      });
      break;
    }
  }
}
