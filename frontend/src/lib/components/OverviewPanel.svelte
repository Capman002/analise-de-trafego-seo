<script lang="ts">
  import MetricCard from './MetricCard.svelte';
  import SectionCard from './SectionCard.svelte';
  import Modal from './Modal.svelte';
  import ModalList from './ModalList.svelte';
  import ListItem from './ListItem.svelte';
  import EmptyState from './EmptyState.svelte';
  import type { TrafficResponse } from '$lib/types';

  interface Props { trafficData: TrafficResponse; comparing?: boolean; }
  let { trafficData, comparing = false }: Props = $props();

  type SectionId = 'topPages' | 'oppTop5' | 'oppPage1' | 'riseQueries' | 'risePages' | 'fallQueries' | 'fallPages' | 'topItems' | 'posTop3' | 'posTop10' | 'posTop20' | 'posBeyond20';
  let openSection = $state<SectionId | null>(null);

  function formatPeriod(days: number): string {
    if (days === 7) return '7 dias';
    if (days === 28) return '28 dias';
    if (days === 90) return '3 meses';
    if (days === 180) return '6 meses';
    if (days === 365) return '12 meses';
    if (days === 480) return '16 meses';
    return `${days} dias`;
  }

  let p = $derived(trafficData?.period || 28);
  let pStart = $derived(trafficData?.startDate || '');
  let pEnd = $derived(trafficData?.endDate || '');

  let timeText = $derived.by(() => {
    if (pStart && pEnd && p === 0) {
      const s = new Date(pStart + 'T00:00:00');
      const e = new Date(pEnd + 'T00:00:00');
      const formatted = `${s.toLocaleDateString('pt-BR', {day:'2-digit', month:'short'})} a ${e.toLocaleDateString('pt-BR', {day:'2-digit', month:'short'})}`;
      return `(${formatted})`;
    }
    return comparing ? `(Últimos ${formatPeriod(p)} vs período anterior)` : `(Últimos ${formatPeriod(p)})`;
  });

  let gscPages = $derived(trafficData?.gsc_pages || []);
  let gscQueries = $derived(trafficData?.gsc_queries || []);
  let overview = $derived(trafficData?.overview);
  let overviewPrev = $derived(trafficData?.overview_prev);
  let posDist = $derived(trafficData?.position_distribution);

  let topPages = $derived([...gscPages].sort((a: any, b: any) => b.clicks - a.clicks).slice(0, 100));
  let oppTop5 = $derived.by(() => {
    return gscQueries.filter((q: any) => {
      const pos = q.position;
      const prevPos = q.prev_position;
      if (pos > 0 && pos > 5 && pos <= 10) return true;
      if (comparing && (!pos || pos <= 0) && prevPos > 0 && prevPos > 5 && prevPos <= 10) return true;
      return false;
    })
    .sort((a: any, b: any) => {
      const aVal = a.impressions || a.prev_impressions || 0;
      const bVal = b.impressions || b.prev_impressions || 0;
      return bVal - aVal;
    })
    .slice(0, 100);
  });
  let oppPage1 = $derived.by(() => {
    return gscQueries.filter((q: any) => {
      const pos = q.position;
      const prevPos = q.prev_position;
      if (pos > 0 && pos > 10 && pos <= 20) return true;
      if (comparing && (!pos || pos <= 0) && prevPos > 0 && prevPos > 10 && prevPos <= 20) return true;
      return false;
    })
    .sort((a: any, b: any) => {
      const aVal = a.impressions || a.prev_impressions || 0;
      const bVal = b.impressions || b.prev_impressions || 0;
      return bVal - aVal;
    })
    .slice(0, 100);
  });

  let riseQueries = $derived(trafficData?.gsc_rise_queries || []);
  let risePages = $derived(trafficData?.gsc_rise_pages || []);
  let fallQueries = $derived(trafficData?.gsc_fall_queries || []);
  let fallPages = $derived(trafficData?.gsc_fall_pages || []);

  let ga4Items = $derived(trafficData?.ga4 || []);
  let topItems = $derived([...ga4Items].sort((a: any, b: any) => Number(b.revenue || 0) - Number(a.revenue || 0)).slice(0, 100));
  let totalRevenue = $derived(ga4Items.reduce((sum: number, item: any) => sum + Number(item.revenue || 0), 0));
  let totalSales = $derived(ga4Items.reduce((sum: number, item: any) => sum + Number(item.items_purchased || item.itemsPurchased || 0), 0));

  let hasGsc = $derived(gscPages.length > 0 || gscQueries.length > 0);
  let hasGa4 = $derived(ga4Items.length > 0);

  // Deltas (só calculados quando comparing é true)
  function calcDelta(current: number | null | undefined, prev: number | null | undefined): number | null {
    if (!comparing || prev == null) return null;
    const cur = current || 0;
    if (prev === 0) return cur > 0 ? 100 : 0;
    return ((cur - prev) / Math.abs(prev)) * 100;
  }

  let dClicks = $derived(calcDelta(overview?.gsc_clicks || 0, overviewPrev?.gsc_clicks || 0));
  let dImpressions = $derived(calcDelta(overview?.gsc_impressions || 0, overviewPrev?.gsc_impressions || 0));
  let dCtr = $derived(comparing && overviewPrev ? ((overview?.gsc_ctr || 0) - (overviewPrev?.gsc_ctr || 0)) / Math.max(overviewPrev?.gsc_ctr || 1, 0.0001) * 100 : null);
  let dPosition = $derived(comparing && overviewPrev ? ((overview?.gsc_position || 0) - (overviewPrev?.gsc_position || 0)) / Math.max(overviewPrev?.gsc_position || 1, 0.0001) * 100 : null);
  let dRevenue = $derived(calcDelta(totalRevenue, overviewPrev?.ga4_revenue || 0));
  let dSales = $derived(calcDelta(totalSales, overviewPrev?.ga4_items_purchased || 0));

  // Position distribution helpers
  let posTotal = $derived((posDist?.top3 || 0) + (posDist?.top10 || 0) + (posDist?.top20 || 0) + (posDist?.beyond20 || 0));
  const posColors = ['#22c55e', '#3b82f6', '#f59e0b', '#8b8677'];
  const posLabels = ['Top 3', 'Top 4-10', 'Top 11-20', '20+'];

  function posPercent(count: number): number {
    return posTotal > 0 ? (count / posTotal) * 100 : 0;
  }

  // Drill-down: queries filtradas por faixa
  function queriesByRange(min: number, max: number): any[] {
    return gscQueries.filter((q: any) => {
      const pos = q.position;
      const prevPos = q.prev_position;
      
      // Se existe posição atual e ela cai na faixa
      if (pos > 0 && pos >= min && pos < max) return true;
      
      // Se não existe posição atual (ou seja, 0 / null / undefined) e estamos comparando,
      // e a posição anterior cai na faixa
      if (comparing && (!pos || pos <= 0) && prevPos > 0 && prevPos >= min && prevPos < max) return true;
      
      return false;
    })
    .sort((a: any, b: any) => {
      const aVal = a.impressions || a.prev_impressions || 0;
      const bVal = b.impressions || b.prev_impressions || 0;
      return bVal - aVal;
    })
    .slice(0, 100);
  }

  let posTop3Queries = $derived(queriesByRange(0, 3.5));
  let posTop10Queries = $derived(queriesByRange(3.5, 10.5));
  let posTop20Queries = $derived(queriesByRange(10.5, 20.5));
  let posBeyond20Queries = $derived(queriesByRange(20.5, 999));

  function fmt(n: number): string { return n.toLocaleString('pt-BR'); }
  function fmtMoney(n: number): string { return `R$ ${n.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}`; }
  function fmtPct(n: number): string { return (n * 100).toFixed(2) + '%'; }

  // Sort options reutilizáveis
  const SORT_GSC = [
    { key: 'clicks', label: 'Cliques' },
    { key: 'impressions', label: 'Impressões' },
    { key: 'ctr', label: 'CTR' },
    { key: 'position', label: 'Posição' },
  ];
  const SORT_GA4 = [
    { key: 'revenue', label: 'Receita' },
    { key: 'itemsPurchased', label: 'Vendas' },
  ];
