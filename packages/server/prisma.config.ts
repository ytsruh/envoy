import "dotenv/config";
import * as path from "node:path";
import { defineConfig } from "prisma/config";

export default defineConfig({
	schema: path.join("src", "db", "schema.prisma"),
	migrations: {
		path: path.join("src", "db", "migrations"),
	},
});
