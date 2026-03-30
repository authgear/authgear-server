export default async function (e: any): Promise<any> {
  const appID = e?.context?.app_id ?? "";
  const period = e?.payload?.usage?.period ?? "";
  if (period === "day" || period === "month") {
    await fetch(`http://hook:2626/deno/${appID}/email-${period}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(e),
    });
  }
  return {};
}
