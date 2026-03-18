import type {
  PrizeOddsRequest,
  PrizeOddsResponse,
  StartOddsRequest,
  StartOddsResponse,
  DrawSupporterOddsRequest,
  DrawSupporterOddsResponse,
  TournamentResponse,
  MatchupStatsRequest,
  MatchupStatsResponse,
  ErrorResponse,
} from '../types/api';

const API_BASE = '/api';

export async function calculatePrizeOdds(
  request: PrizeOddsRequest
): Promise<PrizeOddsResponse> {
  const response = await fetch(`${API_BASE}/prize-odds`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    // Try to parse as JSON, but handle non-JSON error responses (e.g., 502 from gateway)
    let errorMessage: string;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        const error: ErrorResponse = await response.json();
        errorMessage = error.error || `HTTP error! status: ${response.status}`;
      } catch {
        errorMessage = `HTTP error! status: ${response.status}`;
      }
    } else {
      // Non-JSON response (e.g., HTML error page from gateway)
      const text = await response.text();
      errorMessage = text || `HTTP error! status: ${response.status}`;
      // If it's a gateway error, provide a more helpful message
      if (response.status === 502 || response.status === 503) {
        errorMessage = `Service unavailable (${response.status}). The API service may be down or not ready.`;
      }
    }
    throw new Error(errorMessage);
  }

  return response.json();
}

export async function calculateStartOdds(
  request: StartOddsRequest
): Promise<StartOddsResponse> {
  const response = await fetch(`${API_BASE}/start-odds`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    // Try to parse as JSON, but handle non-JSON error responses (e.g., 502 from gateway)
    let errorMessage: string;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        const error: ErrorResponse = await response.json();
        errorMessage = error.error || `HTTP error! status: ${response.status}`;
      } catch {
        errorMessage = `HTTP error! status: ${response.status}`;
      }
    } else {
      // Non-JSON response (e.g., HTML error page from gateway)
      const text = await response.text();
      errorMessage = text || `HTTP error! status: ${response.status}`;
      // If it's a gateway error, provide a more helpful message
      if (response.status === 502 || response.status === 503) {
        errorMessage = `Service unavailable (${response.status}). The API service may be down or not ready.`;
      }
    }
    throw new Error(errorMessage);
  }

  return response.json();
}

export async function calculateDrawSupporterOdds(
  request: DrawSupporterOddsRequest
): Promise<DrawSupporterOddsResponse> {
  const response = await fetch(`${API_BASE}/draw-supporter-odds`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    // Try to parse as JSON, but handle non-JSON error responses (e.g., 502 from gateway)
    let errorMessage: string;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        const error: ErrorResponse = await response.json();
        errorMessage = error.error || `HTTP error! status: ${response.status}`;
      } catch {
        errorMessage = `HTTP error! status: ${response.status}`;
      }
    } else {
      // Non-JSON response (e.g., HTML error page from gateway)
      const text = await response.text();
      errorMessage = text || `HTTP error! status: ${response.status}`;
      // If it's a gateway error, provide a more helpful message
      if (response.status === 502 || response.status === 503) {
        errorMessage = `Service unavailable (${response.status}). The API service may be down or not ready.`;
      }
    }
    throw new Error(errorMessage);
  }

  return response.json();
}

async function handleErrorResponse(response: Response): Promise<never> {
  let errorMessage: string;
  const contentType = response.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    try {
      const error: ErrorResponse = await response.json();
      errorMessage = error.error || `HTTP error! status: ${response.status}`;
    } catch {
      errorMessage = `HTTP error! status: ${response.status}`;
    }
  } else {
    const text = await response.text();
    errorMessage = text || `HTTP error! status: ${response.status}`;
    if (response.status === 502 || response.status === 503) {
      errorMessage = `Service unavailable (${response.status}). The API service may be down or not ready.`;
    }
  }
  throw new Error(errorMessage);
}

export async function fetchTournaments(): Promise<TournamentResponse[]> {
  const response = await fetch(`${API_BASE}/tournaments`);

  if (!response.ok) {
    await handleErrorResponse(response);
  }

  return response.json();
}

export async function fetchArchetypes(): Promise<string[]> {
  const response = await fetch(`${API_BASE}/archetypes`);

  if (!response.ok) {
    await handleErrorResponse(response);
  }

  return response.json();
}

export async function getMatchupStats(
  request: MatchupStatsRequest
): Promise<MatchupStatsResponse> {
  const response = await fetch(`${API_BASE}/matchup-stats`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    await handleErrorResponse(response);
  }

  return response.json();
}

