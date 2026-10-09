export default async function (e: any): Promise<any> {
  return {
    is_allowed: false,
    reasons: [
      {
        type: "invalid_form_field",
        field: { user_profile: { pointer: "/x_not_in_form" } },
        message: "Not a field of the form",
      },
    ],
  };
}
