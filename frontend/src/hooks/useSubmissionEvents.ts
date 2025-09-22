"use client";
import { useEffect, useRef, useState } from 'react';

export interface SubmissionEventPayloadStatus { status: string }
export interface SubmissionEventPayloadJudgeRun { judgeRun: { id: string; status: string; durationMs?: number; createdAt?: string } }
export type SubmissionEventPayload = SubmissionEventPayloadStatus | SubmissionEventPayloadJudgeRun | Record<string, unknown> | undefined;

export interface SubmissionEvent {
  type: string; // e.g. status_update, judge_run_update, completed
  submissionId: string;
  payload?: SubmissionEventPayload;
  ts: number;
}

export interface UseSubmissionEventsOptions {
  submissionId?: string;
  enabled?: boolean;
  url?: string; // override SSE endpoint
  onEvent?: (evt: SubmissionEvent) => void;
  onError?: (err: Event) => void;
  reconnectIntervalMs?: number;
}

export function useSubmissionEvents(opts: UseSubmissionEventsOptions) {
  const { submissionId, enabled = true, url, onEvent, onError, reconnectIntervalMs = 5000 } = opts;
  const [connected, setConnected] = useState(false);
  const [lastEvent, setLastEvent] = useState<SubmissionEvent | null>(null);
  const esRef = useRef<EventSource | null>(null);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (!enabled || !submissionId) return;

    const endpoint = url || `/submissions/${submissionId}/events`; // 假设后端 SSE 路径

    function connect() {
      cleanup();
      const es = new EventSource(endpoint, { withCredentials: true });
      esRef.current = es;
      es.onopen = () => setConnected(true);
      es.onmessage = (e) => {
        try {
          const data: Record<string, unknown> = JSON.parse(e.data);
          const evt: SubmissionEvent = {
            type: (data.type as string) || 'unknown',
            submissionId: ((data.submissionId as string) || submissionId)!,
            payload: data.payload as SubmissionEventPayload,
            ts: Date.now(),
          };
          setLastEvent(evt);
          onEvent?.(evt);
        } catch {
          // ignore parse errors
        }
      };
      es.onerror = (ev) => {
        setConnected(false);
        onError?.(ev);
        scheduleReconnect();
      };
    }

    function scheduleReconnect() {
      if (reconnectTimer.current) return;
      reconnectTimer.current = setTimeout(() => {
        reconnectTimer.current = null;
        connect();
      }, reconnectIntervalMs);
    }

    function cleanup() {
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
        reconnectTimer.current = null;
      }
      if (esRef.current) {
        esRef.current.close();
        esRef.current = null;
      }
    }

    connect();
    return () => cleanup();
  }, [submissionId, enabled, url, onEvent, onError, reconnectIntervalMs]);

  return { connected, lastEvent };
}
