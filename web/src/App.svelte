<script lang="ts">
	import { Router, Route } from 'svelte-routing';

	import Home from './views/home.svelte';
	import Container from './views/container.svelte';

	export let name: string;
</script>

<main>
	<Router>
		<Route path="/" component={Home} />
		<Route path="/containers/*container" let:params>
			<Container containerKey="{params.container}" />
		</Route>
	</Router>
</main>

<style>
	main {
		text-align: center;
		padding: 1em;
		max-width: 240px;
		margin: 0 auto;
	}

	h1 {
		color: #ff3e00;
		text-transform: uppercase;
		font-size: 4em;
		font-weight: 100;
	}

	@media (min-width: 640px) {
		main {
			max-width: none;
		}
	}
</style>