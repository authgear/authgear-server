export default async function (e: any): Promise<any> {
  const u = e.payload.user;
  return {
    is_allowed: true,
    mutations: {
      id_token: {
        payload: {
          ...e.payload.id_token.payload,
          x_echo_sub: u?.id ?? "",
          x_echo_email: u?.standard_attributes?.email ?? "",
          x_echo_is_verified: String(u?.is_verified),
          x_echo_identity_count: String((e.payload.identities ?? []).length),
        },
      },
    },
  };
}
