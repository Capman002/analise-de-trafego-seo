<script lang="ts">
  import type { Snippet } from 'svelte';

  interface SortOption { key: string; label: string; }
  interface Props {
    items: any[];
    pageSize?: number;
    maxItems?: number;
    emptyMessage?: string;
    renderItem: Snippet<[any, number, string]>;
    sortOptions?: SortOption[];
    defaultSort?: string;
    defaultDir?: 'asc' | 'desc';
  }
  let {
    items, pageSize = 10, maxItems = 100,
    emptyMessage = 'Nenhum item encontrado.',
    renderItem, sortOptions, defaultSort, defaultDir = 'desc'
  }: Props = $props();

  let sortKey = $state(defaultSort || '');
  let sortDir = $state<'asc' | 'desc'>(defaultDir);
  let visible = $state(pageSize);

  // Reset visible quando items mudam
  $effect(() => { items; visible = pageSize; });

  function toggleSort(key: string) {
    if (sortKey === key) {
      sortDir = sortDir === 'desc' ? 'asc' : 'desc';
    } else {
      sortKey = key;
      sortDir = 'desc';
    }
    visible = pageSize; // reset paginação ao mudar sort
  }

  let sorted = $derived(() => {
    if (!sortKey) return items;
    return [...items].sort((a, b) => {
      const va = Number(a[sortKey] || 0);
      const vb = Number(b[sortKey] || 0);
      return sortDir === 'desc' ? vb - va : va - vb;
    });
  });

  let capped = $derived(sorted().slice(0, maxItems));
  let shown = $derived(capped.slice(0, visible));
  let hasMore = $derived(visible < capped.length);
  function loadMore() { visible = Math.min(visible + pageSize, capped.length); }
</script>

{#if capped.length === 0}
  <div style="padding: 40px 24px; text-align: center; color: var(--ink-subtle); font-size: 13px;">{emptyMessage}</div>
{:else}
  <div>
    {#if sortOptions && sortOptions.length > 0}
      <div class="sort-bar">
        <div class="sort-bar__rank"></div>
        <div class="sort-bar__body"></div>
        <div class="sort-bar__metrics">
          {#each sortOptions as opt (opt.key)}
            <button
              class="sort-pill"
              class:active={sortKey === opt.key}
              onclick={() => toggleSort(opt.key)}
              type="button"
            >
              <span>{opt.label}</span>
              {#if sortKey === opt.key}
                <span style="position: absolute; left: 100%; margin-left: 4px; display: flex; align-items: center;">
                  <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                    {#if sortDir === 'desc'}
                      <path d="m6 9 6 6 6-6"/>
                    {:else}
                      <path d="m18 15-6-6-6 6"/>
                    {/if}
                  </svg>
                </span>
              {/if}
            </button>
          {/each}
        </div>
        <div class="sort-bar__trend"></div>
      </div>
    {/if}
    {#each shown as item, i (i)}
      {@render renderItem(item, i, sortKey)}
    {/each}
    {#if hasMore}
      <div style="display: flex; justify-content: center; padding: 16px; border-top: 1px solid var(--line-soft); background: var(--surface-2);">
        <button onclick={loadMore} style="padding: 8px 20px; border: 1px solid var(--line-soft); border-radius: 5px; background: var(--surface); color: var(--ink); font-size: 12px; font-weight: 500; cursor: pointer; display: flex; align-items: center; gap: 6px;">
          Ver mais ({capped.length - visible} restantes)
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>
        </button>
      </div>
    {/if}
    <div style="padding: 10px 20px; font-size: 11px; color: var(--ink-subtle); text-align: right; border-top: {hasMore ? 'none' : '1px solid var(--line-soft)'};">
      Exibindo {shown.length} de {capped.length}{items.length > maxItems ? ` (limite de ${maxItems})` : ''}
    </div>
  </div>
{/if}

<style>
  .sort-bar {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 8px 20px;
    border-bottom: 1px solid var(--line-soft);
    background: var(--surface-2);
  }
  .sort-bar__rank {
    width: 28px;
    flex-shrink: 0;
  }
  .sort-bar__body {
    flex: 1;
    min-width: 0;
  }
  .sort-bar__metrics {
    display: flex;
    gap: 20px;
    flex-shrink: 0;
  }
  .sort-pill {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: flex-end;
    width: 90px;
    flex-shrink: 0;
    font-size: 11px;
    font-weight: 600;
    color: var(--ink-muted);
    background: transparent;
    border: none;
    cursor: pointer;
    transition: color 0.12s ease;
    white-space: nowrap;
    text-transform: uppercase;
    padding: 0;
  }
  .sort-pill:hover {
    color: var(--ink);
  }
  .sort-pill.active {
    color: var(--ink);
  }
  .sort-indicator {
    position: absolute;
    left: 100%;
    margin-left: 4px;
    display: flex;
    align-items: center;
    color: var(--ink);
  }
  .sort-bar__trend {
    width: 28px;
    flex-shrink: 0;
  }
</style>
