// Allows the update only when the user already has the attribute filled in the form.
export default async function (e: any): Promise<any> {
  if (e.payload.user.standard_attributes.given_name !== "John") {
    return { is_allowed: false, reason: "given_name is not applied" };
  }
  return { is_allowed: true };
}
