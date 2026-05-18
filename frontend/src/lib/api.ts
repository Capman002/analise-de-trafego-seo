// HTTP client tipado para comunicação com o backend Go.
// Em produção (embutido no Go), chamadas são relativas (mesmo host).
// Em dev, o Vite proxy redireciona /api/* para localhost:8080.

import type { Client, TrafficResponse } from './types';

const API_BASE = import.meta.env.PUBLIC_API_URL || '';

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
    ...init,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(body.error || `HTTP ${res.status}`);
  }

  return res.json();
}

/** Lista todos os clientes sincronizados da planilha. */
export function fetchClients(): Promise<Client[]> {
  return request<Client[]>('/api/clients');
}

/** Busca dados de tráfego agregados para um cliente. */
export function fetchTraffic(clientId: number, period: number = 28, startDate?: string, endDate?: string): Promise<TrafficResponse> {
  const params = new URLSearchParams();
  params.append('period', period.toString());
  if (startDate) params.append('start_date', startDate);
  if (endDate) params.append('end_date', endDate);
  
  return request<TrafficResponse>(`/api/traffic/${clientId}?${params.toString()}`);
}

/** Força re-sync da planilha de clientes. */
export function syncClients(): Promise<{ synced: number; status: string }> {
  return request('/api/sync/clients', { method: 'POST' });
}
