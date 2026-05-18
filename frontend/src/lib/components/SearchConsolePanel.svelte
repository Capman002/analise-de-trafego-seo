<script lang="ts">
  import DataTable from './DataTable.svelte';
  import EmptyState from './EmptyState.svelte';

  interface Props { queries: any[]; pages: any[]; comparing?: boolean; }
  let { queries, pages, comparing = false }: Props = $props();

  let activeTab = $state('pages');

  const TABS = [
    { id: 'pages', label: 'Páginas' },
    { id: 'queries', label: 'Consultas' },
  ];

  let COL_PAGES = $derived.by(() => {
    let cols = [{ key: 'key', label: 'Página' }];
    cols.push({ key: 'clicks', label: 'Cliques' });
    if (comparing) cols.push({ key: 'prev_clicks', label: 'Cliq. (Ant)' });
    cols.push({ key: 'impressions', label: 'Impressões' });
    if (comparing) cols.push({ key: 'prev_impressions', label: 'Impr. (Ant)' });
    cols.push({ key: 'ctr', label: 'CTR %' });
    cols.push({ key: 'position', label: 'Posição' });
    return cols;
  });

  let COL_QUERIES = $derived.by(() => {
    let cols = [{ key: 'key', label: 'Consulta' }];
    cols.push({ key: 'clicks', label: 'Cliques' });
    if (comparing) cols.push({ key: 'prev_clicks', label: 'Cliq. (Ant)' });
    cols.push({ key: 'impressions', label: 'Impressões' });
    if (comparing) cols.push({ key: 'prev_impressions', label: 'Impr. (Ant)' });
    cols.push({ key: 'ctr', label: 'CTR %' });
    cols.push({ key: 'position', label: 'Posição' });
    return cols;
  });

  let tabData = $derived(
    activeTab === 'pages' ? pages :
    activeTab === 'queries' ? queries :
    []
  );

  let columns = $derived(
    activeTab === 'pages' ? COL_PAGES : COL_QUERIES
  );
</script>

<div class="sub-tabs" style="flex-wrap: wrap;">
  {#each TABS as tab (tab.id)}
    <button class="sub-tab" class:active={activeTab === tab.id} onclick={() => activeTab = tab.id} style="font-size: 11px; padding: 6px 10px;">
      {tab.label}
    </button>
  {/each}
</div>

<div class="panel">
  {#if !tabData || tabData.length === 0}
    <EmptyState message="Nenhum dado disponível para esta aba." />
  {:else}
    <DataTable data={tabData} {columns} />
  {/if}
</div>
