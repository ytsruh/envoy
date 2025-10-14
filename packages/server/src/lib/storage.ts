import { S3Client } from "bun";

let s3Client: S3Client | null = null;

function getS3Client(): S3Client {
	if (!s3Client) {
		s3Client = new S3Client({
			accessKeyId: process.env.STORAGE_KEY,
			secretAccessKey: process.env.STORAGE_SECRET,
			bucket: process.env.STORAGE_BUCKET,
			endpoint: process.env.STORAGE_ENDPOINT,
			region: process.env.STORAGE_REGION,
		});
	}
	return s3Client;
}

export const createPutPresignedUrl = (key: string) => {
	const client = getS3Client();
	return client.presign(key, {
		expiresIn: 3600, // 1 hour
		method: "PUT",
		type: "application/octet-stream",
	});
};

export const createGetPresignedUrl = (key: string) => {
	const client = getS3Client();
	return client.presign(key, {
		expiresIn: 3600, // 1 hour
		method: "GET",
		type: "application/octet-stream",
	});
};

interface S3Object {
	key: string;
	lastModified?: string;
	size?: number;
	etag?: string;
}

export async function storeFile(
	key: string,
	data: string | ArrayBuffer | Bun.BunFile | Bun.S3File | Blob | File,
	type?: string,
): Promise<number> {
	const client = getS3Client();
	try {
		const bytesWritten = await client.write(key, data, {
			type: type || "application/octet-stream",
		});
		return bytesWritten;
	} catch (error) {
		console.error("Error storing image:", error);
		return 0;
	}
}
