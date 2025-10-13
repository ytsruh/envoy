import { Elysia } from "elysia";

new Elysia()
	.get("/", async () => {
		return new Response("Hello World!", {
			headers: { "Content-Type": "text/html" },
		});
	})
	.listen(3000, (app) => {
		console.log(`🚀 running on http://${app.hostname}:${app.port}`);
	});
