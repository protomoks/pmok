const handler = async (request: Request) => {
  // add your handler logic here
  return Response.json({ hello: "protomok" });
};

Deno.serve(handler);
