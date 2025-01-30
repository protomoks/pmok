import * as posix from "https://deno.land/std/path/posix/mod.ts";
import { parse } from "jsr:@std/yaml";
import { walk } from "jsr:@std/fs/walk";

const ASCIIART = `

  ___         _                 _   
 | _ \\_ _ ___| |_ ___ _ __  ___| |__
 |  _/ '_/ _ \\  _/ _ \\ '  \\/ _ \\ / /
 |_| |_| \\___/\\__\\___/_|_|_\\___/_\\_\\
                                    
`;

interface Function {
  path: string;
  entrypoint: string;
  methods: string[];
}
interface FunctionConfig {
  [name: string]: Function;
}
interface StaticMock {
  request: {
    method: string;
    path: string;
    headers: Record<string, string[]>;
  };
  response: {
    status: number;
    headers: Record<string, string[]>;
    body?: any;
  };
}

const PROTOMOK_CONFIG_ENCODING = Deno.env.get("PROTOMOK_CONFIG_ENCODING")!;
let functionConfig: FunctionConfig = {};

type Methods = "GET" | "POST" | "PUT" | "DELETE" | "PATCH";

type NodeData = {
  [key in Methods]?: string[];
};
class RadixNode {
  children: Record<string, RadixNode> = {};
  value: NodeData | null = null;
  constructor(value?: NodeData) {
    this.value = value || null;
  }

  insert(
    segments: string[],
    value: {
      method: Methods;
      filePath: string;
    }
  ): RadixNode {
    if (segments.length == 0) {
      if (!this.value) {
        this.value = {
          [value.method]: [value.filePath],
        };
      } else if (this.value[value.method]) {
        this.value[value.method]?.push(value.filePath);
      } else {
        this.value[value.method] = [value.filePath];
      }
      return this;
    }

    const [segment, ...rest] = segments;
    let child = this.children[segment];
    if (!child) {
      child = new RadixNode();
      this.children[segment] = child;
    }
    return child.insert(rest, value);
  }

  get(segments: string[]): NodeData | null {
    if (segments.length == 0) {
      return this.value;
    }

    const [segment, ...rest] = segments;
    const child = this.children[segment];
    if (!child) {
      return null;
    }
    return child.get(rest);
  }

  size(): number {
    let size = 0;
    if (this.value) {
      size += Object.keys(this.value)
        .map((key) => (this.value && this.value[key as Methods]?.length) || 0)
        .reduce((a, b) => a + b, 0);
    }
    for (const key in this.children) {
      size += this.children[key].size();
    }
    return size;
  }
}

type LogLevel = "DEBUG" | "INFO" | "WARN" | "ERROR";

class Logger {
  public readonly logLevel: LogLevel;

  constructor() {
    const envLogLevel = Deno.env.get("PMOK_LOG_LEVEL") || "INFO";
    this.logLevel = this.parseLogLevel(envLogLevel);
  }

  private parseLogLevel(level: string): LogLevel {
    const levels: LogLevel[] = ["DEBUG", "INFO", "WARN", "ERROR"];
    if (levels.includes(level as LogLevel)) {
      return level as LogLevel;
    }
    return "INFO";
  }

  private shouldLog(level: LogLevel): boolean {
    const levels: LogLevel[] = ["DEBUG", "INFO", "WARN", "ERROR"];
    return levels.indexOf(level) >= levels.indexOf(this.logLevel);
  }

  debug(message: string, ...args: any[]): void {
    if (this.shouldLog("DEBUG")) {
      console.debug(`[DEBUG] ${message}`, ...args);
    }
  }

  info(message: string, ...args: any[]): void {
    if (this.shouldLog("INFO")) {
      console.info(`[INFO] ${message}`, ...args);
    }
  }

  warn(message: string, ...args: any[]): void {
    if (this.shouldLog("WARN")) {
      console.warn(`[WARN] ${message}`, ...args);
    }
  }

  error(message: string, ...args: any[]): void {
    if (this.shouldLog("ERROR")) {
      console.error(`[ERROR] ${message}`, ...args);
    }
  }
}

const logger = new Logger();

