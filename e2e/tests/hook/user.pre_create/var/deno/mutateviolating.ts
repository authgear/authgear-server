// Writes a value that violates the pattern of x_employee_id.
export default async function (e: any): Promise<any> {
  return {
    is_allowed: true,
    mutations: {
      user: {
        custom_attributes: { x_employee_id: "123" },
      },
    },
  };
}
