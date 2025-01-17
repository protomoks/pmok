const HOST_PORT = Deno.env.get("PROTOMOK_INTERNAL_HOST_PORT")!;
import * as posix from "https://deno.land/std/path/posix/mod.ts";

interface Function {
  path: string;
  entrypoint: string;
  methods: string[];
}
interface FunctionConfig {
  [name: string]: Function;
}

const PROTOMOK_FUNCTION_CONFIG = Deno.env.get("PROTOMOK_FUNCTION_CONFIG")!;
let functionConfig: FunctionConfig = {};
(function parseConfig() {
  try {
    functionConfig = JSON.parse(PROTOMOK_FUNCTION_CONFIG);
  } catch (e) {
    console.error(
      `Unable to parse function config ${PROTOMOK_FUNCTION_CONFIG}`
    );
  }
})();

const findMatch = (
  req: Request
): [string, Function | null, Record<string, string | undefined>] => {
  let match: Function | null = null;
  let key = "";
  let params: Record<string, string | undefined> = {};
  for (const name in functionConfig) {
    const fn = functionConfig[name];
    // don't even consider the function if there is not match on the method
    if (
      fn.methods.indexOf("*") < 0 &&
      fn.methods.indexOf(req.method.toUpperCase()) < 0
    )
      continue;

    const patternMatch = new URLPattern({ pathname: fn.path }).exec(req.url);
    if (!patternMatch) {
      continue;
    }
    params = patternMatch.pathname.groups;
    match = fn;
    key = name;
    break;
  }
  return [key, match, params];
};

Deno.serve({
  handler: async (req: Request) => {
    console.error("Received a request", req.url);
    // look for a match
    const [name, fn, params] = findMatch(req);
    if (!fn) {
      // todo account of content-type header
      return new Response("Not Found", {
        status: 404,
        statusText: "not found",
      });
    }
    const paramEnv = [`PROTOMOK_PARAMS`, JSON.stringify(params)];
    const absEntryPoint = posix.join(
      Deno.cwd(),
      `protomok/functions/${name}/index.ts`
    );
    const maybeEntryPoint = posix.toFileUrl(absEntryPoint).href;
    console.log(`MaybeEntryPoint ${maybeEntryPoint}`);
    const servicePath = posix.dirname(`protomok/functions/${name}/index.ts`);
    console.log(`Serving request on ${servicePath}`);
    console.log(`Supplying param env ${paramEnv}`);
    try {
      const worker = await EdgeRuntime.userWorkers.create({
        servicePath,
        memoryLimitMb: 256,
        workerTimeoutMs: 2000,
        noModuleCache: false,
        envVars: [paramEnv],
        forceCreate: false,
        customModuleRoot: "",
        cpuTimeSoftLimitMs: 1000,
        cpuTimeHardLimit: 2000,
        decoratorType: "tc39",
        maybeEntryPoint,
        context: {
          useReadSyncFileAPI: true,
        },
      });
      return await worker.fetch(req);
    } catch (e) {
      console.error(e);
      return new Response(`An Error occured\n${e}`, { status: 500 });
    }
  },
  onListen: () => {
    console.log(`Serving mock handlers on http://127.0.0.1:${HOST_PORT}`);
  },
  onError: (e) => {
    return Response.json({ message: e }, { status: 500 });
  },
});
