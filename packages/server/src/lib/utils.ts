const REQUIRED_ENV_VARS = [
	"ENVIRONMENT",
	"DATABASE_URL",
	"STORAGE_BUCKET",
	"STORAGE_KEY",
	"STORAGE_SECRET",
	"STORAGE_ENDPOINT",
	"STORAGE_REGION",
] as const;

export type EnvVars = Record<(typeof REQUIRED_ENV_VARS)[number], string>;

export const envVarCheck = () => {
	const missing = REQUIRED_ENV_VARS.filter((varName) => !process.env[varName]);

	if (missing.length > 0) {
		throw new Error(
			`Missing required environment variables: ${missing.join(", ")}`,
		);
	}
};
