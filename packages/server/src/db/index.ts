import { storeFile } from "../lib/storage";
import { PrismaClient } from "./generated/client";

let db: PrismaClient | null;

export const getDB = (): PrismaClient => {
	if (!db) {
		db = new PrismaClient();
		return db;
	}
	return db;
};

export const initializeDB = async (): Promise<void> => {
	if (!process.env.DATABASE_URL) {
		console.log("DATABASE_URL environment variable is not set");
		process.exit(1);
	}
	// Create the prisma client if it doesn't exist
	if (!db) {
		db = new PrismaClient();
		// Make sure the database is in WAL mode
		await db.$queryRaw`PRAGMA journal_mode = WAL;"`;
		console.log("Database initialized in WAL mode");
	}
	return;
};

export async function backupDB(): Promise<void> {
	if (!process.env.DATABASE_URL) {
		throw new Error(
			"DATABASE_URL environment variable or connection string is not set",
		);
	}

	if (!db) {
		throw new Error("Database not initialized");
	}

	// Checkpoint the WAL - this flushes WAL to the main database file & close the connection
	await db.$queryRaw`PRAGMA wal_checkpoint(TRUNCATE);"`;

	// Remove 'file:../.' prefix from the connection string
	const connstring = process.env.DATABASE_URL.split("file:../.");

	// Read the database file
	const dbBuffer = Bun.file(connstring[1]);

	// Upload to S3
	const dbKey = `backups/db/${Date.now()}.sqlite`;
	await storeFile(dbKey, dbBuffer);

	console.log(`Database backed up successfully`);
}
