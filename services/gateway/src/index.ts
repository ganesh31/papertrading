import Fastify from "fastify";
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
