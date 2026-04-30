import Fastify from "fastify";
import client from "prom-client";

const port = Number(process.env.GATEWAY_PORT || 4000);

const register = new client.Registry();
client.collectDefaultMetrics({ register });

const helloCounter = new client.Counter({
  name: "hello_requests_total",
  help: "Count of hello-world requests",
  registers: [register],
});

const app = Fastify({ logger: true });

app.get("/healthz", async () => {
  helloCounter.inc();
  return { ok: true, service: "gateway" };
});

app.get("/metrics", async (_req, reply) => {
  reply.header("Content-Type", register.contentType);
  return await register.metrics();
});

await app.listen({ host: "0.0.0.0", port });
