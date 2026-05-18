<script lang="ts">
  import { onMount } from 'svelte';
  import ClientSelector from '$lib/components/ClientSelector.svelte';
  import Filters from '$lib/components/Filters.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import GscChart from '$lib/components/GscChart.svelte';
  import OverviewPanel from '$lib/components/OverviewPanel.svelte';
  import SearchConsolePanel from '$lib/components/SearchConsolePanel.svelte';
  import GA4Panel from '$lib/components/GA4Panel.svelte';
  import BingPanel from '$lib/components/BingPanel.svelte';
  import { fetchClients, fetchTraffic } from '$lib/api';
  import type { Client, TrafficResponse } from '$lib/types';

  // ── Estado reativo (Svelte 5 runes) ──
  let clients = $state<Client[]>([]);
  let selectedClient = $state<Client | null>(null);
  let period = $state(28);
  let startDate = $state('');
  let endDate = $state('');
  let location = $state('ALL');
  let trafficData = $state<TrafficResponse | null>(null);
  let loadingClients = $state(true);
  let loadingTraffic = $state(false);
  let error = $state('');

  // ── Tabs de fonte ──
  type SourceTab = 'overview' | 'gsc' | 'ga4' | 'bing';
  let activeTab = $state<SourceTab>('overview');

  // ── Expansão do Container ──
  let isExpanded = $state(false);

  // ── Toggle de comparação ──
  let comparing = $state(false);

  function toggleExpand() {
    isExpanded = !isExpanded;
    if (typeof document !== 'undefined') {
      if (isExpanded) {
        document.body.classList.add('container-expanded');
      } else {
        document.body.classList.remove('container-expanded');
      }
    }
  }

  // ── Dados derivados ──
  let hasData = $derived(
    trafficData !== null &&
    (trafficData.gsc_queries?.length > 0 || trafficData.gsc_pages?.length > 0 || trafficData.ga4?.length > 0 || trafficData.bing?.length > 0)
  );

  // ── Lifecycle ──
  onMount(async () => {
    try {
      clients = await fetchClients();
    } catch (err: unknown) {
      error = err instanceof Error ? err.message : 'Falha ao carregar clientes';
    } finally {
      loadingClients = false;
    }
  });

  // ── Handlers ──
  async function handleClientSelect(client: Client) {
    selectedClient = client;
    error = '';
    loadingTraffic = true;
    try {
      trafficData = await fetchTraffic(client.id, period, startDate, endDate);
    } catch (err: unknown) {
      error = err instanceof Error ? err.message : 'Falha ao carregar dados';
    } finally {
      loadingTraffic = false;
    }
  }

  async function handlePeriodChange(newPeriod: number, start: string = '', end: string = '') {
    period = newPeriod;
    startDate = start;
    endDate = end;
    
    let isTooLong = period > 180;
    if (startDate && endDate) {
      const s = new Date(startDate + 'T00:00:00');
      const e = new Date(endDate + 'T00:00:00');
      const diffDays = Math.ceil((e.getTime() - s.getTime()) / (1000 * 60 * 60 * 24));
      if (diffDays > 180) isTooLong = true;
    }
    
    if (isTooLong) comparing = false;
    
    if (selectedClient) await handleClientSelect(selectedClient);
  }

  let comparisonDisabled = $derived.by(() => {
    let isTooLong = period > 180;
    if (startDate && endDate) {
      const s = new Date(startDate + 'T00:00:00');
      const e = new Date(endDate + 'T00:00:00');
      const diffDays = Math.ceil((e.getTime() - s.getTime()) / (1000 * 60 * 60 * 24));
      if (diffDays > 180) isTooLong = true;
    }
    return isTooLong;
  });

  function handleLocationChange(newLocation: string) {
    location = newLocation;
  }
</script>

<svelte:head>
  <title>{selectedClient ? `${selectedClient.name} — Análise de Tráfego` : 'Análise de Tráfego'}</title>
</svelte:head>

