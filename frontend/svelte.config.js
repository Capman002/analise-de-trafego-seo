import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
		runes: ({ filename }) => (filename.split(/[/\\]/).includes('node_modules') ? undefined : true)
	},
	kit: {
		adapter: adapter({
			// Output estático para ser embutido no binário Go via embed.FS
			pages: 'build',
			assets: 'build',
			fallback: 'index.html', // SPA fallback — Go serve index.html para rotas desconhecidas
			strict: false
		})
	}
};

export default config;
