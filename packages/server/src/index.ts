import { cors } from "@elysiajs/cors";
import { Elysia } from "elysia";
import { initializeDB } from "./db";
import { envVarCheck } from "./lib/utils";
import { etag } from "./plugins/etag";
import { logger } from "./plugins/logger";
import { requestID } from "./plugins/requestid";

console.log("-----Server setup-----");
envVarCheck();
await initializeDB();
console.log("-----Server setup-----");

new Elysia()
	.use(cors())
	.use(logger())
	.use(etag())
	.use(requestID())
	.get("/", async () => {
		return new Response("Hello Envoy!", {
			headers: { "Content-Type": "text/html" },
		});
	})
	.onError((ctx) => {
		if (ctx.code === "NOT_FOUND") return ctx.status(404, "Not Found :(");
		console.log(`Error: ${ctx.code} - ${ctx.status}`);
		return ctx.status(500, "Internal Server Error");
	})
	.listen(3000, (app) => {
		console.log(`🚀 running on http://${app.hostname}:${app.port}`);
	});
