// Allows only form "employee" whose user has the submitted x_employee_id applied.
export default async function (e: any): Promise<any> {
  const value = e.payload.form.fields.find(
    (f: any) => f.field.user_profile.pointer === "/x_employee_id",
  )?.value;
  if (e.payload.form.name !== "employee") {
    return { is_allowed: false, reason: "unexpected form name" };
  }
  if (e.payload.user.custom_attributes.x_employee_id !== value) {
    return { is_allowed: false, reason: "the user does not have the submitted value" };
  }
  if (value === "E99999") {
    return {
      is_allowed: false,
      title: "Invalid employee ID",
      reasons: [
        {
          type: "invalid_form_field",
          field: { user_profile: { pointer: "/x_employee_id" } },
          message: "Unknown employee",
        },
      ],
    };
  }
  return { is_allowed: true };
}
