export default async function (e: any): Promise<any> {
  const appID = e?.context?.app_id ?? "";
  await fetch(`http://hook:2626/deno/${appID}/usage`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(e),
  });
  return {};
}
