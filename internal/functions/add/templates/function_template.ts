const handler = async (
  request: Request,
  params: Record<string, string>,
  staticMocks: any
) => {
  // add your handler logic here
  return Response.json({ hello: "protomok" });
};

export default handler;
// Deno.serve(handler);
