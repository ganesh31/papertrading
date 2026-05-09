import Fastify, { type FastifyReply, type FastifyRequest } from "fastify";
import client from "prom-client";
import websocket from "@fastify/websocket";
import { WebSocket } from "ws";

const port = Number(process.env.GATEWAY_PORT || 4000);
const mdBaseUrl = process.env.MD_BASE_URL || "http://md:6011";

const register = new client.Registry();
client.collectDefaultMetrics({ register });

const helloCounter = new client.Counter({
  name: "hello_requests_total",
  help: "Count of hello-world requests",
  registers: [register],
});

const app = Fastify({ logger: true });
await app.register(websocket);

function mdOrigin(): string {
  return mdBaseUrl.replace(/\/$/, "");
}

/** Pass-through GET to md (Phase 1 REST). */
async function proxyMdGet(req: FastifyRequest, reply: FastifyReply, mdPath: string) {
  const incoming = new URL(req.url, "http://127.0.0.1");
  const target = new URL(mdPath + incoming.search, `${mdOrigin()}/`);
  try {
    const res = await fetch(target);
    reply.code(res.status);
    const ct = res.headers.get("content-type");
    if (ct) reply.header("content-type", ct);
    const buf = Buffer.from(await res.arrayBuffer());
    return reply.send(buf);
  } catch (err) {
    app.log.warn({ err }, "md proxy fetch failed");
    return reply.code(502).send({ error: "md unreachable" });
  }
}

app.get("/instruments", async (req, reply) => proxyMdGet(req, reply, "/instruments"));

app.get("/candles", async (req, reply) => proxyMdGet(req, reply, "/candles"));

app.get("/market/status", async (req, reply) => proxyMdGet(req, reply, "/market/status"));

app.get("/replay/status", async (req, reply) => proxyMdGet(req, reply, "/replay/status"));

app.get("/healthz", async () => {
  helloCounter.inc();
  return { ok: true, service: "gateway" };
});

app.get("/metrics", async (_req, reply) => {
  reply.header("Content-Type", register.contentType);
  return await register.metrics();
});

type SubscribeMsg = { subscribe: string[] };

app.get(
  "/stream",
  { websocket: true },
  (socket, _req) => {
    const upstream = new WebSocket(`${mdBaseUrl.replace(/^http/, "ws")}/stream`);
    let pendingSubscribe: string | null = null;

    const closeAll = (code = 1000, reason = "bye") => {
      try {
        socket.close(code, reason);
      } catch {}
      try {
        upstream.close(code, reason);
      } catch {}
    };

    upstream.on("open", () => {
      if (pendingSubscribe) {
        upstream.send(pendingSubscribe);
        pendingSubscribe = null;
      }
    });

    upstream.on("message", (data: any) => {
      if (socket.readyState === socket.OPEN) {
        socket.send(data.toString());
      }
    });

    upstream.on("close", (code: number, reason: Buffer) => {
      closeAll(code, reason.toString() || "upstream closed");
    });

    upstream.on("error", (err: Error) => {
      app.log.warn({ err }, "md upstream ws error");
      closeAll(1011, "upstream error");
    });

    socket.on("message", (data) => {
      // Best-effort validate shape; still proxy raw for forwards compat.
      try {
        const msg = JSON.parse(data.toString()) as SubscribeMsg;
        if (!msg || !Array.isArray(msg.subscribe)) return;
      } catch {
        return;
      }
      const payload = data.toString();
      if (upstream.readyState === upstream.OPEN) {
        upstream.send(payload);
      } else {
        pendingSubscribe = payload;
      }
    });

    socket.on("close", (code: number, reason: Buffer) => {
      closeAll(code, reason.toString() || "client closed");
    });

    socket.on("error", (err: Error) => {
      app.log.warn({ err }, "gateway ws client error");
      closeAll(1011, "client error");
    });
  },
);

await app.listen({ host: "0.0.0.0", port });
