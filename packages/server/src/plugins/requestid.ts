import { randomUUID } from "node:crypto";
import Elysia from "elysia";

type Options = {
	uuid?: () => string;
	header?: string;
};

export const requestID = ({
	uuid = randomUUID,
	header = "X-Request-ID",
}: Readonly<Options> = {}) => {
	return new Elysia({ name: "requestid", seed: header })
		.onRequest(({ set, request: { headers } }) => {
			set.headers[header] = headers.get(header) || uuid();
		})
		.derive(({ set }) => {
			return {
				requestID: set.headers[header],
			};
		});
};
