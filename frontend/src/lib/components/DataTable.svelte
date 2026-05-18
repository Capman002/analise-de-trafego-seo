<script lang="ts">
  interface Column { key: string; label: string; }
  interface Props { data: any[]; columns: Column[]; }
  let { data, columns }: Props = $props();

  function formatCell(col: Column, row: any): { html: boolean; value: string } {
    let val = row[col.key];
    if (typeof val === 'number' && Number.isFinite(val)) {
      if (col.key.endsWith('_diff')) {
        const v = val > 0 ? 'up' : val < 0 ? 'down' : 'flat';
        const s = val > 0 ? '+' : '';
        return { html: true, value: `<span class="cell-diff cell-diff--${v}">${s}${val.toLocaleString('pt-BR')}</span>` };
      }
      if (['ctr', 'position', 'avgPosition'].includes(col.key)) return { html: false, value: val.toFixed(2) };
      if (col.key === 'engagementRate') return { html: false, value: `${(val * 100).toFixed(1)}%` };
      if (['revenue', 'itemRevenue', 'totalRevenue'].includes(col.key))
        return { html: false, value: `R$ ${val.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}` };
      return { html: false, value: val.toLocaleString('pt-BR') };
    }
    return { html: false, value: val ?? '' };
  }
  let safeData = $derived(Array.isArray(data) ? data : []);
  let safeColumns = $derived(Array.isArray(columns) ? columns : []);
</script>

<div class="tbl-wrap">
  <div class="tbl-scroll">
    <table class="tbl">
      <thead><tr>{#each safeColumns as col (col.key)}<th>{col.label}</th>{/each}</tr></thead>
      <tbody>
        {#each safeData as row, i (i)}
          <tr>
            {#each safeColumns as col (col.key)}
              {@const cell = formatCell(col, row)}
              {#if cell.html}<td>{@html cell.value}</td>{:else}<td>{cell.value}</td>{/if}
            {/each}
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