const buildRadixTree = async (): Promise<RadixNode> => {
  const root = new RadixNode();
  const mockDir = posix.join(Deno.cwd(), "protomok/mocks");
  for await (const entry of walk(mockDir, { exts: ["json"] })) {
    const decoder = new TextDecoder();
    const data = await Deno.readFile(entry.path);
    const json = JSON.parse(decoder.decode(data));
    const mockPath = json.request.path;
    const mockMethod = json.request.method;
    const mockFilePath = entry.path;
    const segments = mockPath.split("/");
    const node = root.insert(segments, {
      method: mockMethod.toUpperCase(),
      filePath: mockFilePath,
    });

    logger.debug(`Inserted ${mockPath} with method ${mockMethod}`);
    logger.debug(`Node value`, node.value);
  }

  return root;
};

const readConfig = async () => {
  const decoder = new TextDecoder();
  const configPath = posix.join(
    Deno.cwd(),
    `protomok/pmok.${PROTOMOK_CONFIG_ENCODING}`
  );
  logger.debug("Reading config at", configPath);
  const data = await Deno.readFile(configPath);
  let json;
  if (PROTOMOK_CONFIG_ENCODING === "json") {
    json = JSON.parse(decoder.decode(data));
  } else {
    json = parse(decoder.decode(data));
  }
  return json;
};

const findFunctionMatch = (
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

const findStaticMatch = async (
  req: Request,
  root: RadixNode
): Promise<StaticMock | null> => {
  const segments = new URL(req.url).pathname.split("/");
  const method = req.method.toUpperCase() as Methods;
  const value = root.get(segments);
  if (value && req.method.toUpperCase() in value) {
    const filePath = value[method] ? value[method][0] : null;
    if (filePath) {
      const data = await Deno.readFile(filePath);
      const json = JSON.parse(new TextDecoder().decode(data));
      return json;
    }
  }
  return null;
};

const toHeaders = (headers: Record<string, string[]>) => {
  const h = new Headers();
  for (const key in headers) {
    h.set(key, headers[key].join(","));
  }
  return h;
};

const executeUserFunction = async (
  req: Request,
  params: Record<string, string | undefined>,
  handler: (
    req: Request,
    params: Record<string, string | undefined>,
    staticMock: StaticMock | null
  ) => Promise<Response>,
  staticMock: StaticMock | null
): Promise<Response> => {
  return await handler(req, params, staticMock);
};

const importUserModule = async (
  name: string,
  entrypoint: string
): Promise<{
  default: (
    req: Request,
    params: Record<string, string | undefined>,
    staticMock: StaticMock | null
  ) => Promise<Response>;
}> => {
  return await import(
    posix.join(
      Deno.cwd(),
      `protomok/functions/${name}/${entrypoint}?t=${Date.now()}`
    )
  );
};

const main = async () => {
  const config = await readConfig();
  logger.debug("Config");
  logger.debug(config);
  functionConfig = config.functions;

  const radixTree = await buildRadixTree();

  Deno.serve({
    handler: async (req: Request) => {
      console.error("Received a request", req.url);
      // look for a match
      // 1. look if we have a static mock stored in our radix tree
      const staticMatch = await findStaticMatch(req, radixTree);
      if (staticMatch) {
        logger.debug(`Matched a static mock for ${req.url}`);
      }
      // 2. look if we have a function match
      const [name, fn, params] = findFunctionMatch(req);
      // if we don't have a function match and we don't have a static match return 404
      if (!fn && !staticMatch) {
        // todo account of content-type header
        return new Response("Not Found", {
          status: 404,
          statusText: "not found",
        });
      }
      // if we don't have a function match, simply return the static match
      if (!fn && staticMatch) {
        return new Response(JSON.stringify(staticMatch.response.body), {
          status: staticMatch.response.status,
          headers: new Headers(toHeaders(staticMatch.response.headers)),
        });
      }

      // if we have a function match, execute the function
      if (fn) {
        const module = await importUserModule(name, fn.entrypoint);
        return await executeUserFunction(
          req,
          params,
          module.default,
          staticMatch
        );
      }
      if (staticMatch) {
        return new Response(JSON.stringify(staticMatch.response), {
          status: staticMatch.response.status,
          headers: new Headers(toHeaders(staticMatch.response.headers)),
        });
      }
      return new Response("Not Found", {
        status: 404,
        statusText: "not found",
      });
    },
    onListen: () => {
      console.log(ASCIIART);
      console.log(
        `Version:\t0.0.1\nAddress:\thttp://127.0.0.1:8000\nStatic Mocks:\t${radixTree.size()}\nFunctions:\t${
          Object.keys(functionConfig).length
        }\nLog Level:\t${logger.logLevel}`
      );
    },
  });
};

main();
