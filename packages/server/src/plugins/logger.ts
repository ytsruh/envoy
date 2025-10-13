import Elysia from "elysia";

interface Writer {
	write: (message: string) => void;
}

const consoleWriter: Writer = {
	write(message: string) {
		console.log(message);
	},
};

interface Options {
	logIP?: boolean;
	writer?: Writer;
}

export const logger = (options?: Options) => {
	const { write } = options?.writer || consoleWriter;
	return new Elysia({
		name: "logger",
	})
		.onRequest((ctx) => {
			ctx.store = { ...ctx.store, beforeTime: process.hrtime.bigint() };
		})

		.onBeforeHandle({ as: "global" }, (ctx) => {
			ctx.store = { ...ctx.store, beforeTime: process.hrtime.bigint() };
		})
		.onAfterHandle({ as: "global" }, ({ request, store }) => {
			const logStr: string[] = [];
			if (options?.logIP) {
				if (request.headers.get("X-Forwarded-For")) {
					logStr.push(`[${request.headers.get("X-Forwarded-For")}]`);
				}
			}

			logStr.push(request.method);

			logStr.push(new URL(request.url).pathname);
			const beforeTime: bigint = (store as { beforeTime: bigint }).beforeTime;

			logStr.push(durationString(beforeTime));

			write(logStr.join(" "));
		})
		.onError({ as: "global" }, ({ request, error, store }) => {
			const logStr: string[] = [];

			logStr.push(request.method);

			logStr.push(new URL(request.url).pathname);

			logStr.push("Error");

			if ("status" in error) {
				logStr.push(String(error.status));
			}

			if ("message" in error && typeof error.message === "string") {
				logStr.push(error.message);
			}
			const beforeTime: bigint = (store as { beforeTime: bigint }).beforeTime;

			logStr.push(durationString(beforeTime));

			write(logStr.join(" "));
		});
};

function durationString(beforeTime: bigint): string {
	const now = process.hrtime.bigint();
	const timeDifference = now - beforeTime;
	const nanoseconds = Number(timeDifference);

	const durationInMicroseconds = (nanoseconds / 1e3).toFixed(0); // Convert to microseconds
	const durationInMilliseconds = (nanoseconds / 1e6).toFixed(0); // Convert to milliseconds
	let timeMessage: string = "";

	if (nanoseconds >= 1e9) {
		const seconds = (nanoseconds / 1e9).toFixed(2);
		timeMessage = `| ${seconds}s`;
	} else if (nanoseconds >= 1e6) {
		timeMessage = `| ${durationInMilliseconds}ms`;
	} else if (nanoseconds >= 1e3) {
		timeMessage = `| ${durationInMicroseconds}µs`;
	} else {
		timeMessage = `| ${nanoseconds}ns`;
	}

	return timeMessage;
}