<!-- ── Controles ── -->
<div class="card ctrl">
  <div class="ctrl-row">
    <ClientSelector {clients} selected={selectedClient} onselect={handleClientSelect} loading={loadingClients} />
    <Filters {period} {startDate} {endDate} {location} onperiodchange={handlePeriodChange} onlocationchange={handleLocationChange} />
    <div class="field">
      <span class="field-label">&nbsp;</span>
      <button
        class="btn-compare"
        class:active={comparing}
        disabled={comparisonDisabled}
        onclick={() => {
          if (!comparisonDisabled) comparing = !comparing;
        }}
        title={comparisonDisabled ? "Apenas até 6 meses" : "Comparar com período anterior"}
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 3h5v5"/><path d="M8 3H3v5"/><path d="M21 3l-7 7"/><path d="M3 3l7 7"/><path d="M16 21h5v-5"/><path d="M8 21H3v-5"/><path d="M21 21l-7-7"/><path d="M3 21l7-7"/></svg>
        Comparar
      </button>
    </div>
    {#if loadingTraffic}
      <div style="display: flex; align-items: center; gap: 6px; font-size: 13px; color: var(--ink-subtle);">
        <span class="spinner"></span> Carregando dados...
      </div>
    {/if}
  </div>
  {#if error}
    <div class="ctrl-error">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><path d="M12 9v4"/><path d="M12 17h.01"/></svg>
      {error}
    </div>
  {/if}
</div>

<!-- ── Dashboard ── -->
{#if !selectedClient}
  <div class="card card--full"><EmptyState /></div>
{:else if hasData && trafficData}
  <div class="card card--full" style="transition: opacity 0.4s cubic-bezier(0.16, 1, 0.3, 1); opacity: {loadingTraffic ? 0.5 : 1}; pointer-events: {loadingTraffic ? 'none' : 'auto'}">
    <!-- Source Tabs com SVG icons -->
    <div class="src-tabs">
      <button class="src-tab" class:active={activeTab === 'overview'} onclick={() => activeTab = 'overview'}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
        Overview
      </button>
      <button class="src-tab" class:active={activeTab === 'gsc'} onclick={() => activeTab = 'gsc'}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
        Search Console
      </button>
      <button class="src-tab" class:active={activeTab === 'ga4'} onclick={() => activeTab = 'ga4'}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v18h18"/><path d="M18 17V9"/><path d="M13 17V5"/><path d="M8 17v-3"/></svg>
        Analytics 4
      </button>
      <button class="src-tab" class:active={activeTab === 'bing'} onclick={() => activeTab = 'bing'}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 2v20l6-3 6 3 4-2V4l-4 2-6-3-6 3z"/></svg>
        Bing
      </button>

      <!-- Expand Button -->
      <button class="src-tab expand-btn" onclick={toggleExpand} title={isExpanded ? "Restaurar container" : "Expandir container"}>
        {#if isExpanded}
          <!-- Minimize icon -->
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 3v3a2 2 0 0 1-2 2H3"/><path d="M21 8h-3a2 2 0 0 1-2-2V3"/><path d="M3 16h3a2 2 0 0 1 2 2v3"/><path d="M16 21v-3a2 2 0 0 1 2-2h3"/></svg>
        {:else}
          <!-- Maximize icon -->
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h6v6"/><path d="M9 21H3v-6"/><path d="M21 3l-7 7"/><path d="M3 21l7-7"/></svg>
        {/if}
      </button>
    </div>

    <!-- Panel content -->
    {#if activeTab === 'overview'}
      <div class="panel">
        <OverviewPanel {trafficData} {comparing} />
      </div>
    {:else if activeTab === 'gsc'}
      {#if trafficData.gsc_chart && trafficData.gsc_chart.length > 0}
        <div style="padding: var(--s-5) var(--s-6) 0;">
          <GscChart data={trafficData.gsc_chart} prevData={comparing ? trafficData.gsc_chart_prev : []} {comparing} />
        </div>
      {/if}
      <SearchConsolePanel queries={trafficData.gsc_queries || []} pages={trafficData.gsc_pages || []} {comparing} />
    {:else if activeTab === 'ga4'}
      <GA4Panel data={trafficData.ga4 || []} />
    {:else if activeTab === 'bing'}
      <BingPanel data={trafficData.bing || []} />
    {/if}
  </div>
{:else if selectedClient && !loadingTraffic}
  <div class="card card--full">
    <EmptyState message={`Nenhum dado encontrado para ${selectedClient.name}. Sincronize os dados primeiro.`} />
  </div>
{/if}
