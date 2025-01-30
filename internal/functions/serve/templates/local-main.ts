import * as posix from "https://deno.land/std/path/posix/mod.ts";
import { parse } from "jsr:@std/yaml";

interface Function {
  path: string;
  entrypoint: string;
  methods: string[];
}
interface FunctionConfig {
  [name: string]: Function;
}

//const PROTOMOK_FUNCTION_CONFIG = Deno.env.get("PROTOMOK_FUNCTION_CONFIG")!;
const PROTOMOK_CONFIG_ENCODING = Deno.env.get("PROTOMOK_CONFIG_ENCODING")!;
let functionConfig: FunctionConfig = {};
// (function parseConfig() {
//   try {
//     functionConfig = JSON.parse(PROTOMOK_FUNCTION_CONFIG);
//   } catch (e) {
//     console.error(
//       `Unable to parse function config ${PROTOMOK_FUNCTION_CONFIG}`
//     );
//   }
// })();

const readConfig = async () => {
  const decoder = new TextDecoder();
  const configPath = posix.join(
    Deno.cwd(),
    `protomok/pmok.${PROTOMOK_CONFIG_ENCODING}`
  );
  console.log("Reading config at", configPath);
  const data = await Deno.readFile(configPath);
  let json;
  if (PROTOMOK_CONFIG_ENCODING === "json") {
    json = JSON.parse(decoder.decode(data));
  } else {
    json = parse(decoder.decode(data));
  }
  return json;
};

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

const main = async () => {
  const config = await readConfig();
  console.log("Config");
  console.log(config);
  functionConfig = config.functions;

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
      // execute the function
      const entrypoint = posix.join(
        Deno.cwd(),
        `protomok/functions/${name}/${fn.entrypoint}`
      );
      console.log("Entrypoint", entrypoint);
      // add a timestamp to the import to avoid caching
      const module = await import(`${entrypoint}?t=${Date.now()}`);
      const handler = module.default;
      return await handler(req, params);
    },
    onListen: () => {
      console.error("Listening on http://127.0.0.1:8000");
    },
  });
};

main();
