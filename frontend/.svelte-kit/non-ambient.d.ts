
// this file is generated — do not edit it


declare module "svelte/elements" {
	export interface HTMLAttributes<T> {
		'data-sveltekit-keepfocus'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-noscroll'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-preload-code'?:
			| true
			| ''
			| 'eager'
			| 'viewport'
			| 'hover'
			| 'tap'
			| 'off'
			| undefined
			| null;
		'data-sveltekit-preload-data'?: true | '' | 'hover' | 'tap' | 'off' | undefined | null;
		'data-sveltekit-reload'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-replacestate'?: true | '' | 'off' | undefined | null;
	}
}

export {};


declare module "$app/types" {
	type MatcherParam<M> = M extends (param : string) => param is (infer U extends string) ? U : string;

	export interface AppTypes {
		RouteId(): "/" | "/edit" | "/edit/[id]" | "/input" | "/read" | "/read/[id]";
		RouteParams(): {
			"/edit/[id]": { id: string };
			"/read/[id]": { id: string }
		};
		LayoutParams(): {
			"/": { id?: string | undefined };
			"/edit": { id?: string | undefined };
			"/edit/[id]": { id: string };
			"/input": Record<string, never>;
			"/read": { id?: string | undefined };
			"/read/[id]": { id: string }
		};
		Pathname(): "/" | `/edit/${string}` & {} | "/input" | `/read/${string}` & {};
		ResolvedPathname(): `${"" | `/${string}`}${ReturnType<AppTypes['Pathname']>}`;
		Asset(): "/robots.txt" | string & {};
	}
}