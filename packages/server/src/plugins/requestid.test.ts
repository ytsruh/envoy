import { expect, test } from "bun:test";
import Elysia from "elysia";
import { requestID } from "./requestid.js";

test("new request ID header is set when one isn't set", async () => {
	const response = await new Elysia()
		.use(requestID({ uuid: () => "random-uuid" }))
		.get("/", () => "random")
		.handle(new Request("https://random.com/"));

	expect(response.headers.get("X-Request-ID")).toBe("random-uuid");
});

test("forwards the request ID header when available", async () => {
	const response = await new Elysia()
		.use(requestID({ uuid: () => "random-uuid" }))
		.get("/", () => "test")
		.handle(
			new Request("https://test.com/", {
				headers: { "X-Request-ID": "existing-uuid" },
			}),
		);

	expect(response.headers.get("X-Request-ID")).toBe("existing-uuid");
});

test("allows the request ID header name to be set", async () => {
	const response = await new Elysia()
		.use(
			requestID({ uuid: () => "random-uuid", header: "X-Custom-Request-ID" }),
		)
		.get("/", () => "test")
		.handle(new Request("https://test.com/"));

	expect(response.headers.get("X-Custom-Request-ID")).toBe("random-uuid");
});

test("forwards custom request ID when available", async () => {
	const response = await new Elysia()
		.use(
			requestID({ uuid: () => "random-uuid", header: "X-Custom-Request-ID" }),
		)
		.get("/", () => "test")
		.handle(
			new Request("https://test.com/", {
				headers: { "X-Custom-Request-ID": "existing-uuid" },
			}),
		);

	expect(response.headers.get("X-Custom-Request-ID")).toBe("existing-uuid");
});
