<script lang="ts">
  import { fade, fly } from 'svelte/transition';
  import type { Snippet } from 'svelte';
  interface Props {
    open: boolean; title: string; subtitle?: string; color?: string;
    onclose: () => void; icon?: Snippet; children: Snippet;
  }
  let { open, title, subtitle, color, onclose, icon, children }: Props = $props();

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onclose();
  }
</script>

{#if open}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="modal-overlay" transition:fade={{ duration: 180 }} onclick={onclose} onkeydown={handleKeydown} role="dialog" aria-modal="true" tabindex="-1">
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="modal-content" transition:fly={{ y: 12, duration: 200 }} onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="document">
      <div class="modal-header">
        <div class="modal-header__left">
          {#if icon}
            <div class="modal-header__icon" style="color: {color}">
              {@render icon()}
            </div>
          {/if}
          <div>
            <h2 class="modal-header__title">{title}</h2>
            {#if subtitle}<p class="modal-header__sub">{subtitle}</p>{/if}
          </div>
        </div>
        <button class="modal-close" onclick={onclose} aria-label="Fechar">×</button>
      </div>
      <div class="modal-body">
        {@render children()}
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay { position: fixed; inset: 0; background: rgba(15,14,10,0.55); backdrop-filter: blur(2px); display: flex; align-items: center; justify-content: center; z-index: 1000; padding: 24px; }
  .modal-content { width: min(1280px, 96vw); max-height: 90vh; background: var(--surface); border: 1px solid var(--line-soft); border-radius: 5px; display: flex; flex-direction: column; overflow: hidden; }
  .modal-header { display: flex; align-items: center; justify-content: space-between; padding: 16px 20px; border-bottom: 1px solid var(--line-soft); gap: 16px; }
  .modal-header__left { display: flex; align-items: center; gap: 12px; min-width: 0; }
  .modal-header__icon { width: 36px; height: 36px; border-radius: 5px; background: var(--surface-2); display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
  .modal-header__title { margin: 0; font-size: 15px; font-weight: 600; color: var(--ink); }
  .modal-header__sub { margin: 2px 0 0; font-size: 12px; color: var(--ink-subtle); }
  .modal-close { width: 32px; height: 32px; border-radius: 5px; border: 1px solid var(--line-soft); background: transparent; color: var(--ink-subtle); cursor: pointer; display: flex; align-items: center; justify-content: center; font-size: 18px; flex-shrink: 0; line-height: 1; }
  .modal-body { flex: 1; overflow-y: auto; }
</style>
