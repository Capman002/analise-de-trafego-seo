<script lang="ts">
  interface Metric { label: string; value: string | number; highlight?: boolean; delta?: number | null; deltaInverted?: boolean; prevValue?: string | number | null; }
  interface Props { primary: string; secondary: string; metrics: Metric[]; trend?: 'up' | 'down' | 'flat'; rank?: number; hideLabels?: boolean; }
  let { primary, secondary, metrics, trend, rank, hideLabels = false }: Props = $props();
  let trendColor = $derived(trend === 'up' ? '#2d5f3f' : trend === 'down' ? '#b4422a' : '#8b8677');
  let trendBg = $derived(trend === 'up' ? '#dce8df' : trend === 'down' ? '#f7e1d6' : '#f2ecdb');
</script>

<div class="list-item">
  {#if rank !== undefined}
    <div class="list-item__rank" class:top={rank <= 3}>{rank}</div>
  {/if}
  <div class="list-item__body">
    <div class="list-item__primary" title={primary}>{primary}</div>
    <div class="list-item__secondary" title={secondary}>{secondary}</div>
  </div>
  <div class="list-item__metrics">
    {#each metrics as m, i (i)}
      <div class="list-item__metric">
        <div class="list-item__metric-val" style={m.highlight ? `font-weight:600;color:${trendColor}` : ''}>{m.value}</div>
        <div class="list-item__metric-lbl">
          {#if !hideLabels}
            {m.label}
          {/if}
          {#if m.delta !== undefined && m.delta !== null}
            {@const isPos = m.deltaInverted ? m.delta < 0 : m.delta > 0}
            {@const isNeg = m.deltaInverted ? m.delta > 0 : m.delta < 0}
            {@const color = isPos ? 'var(--ok)' : isNeg ? 'var(--danger)' : 'var(--ink-muted)'}
            <div style="display: flex; align-items: center; justify-content: flex-end; gap: 4px; margin-top: 2px;">
              {#if m.prevValue !== undefined && m.prevValue !== null}
                <span style="color: var(--ink-muted); font-size: 10px; font-weight: 500; letter-spacing: -0.2px;">
                  {m.prevValue}
                </span>
              {/if}
              <span style="color: {color}; font-weight: 700;">
                {m.delta > 0 ? '+' : ''}{m.delta.toFixed(1)}%
              </span>
            </div>
          {/if}
        </div>
      </div>
    {/each}
  </div>
  {#if trend}
    <div class="list-item__trend" style="background:{trendBg};color:{trendColor}">
      {#if trend === 'up'}
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 7 13.5 15.5 8.5 10.5 2 17"/><polyline points="16 7 22 7 22 13"/></svg>
      {:else if trend === 'down'}
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 17 13.5 8.5 8.5 13.5 2 7"/><polyline points="16 17 22 17 22 11"/></svg>
      {:else}
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/></svg>
      {/if}
    </div>
  {/if}
</div>

<style>
  .list-item { display: flex; align-items: center; padding: 12px 20px; border-bottom: 1px solid var(--line-soft); gap: 16px; }
  .list-item__rank { width: 28px; height: 28px; border-radius: 5px; background: var(--surface-3); display: flex; align-items: center; justify-content: center; font-size: 13px; font-weight: 700; color: var(--ink-subtle); flex-shrink: 0; }
  .list-item__rank.top { background: linear-gradient(135deg, #b87333, #c47d3a); color: #fff; }
  .list-item__body { flex: 1; min-width: 0; }
  .list-item__primary { font-size: 15px; font-weight: 600; color: var(--ink); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin-bottom: 2px; }
  .list-item__secondary { font-size: 12px; color: var(--ink-subtle); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .list-item__metrics { display: flex; gap: 20px; flex-shrink: 0; }
  .list-item__metric { text-align: right; width: 90px; flex-shrink: 0; }
  .list-item__metric-val { font-size: 15px; color: var(--ink); font-variant-numeric: tabular-nums; }
  .list-item__metric-lbl { font-size: 11px; font-weight: 500; color: var(--ink-subtle); text-transform: uppercase; margin-top: 2px; }
  .list-item__trend { width: 28px; height: 28px; border-radius: 5px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
</style>
