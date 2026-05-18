<script lang="ts">
  import DataTable from './DataTable.svelte';
  import EmptyState from './EmptyState.svelte';

  interface Props { data: any[] | null; }
  let { data }: Props = $props();

  let activeTab = $state('landingPages');

  const TABS = [
    { id: 'landingPages', label: 'Landing Pages' },
    { id: 'items', label: 'Produtos' },
    { id: 'categories', label: 'Categorias' },
    { id: 'devices', label: 'Dispositivos' },
    { id: 'ai', label: 'Tráfego de IA' },
    { id: 'shopping', label: 'Loja/Shopping' },
  ];

  const COL_LANDING = [
    { key: 'date', label: 'Data' },
    { key: 'itemName', label: 'Item' },
    { key: 'sessions', label: 'Sessões' },
    { key: 'engagedSessions', label: 'Engajadas' },
    { key: 'conversions', label: 'Conversões' },
    { key: 'revenue', label: 'Receita (R$)' },
  ];

  let isStub = $derived(activeTab !== 'landingPages');
  let tabData = $derived(activeTab === 'landingPages' ? data : null);
</script>

<div class="sub-tabs" style="flex-wrap: wrap;">
  {#each TABS as tab (tab.id)}
    <button class="sub-tab" class:active={activeTab === tab.id} onclick={() => activeTab = tab.id} style="font-size: 11px; padding: 6px 10px;">
      {tab.label}
    </button>
  {/each}
</div>

<div class="panel">
  {#if isStub}
    <EmptyState message="Esta aba requer integração adicional no backend. A estrutura visual está preparada para quando os dados forem disponibilizados." />
  {:else if !tabData || tabData.length === 0}
    <EmptyState message="Sem dados do Google Analytics 4 para este período." />
  {:else}
    <DataTable data={tabData} columns={COL_LANDING} />
  {/if}
</div>
