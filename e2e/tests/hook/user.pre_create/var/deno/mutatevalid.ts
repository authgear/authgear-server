export default async function (e: any): Promise<any> {
  return {
    is_allowed: true,
    mutations: {
      user: {
        custom_attributes: { x_employee_id: "E123" },
      },
    },
  };
}
