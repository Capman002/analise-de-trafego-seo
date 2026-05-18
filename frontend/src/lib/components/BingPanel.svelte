<script lang="ts">
  import DataTable from './DataTable.svelte';
  import EmptyState from './EmptyState.svelte';

  interface Props { data: any[] | null; }
  let { data }: Props = $props();

  const COL_QUERIES = [
    { key: 'key', label: 'Consulta' },
    { key: 'clicks', label: 'Cliques' },
    { key: 'impressions', label: 'Impressões' },
    { key: 'position', label: 'Posição Média' },
  ];

  let activeTab = $state('queries');
  const TABS = [{ id: 'queries', label: 'Consultas' }];
</script>

<div class="sub-tabs">
  {#each TABS as tab (tab.id)}
    <button class="sub-tab" class:active={activeTab === tab.id} onclick={() => activeTab = tab.id}>
      {tab.label}
    </button>
  {/each}
</div>

<div class="panel">
  {#if !data || data.length === 0}
    <EmptyState message="Sem dados do Bing Webmaster para este período." />
  {:else}
    <DataTable data={data} columns={COL_QUERIES} />
  {/if}
</div>