</script>

{#if !hasGsc && !hasGa4}
  <EmptyState message="Selecione um cliente para visualizar o dashboard de resumo." />
{:else}
  <div style="padding: 0; background: transparent;">
    <!-- KPI Metric Cards -->
    {#if hasGsc || hasGa4}
      <div class="metrics-row">
        {#if hasGsc}
          <MetricCard label="Total de Cliques" value={fmt(overview?.gsc_clicks || 0)} subValue="GSC" color="#3b82f6" delta={dClicks}>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 10v4"/><path d="m12 7-3 3 3 3"/><rect x="2" y="6" width="20" height="12" rx="2"/></svg>{/snippet}
          </MetricCard>
          <MetricCard label="Impressões" value={fmt(overview?.gsc_impressions || 0)} subValue="GSC" color="#8b5cf6" delta={dImpressions}>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/><circle cx="12" cy="12" r="3"/></svg>{/snippet}
          </MetricCard>
          <MetricCard label="CTR Médio" value={fmtPct(overview?.gsc_ctr || 0)} subValue="Click-through rate" color="#10b981" delta={dCtr}>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>{/snippet}
          </MetricCard>
          <MetricCard label="Posição Média" value={(overview?.gsc_position || 0).toFixed(1)} subValue="Média geral" color="#f59e0b" delta={dPosition} deltaInverted>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v18h18"/><path d="M18 17V9"/><path d="M13 17V5"/><path d="M8 17v-3"/></svg>{/snippet}
          </MetricCard>
        {/if}
        {#if hasGa4}
          <MetricCard label="Receita Total" value={fmtMoney(totalRevenue)} subValue="Produtos orgânicos" color="#10b981" delta={dRevenue}>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>{/snippet}
          </MetricCard>
          <MetricCard label="Vendas" value={fmt(totalSales)} subValue="Unidades" color="#ec4899" delta={dSales}>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m7.5 4.27 9 5.15"/><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"/><path d="m3.3 7 8.7 5 8.7-5"/><path d="M12 22V12"/></svg>{/snippet}
          </MetricCard>
        {/if}
      </div>
    {/if}

    <!-- Position Distribution -->
    {#if hasGsc && posDist && posTotal > 0}
      <div class="pos-dist" style="--pos-cols: {comparing ? '70px 1fr 55px 45px' : '70px 1fr 55px'}">
        <div class="pos-dist__title">Distribuição de Posições</div>
        {#each [
          { label: posLabels[0], count: posDist.top3, prev: posDist.prev_top3, color: posColors[0], section: 'posTop3' as SectionId },
          { label: posLabels[1], count: posDist.top10, prev: posDist.prev_top10, color: posColors[1], section: 'posTop10' as SectionId },
          { label: posLabels[2], count: posDist.top20, prev: posDist.prev_top20, color: posColors[2], section: 'posTop20' as SectionId },
          { label: posLabels[3], count: posDist.beyond20, prev: posDist.prev_beyond20, color: posColors[3], section: 'posBeyond20' as SectionId },
        ] as row}
          <button class="pos-dist__row" onclick={() => openSection = row.section} type="button">
            <span class="pos-dist__label">{row.label}</span>
            <div class="pos-dist__bar-wrap">
              <div class="pos-dist__bar" style="width: {posPercent(row.count)}%; background: {row.color};"></div>
            </div>
            <span class="pos-dist__count">{fmt(row.count)}</span>
            {#if comparing}
              {@const diff = row.count - row.prev}
              <span class="pos-dist__delta" style="color: {diff > 0 ? 'var(--ok)' : diff < 0 ? 'var(--danger)' : 'var(--ink-muted)'}">
                {diff > 0 ? '+' : ''}{diff}
              </span>
            {/if}
          </button>
        {/each}
      </div>
    {/if}

    <!-- Section Cards Grid -->
    {#if hasGsc || hasGa4}
      <div class="section-grid">
        {#if hasGsc}
          <SectionCard title="Top Páginas" subtitle="Páginas com mais cliques absolutos" count={topPages.length} color="#f59e0b" trend="up" onclick={() => openSection = 'topPages'}>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9H4.5a2.5 2.5 0 0 1 0-5C7.4 4 9 7 9 7"/><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5C16.6 4 15 7 15 7"/><path d="M4 22h16"/><path d="M10 14.66V17c0 .55-.47.98-.97 1.21C7.85 18.75 7 20.24 7 22"/><path d="M14 14.66V17c0 .55.47.98.97 1.21C16.15 18.75 17 20.24 17 22"/><path d="M18 2H6v7a6 6 0 0 0 12 0V2Z"/></svg>{/snippet}
          </SectionCard>
          <SectionCard title="Oportunidade Top 5" subtitle="Consultas nas posições 6-10" count={oppTop5.length} color="#3b82f6" trend="up" onclick={() => openSection = 'oppTop5'}>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/></svg>{/snippet}
          </SectionCard>
          <SectionCard title="Oportunidade 1ª Página" subtitle="Consultas nas posições 11-20" count={oppPage1.length} color="#8b5cf6" trend="up" onclick={() => openSection = 'oppPage1'}>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/></svg>{/snippet}
          </SectionCard>
          <SectionCard title="Maiores Altas — Termos" subtitle="Termos que mais ganharam cliques" count={riseQueries.length} color="#22c55e" trend="up" onclick={() => openSection = 'riseQueries'}>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M7 7h10v10"/><path d="M7 17 17 7"/></svg>{/snippet}
          </SectionCard>
          <SectionCard title="Maiores Altas — Páginas" subtitle="Páginas que mais ganharam cliques" count={risePages.length} color="#22c55e" trend="up" onclick={() => openSection = 'risePages'}>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M7 7h10v10"/><path d="M7 17 17 7"/></svg>{/snippet}
          </SectionCard>
          <SectionCard title="Maiores Quedas — Termos" subtitle="Termos que mais perderam cliques" count={fallQueries.length} color="#ef4444" trend="down" onclick={() => openSection = 'fallQueries'}>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 17 13.5 8.5 8.5 13.5 2 7"/><polyline points="16 17 22 17 22 11"/></svg>{/snippet}
          </SectionCard>
          <SectionCard title="Maiores Quedas — Páginas" subtitle="Páginas que mais perderam cliques" count={fallPages.length} color="#ef4444" trend="down" onclick={() => openSection = 'fallPages'}>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M16 13H8"/><path d="M16 17H8"/><path d="M10 9H8"/></svg>{/snippet}
          </SectionCard>
        {/if}
        {#if hasGa4}
          <SectionCard title="Top Produtos por Receita" subtitle="Produtos com maior faturamento orgânico" count={topItems.length} color="#10b981" trend="up" onclick={() => openSection = 'topItems'}>
            {#snippet icon()}<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m7.5 4.27 9 5.15"/><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"/><path d="m3.3 7 8.7 5 8.7-5"/><path d="M12 22V12"/></svg>{/snippet}
          </SectionCard>
        {/if}
      </div>
    {/if}

    <!-- Modais existentes -->
    <Modal open={openSection === 'topPages'} title="Top Páginas" subtitle={`Páginas com mais cliques absolutos no período ${timeText}`} color="#f59e0b" onclose={() => openSection = null}>
      <ModalList items={topPages} sortOptions={SORT_GSC} defaultSort="clicks" emptyMessage="Nenhuma página encontrada.">
        {#snippet renderItem(page: any, i: number, sortKey: string)}
          <ListItem hideLabels={true} rank={i + 1} primary={page.key?.replace(/^https?:\/\/[^\/]+/, '') || '/'} secondary={page.key || ''} trend="up" metrics={[
            { label: 'Cliques', value: fmt(page.clicks || 0), highlight: sortKey === 'clicks', delta: calcDelta(page.clicks, page.prev_clicks), prevValue: fmt(page.prev_clicks || 0) },
            { label: 'Impressões', value: fmt(page.impressions || 0), highlight: sortKey === 'impressions', delta: calcDelta(page.impressions, page.prev_impressions), prevValue: fmt(page.prev_impressions || 0) },
            { label: 'CTR', value: fmtPct(page.ctr || 0), highlight: sortKey === 'ctr', delta: calcDelta(page.ctr, page.prev_ctr), prevValue: fmtPct(page.prev_ctr || 0) },
            { label: 'Posição', value: page.position ? page.position.toFixed(1) : '-', highlight: sortKey === 'position', delta: calcDelta(page.position, page.prev_position), deltaInverted: true, prevValue: page.prev_position ? page.prev_position.toFixed(1) : '-' },
          ]} />
        {/snippet}
      </ModalList>
    </Modal>

    <Modal open={openSection === 'oppTop5'} title="Oportunidade para Top 5" subtitle={`Consultas nas posições 6-10 com alto potencial ${timeText}`} color="#3b82f6" onclose={() => openSection = null}>
      <ModalList items={oppTop5} sortOptions={SORT_GSC} defaultSort="impressions" emptyMessage="Nenhuma oportunidade nesta faixa.">
        {#snippet renderItem(q: any, i: number, sortKey: string)}
          <ListItem hideLabels={true} rank={i + 1} primary={q.key} secondary={'Posição média: ' + (q.position ? q.position.toFixed(1) : '-')} trend="up" metrics={[
            { label: 'Cliques', value: fmt(q.clicks || 0), highlight: sortKey === 'clicks', delta: calcDelta(q.clicks, q.prev_clicks), prevValue: fmt(q.prev_clicks || 0) },
            { label: 'Impressões', value: fmt(q.impressions || 0), highlight: sortKey === 'impressions', delta: calcDelta(q.impressions, q.prev_impressions), prevValue: fmt(q.impressions || 0) },
            { label: 'CTR', value: fmtPct(q.ctr || 0), highlight: sortKey === 'ctr', delta: calcDelta(q.ctr, q.prev_ctr), prevValue: fmtPct(q.prev_ctr || 0) },
            { label: 'Posição', value: q.position ? q.position.toFixed(1) : '-', highlight: sortKey === 'position', delta: calcDelta(q.position, q.prev_position), deltaInverted: true, prevValue: q.prev_position ? q.prev_position.toFixed(1) : '-' },
          ]} />
        {/snippet}
      </ModalList>
    </Modal>

    <Modal open={openSection === 'oppPage1'} title="Oportunidade para 1ª Página" subtitle={`Consultas nas posições 11-20 ${timeText}`} color="#8b5cf6" onclose={() => openSection = null}>
      <ModalList items={oppPage1} sortOptions={SORT_GSC} defaultSort="impressions" emptyMessage="Nenhuma oportunidade nesta faixa.">
        {#snippet renderItem(q: any, i: number, sortKey: string)}
          <ListItem hideLabels={true} rank={i + 1} primary={q.key} secondary={'Posição média: ' + (q.position ? q.position.toFixed(1) : '-')} trend="up" metrics={[
            { label: 'Cliques', value: fmt(q.clicks || 0), highlight: sortKey === 'clicks', delta: calcDelta(q.clicks, q.prev_clicks), prevValue: fmt(q.prev_clicks || 0) },
            { label: 'Impressões', value: fmt(q.impressions || 0), highlight: sortKey === 'impressions', delta: calcDelta(q.impressions, q.prev_impressions), prevValue: fmt(q.prev_impressions || 0) },
            { label: 'CTR', value: fmtPct(q.ctr || 0), highlight: sortKey === 'ctr', delta: calcDelta(q.ctr, q.prev_ctr), prevValue: fmtPct(q.prev_ctr || 0) },
            { label: 'Posição', value: q.position ? q.position.toFixed(1) : '-', highlight: sortKey === 'position', delta: calcDelta(q.position, q.prev_position), deltaInverted: true, prevValue: q.prev_position ? q.prev_position.toFixed(1) : '-' },
          ]} />
        {/snippet}
      </ModalList>
    </Modal>

    <Modal open={openSection === 'riseQueries'} title="Maiores Altas — Termos" subtitle={`Termos que mais ganharam cliques ${timeText}`} color="#22c55e" onclose={() => openSection = null}>
      <ModalList items={riseQueries} emptyMessage="Nenhuma alta registrada.">
        {#snippet renderItem(q: any, i: number, sortKey: string)}
          <ListItem rank={i + 1} primary={q.key} secondary={`Anterior: ${fmt(q.prev_clicks || 0)}`} trend="up" metrics={[
            { label: 'Delta', value: '+' + fmt(q.clicks_diff || 0), highlight: true },
            { label: 'Cliques (Atual)', value: fmt(q.clicks || 0) },
          ]} />
        {/snippet}
      </ModalList>
    </Modal>
    <Modal open={openSection === 'risePages'} title="Maiores Altas — Páginas" subtitle={`Páginas que mais ganharam cliques ${timeText}`} color="#22c55e" onclose={() => openSection = null}>
      <ModalList items={risePages} emptyMessage="Nenhuma alta registrada.">
        {#snippet renderItem(p: any, i: number, sortKey: string)}
          <ListItem rank={i + 1} primary={p.key?.replace(/^https?:\/\/[^\/]+/, '') || '/'} secondary={`Anterior: ${fmt(p.prev_clicks || 0)}`} trend="up" metrics={[
            { label: 'Delta', value: '+' + fmt(p.clicks_diff || 0), highlight: true },
            { label: 'Cliques (Atual)', value: fmt(p.clicks || 0) },
          ]} />
        {/snippet}
      </ModalList>
    </Modal>
    <Modal open={openSection === 'fallQueries'} title="Maiores Quedas — Termos" subtitle={`Termos que mais perderam cliques ${timeText}`} color="#ef4444" onclose={() => openSection = null}>
      <ModalList items={fallQueries} emptyMessage="Nenhuma queda registrada.">
        {#snippet renderItem(q: any, i: number, sortKey: string)}
          <ListItem rank={i + 1} primary={q.key} secondary={`Anterior: ${fmt(q.prev_clicks || 0)}`} trend="down" metrics={[
            { label: 'Delta', value: fmt(q.clicks_diff || 0), highlight: true },
            { label: 'Cliques (Atual)', value: fmt(q.clicks || 0) },
          ]} />
        {/snippet}
      </ModalList>
    </Modal>
    <Modal open={openSection === 'fallPages'} title="Maiores Quedas — Páginas" subtitle={`Páginas que mais perderam cliques ${timeText}`} color="#ef4444" onclose={() => openSection = null}>
      <ModalList items={fallPages} emptyMessage="Nenhuma queda registrada.">
        {#snippet renderItem(p: any, i: number, sortKey: string)}
          <ListItem rank={i + 1} primary={p.key?.replace(/^https?:\/\/[^\/]+/, '') || '/'} secondary={`Anterior: ${fmt(p.prev_clicks || 0)}`} trend="down" metrics={[
            { label: 'Delta', value: fmt(p.clicks_diff || 0), highlight: true },
            { label: 'Cliques (Atual)', value: fmt(p.clicks || 0) },
          ]} />
        {/snippet}
      </ModalList>
    </Modal>

    <Modal open={openSection === 'topItems'} title="Top Produtos por Receita" subtitle={`Produtos com maior faturamento orgânico (GA4) ${timeText}`} color="#10b981" onclose={() => openSection = null}>
      <ModalList items={topItems} sortOptions={SORT_GA4} defaultSort="revenue" emptyMessage="Nenhum produto encontrado.">
        {#snippet renderItem(item: any, i: number, sortKey: string)}
          <ListItem hideLabels={true} rank={i + 1} primary={item.itemName || item.item_name || '(sem nome)'} secondary={`${fmt(item.itemsPurchased || item.items_purchased || 0)} vendas`} trend="up" metrics={[
            { label: 'Receita', value: fmtMoney(Number(item.revenue || item.itemRevenue || 0)), highlight: sortKey === 'revenue', delta: calcDelta(item.revenue, item.prev_revenue), prevValue: fmtMoney(Number(item.prev_revenue || 0)) },
            { label: 'Vendas', value: fmt(Number(item.itemsPurchased || item.items_purchased || 0)), highlight: sortKey === 'itemsPurchased', delta: calcDelta(item.itemsPurchased, item.prev_itemsPurchased), prevValue: fmt(Number(item.prev_itemsPurchased || 0)) },
          ]} />
        {/snippet}
      </ModalList>
    </Modal>

    <!-- Modais de drill-down de distribuição de posições -->
    {#each [
      { id: 'posTop3' as SectionId, title: 'Queries Top 3', subtitle: `Consultas nas posições 1-3 ${timeText}`, color: '#22c55e', items: posTop3Queries },
      { id: 'posTop10' as SectionId, title: 'Queries Posição 4-10', subtitle: `Consultas nas posições 4-10 ${timeText}`, color: '#3b82f6', items: posTop10Queries },
      { id: 'posTop20' as SectionId, title: 'Queries Posição 11-20', subtitle: `Consultas nas posições 11-20 ${timeText}`, color: '#f59e0b', items: posTop20Queries },
      { id: 'posBeyond20' as SectionId, title: 'Queries Posição 20+', subtitle: `Consultas com posição acima de 20 ${timeText}`, color: '#8b8677', items: posBeyond20Queries },
    ] as modal}
      <Modal open={openSection === modal.id} title={modal.title} subtitle={modal.subtitle} color={modal.color} onclose={() => openSection = null}>
        <ModalList items={modal.items} sortOptions={SORT_GSC} defaultSort="impressions" emptyMessage="Nenhuma query nesta faixa.">
          {#snippet renderItem(q: any, i: number, sortKey: string)}
            <ListItem hideLabels={true} rank={i + 1} primary={q.key} secondary={'Posição: ' + (q.position ? q.position.toFixed(1) : '-')} trend="up" metrics={[
              { label: 'Cliques', value: fmt(q.clicks || 0), highlight: sortKey === 'clicks', delta: calcDelta(q.clicks, q.prev_clicks), prevValue: fmt(q.prev_clicks || 0) },
              { label: 'Impressões', value: fmt(q.impressions || 0), highlight: sortKey === 'impressions', delta: calcDelta(q.impressions, q.prev_impressions), prevValue: fmt(q.prev_impressions || 0) },
              { label: 'CTR', value: fmtPct(q.ctr || 0), highlight: sortKey === 'ctr', delta: calcDelta(q.ctr, q.prev_ctr), prevValue: fmtPct(q.prev_ctr || 0) },
              { label: 'Posição', value: q.position ? q.position.toFixed(1) : '-', highlight: sortKey === 'position', delta: calcDelta(q.position, q.prev_position), deltaInverted: true, prevValue: q.prev_position ? q.prev_position.toFixed(1) : '-' },
            ]} />
          {/snippet}
        </ModalList>
      </Modal>
    {/each}
  </div>
{/if}
