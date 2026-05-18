<script lang="ts">
  import type { Client } from '$lib/types';

  interface Props {
    clients: Client[];
    selected: Client | null;
    onselect: (client: Client) => void;
    loading?: boolean;
  }

  let { clients, selected, onselect, loading = false }: Props = $props();
</script>

<div class="field" style="min-width: 260px;">
  <label for="client-select" class="field-label">Cliente</label>
  <select
    id="client-select"
    class="select"
    value={selected?.id ?? ''}
    onchange={(e) => {
      const id = Number((e.target as HTMLSelectElement).value);
      const client = clients.find(c => c.id === id);
      if (client) onselect(client);
    }}
    disabled={loading}
    style="min-width: 260px;"
  >
    <option value="" disabled>
      {loading ? 'Carregando...' : 'Selecione um cliente'}
    </option>
    {#each clients as client (client.id)}
      <option value={client.id}>
        {client.name}
        {#if client.permissionError}
          — (sem permissão)
        {:else if !client.gscUrl && !client.ga4Id}
          — (sem dados)
        {/if}
      </option>
    {/each}
  </select>
</div>
