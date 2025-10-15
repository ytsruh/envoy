import type { SupportedCryptoAlgorithms } from "bun";
import { CryptoHasher } from "bun";
import type { MaybePromise } from "elysia";
import { Elysia } from "elysia";

export function etag(options: ETagOptions = {}) {
	if (typeof options.algorithm !== "string") {
		options.algorithm = "sha1";
	}

	const { serialize } = options;

	const hash = buildHashFn(options as Required<ETagOptions>);

	return new Elysia({ name: "etag", seed: options })
		.derive((ctx) => {
			let matchEtagValues: string[];
			let noneMatchEtagValues: string[];

			return {
				setETag(etag) {
					ctx.set.headers.etag = etag;
				},
				buildETagFor(response) {
					return hash(response);
				},
				isMatch(etag) {
					if (!matchEtagValues) {
						matchEtagValues = parseMatchHeader(ctx.headers["if-match"]);
					}

					return (
						matchEtagValues.includes(etag) || matchEtagValues.includes("*")
					);
				},
				isNoneMatch(etag) {
					if (!noneMatchEtagValues) {
						noneMatchEtagValues = parseMatchHeader(
							ctx.headers["if-none-match"],
						);
					}

					return (
						noneMatchEtagValues.includes(etag) ||
						noneMatchEtagValues.includes("*")
					);
				},
				setVary(headers) {
					ctx.set.headers.vary =
						typeof headers === "string" ? headers : headers.join(", ");
				},
			} satisfies ETagContextApi;
		})
		.onAfterHandle(async (ctx) => {
			const { request, set, response } = ctx;
			let etag = set.headers.etag;

			if (!etag) {
				let toHash: Bun.StringOrBuffer | undefined;

				if (canBeHashed(response)) {
					toHash = response;
				} else {
					if (typeof serialize === "function") {
						toHash = await serialize(response);
					}
					if (typeof toHash === "undefined") {
						return;
					}
				}

				etag = ctx.buildETagFor(toHash);
				ctx.setETag(etag);
			}

			if (ctx.isNoneMatch(etag)) {
				switch (request.method) {
					case "GET":
					case "HEAD":
						set.status = 304; // Not Modified
						break;
					default:
						set.status = 412; // Precondition Failed
						break;
				}

				ctx.responseValue = null;
			}
		})
		.as("global");
}

export type ETagOptions = {
	/**
	 * @default "sha1"
	 * @see https://bun.sh/docs/api/hashing
	 */
	algorithm?: ETagHashAlgorithm;
	/**
	 * @default false
	 */
	weak?: boolean;
	/**
	 * Provide your own hash function.
	 *
	 * If `hash` is provided option `algorithm` is ignored.
	 */
	hash?: ETagHashFunction;
	/**
	 * Converts incompatible data to a hashable format.
	 *
	 * @returns `undefined` if unsupported.
	 */
	serialize?: (response: unknown) => MaybePromise<ETagHashData | undefined>;
};

type ETagHashAlgorithm = "wyhash" | SupportedCryptoAlgorithms;
type ETagHashData = Bun.StringOrBuffer;
type ETagHashFunction = (response: ETagHashData) => string;

type ETagContextApi = {
	/**
	 * @see https://http.dev/etag
	 * @see https://www.rfc-editor.org/rfc/rfc2616#section-14.19
	 */
	setETag(etag: string): void;
	buildETagFor(response: ETagHashData): string;
	/**
	 * @see https://http.dev/if-match
	 * @see https://www.rfc-editor.org/rfc/rfc2616#section-14.24
	 */
	isMatch(etag: string): boolean;
	/**
	 * @see https://http.dev/if-none-match
	 * @see https://www.rfc-editor.org/rfc/rfc2616#section-14.26
	 */
	isNoneMatch(etag: string): boolean;
	/**
	 * @see https://http.dev/vary
	 * @see https://www.rfc-editor.org/rfc/rfc2616#section-14.44
	 */
	setVary(headers: "*" | string | string[]): void;
};

function parseMatchHeader(header?: string) {
	return header?.split(", ") ?? [];
}

function canBeHashed(response: unknown): response is Bun.StringOrBuffer {
	return (
		typeof response === "string" ||
		response instanceof ArrayBuffer ||
		response instanceof SharedArrayBuffer ||
		ArrayBuffer.isView(response)
	);
}

function validateAlgorithm(algorithm: ETagHashAlgorithm) {
	if (algorithm === "wyhash") {
		return true;
	}

	if (!CryptoHasher.algorithms.includes(algorithm)) {
		throw new TypeError(`Algorithm ${algorithm} not supported.`);
	}
}

function buildHashFn({
	algorithm,
	weak,
	hash,
}: Required<ETagOptions>): ETagHashFunction {
	const prefix = weak ? 'W/"' : '"';

	if (hash) {
		return (response) => `${prefix + hash(response)}"`;
	}

	validateAlgorithm(algorithm);

	if (algorithm === "wyhash") {
		return (response) => `${prefix + Bun.hash.wyhash(response)}"`;
	}

	return (response) =>
		`${prefix + CryptoHasher.hash(algorithm, response, "base64")}"`;
}
