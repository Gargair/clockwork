type ViteEnv = { VITE_API_BASE_URL?: string };
type ImportMetaWithEnv = ImportMeta & { env?: ViteEnv };

const env: ViteEnv | undefined = (import.meta as unknown as ImportMetaWithEnv).env;

export const API_BASE_URL: string =
	trimTrailingSlash(env?.VITE_API_BASE_URL ??
	(typeof globalThis !== 'undefined' && globalThis.location ? globalThis.location.origin : ''));

function trimTrailingSlash(url: string): string {
	if(url.endsWith('/')) {
		return url.slice(0, -1);
	}
	return url;
}