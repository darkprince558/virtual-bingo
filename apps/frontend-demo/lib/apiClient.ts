function normalizeApiBaseUrl(rawUrl: string, shouldAppendApiPrefix: boolean) {
  const trimmed = rawUrl.trim().replace(/\/+$/, '')

  if (!shouldAppendApiPrefix || trimmed.endsWith('/api/v1')) {
    return trimmed
  }

  return `${trimmed}/api/v1`
}

export const API_BASE_URL = normalizeApiBaseUrl(
  process.env.NEXT_PUBLIC_API_URL ||
    process.env.NEXT_PUBLIC_API_BASE_URL ||
    'http://localhost:8080/api/v1',
  !process.env.NEXT_PUBLIC_API_URL && !!process.env.NEXT_PUBLIC_API_BASE_URL
)

export interface ApiClientOptions extends RequestInit {
  // Pass dynamic user info for dev auth simulation
  devUserEmail?: string;
  devUserName?: string;
  devUserRole?: string;
}

export async function apiClient<T>(endpoint: string, options: ApiClientOptions = {}): Promise<T> {
  const { devUserEmail, devUserName, devUserRole, ...customConfig } = options;

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(customConfig.headers as Record<string, string>),
  };

  // Inject dev auth headers if provided (defaults to host auth per dev.go logic if omitted)
  if (devUserEmail) headers['X-Dev-User-Email'] = devUserEmail;
  if (devUserName) headers['X-Dev-User-Name'] = devUserName;
  if (devUserRole) headers['X-Dev-User-Role'] = devUserRole;

  const config: RequestInit = {
    method: customConfig.method || 'GET',
    ...customConfig,
    headers,
  };

  if (customConfig.body) {
    config.body = customConfig.body;
  }

  let response: Response
  try {
    response = await fetch(`${API_BASE_URL}${endpoint}`, config)
  } catch {
    throw new Error(`Unable to reach the Virtual Bingo API at ${API_BASE_URL}. Check NEXT_PUBLIC_API_URL and make sure the Go backend is running.`)
  }

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData?.error?.message || `API Error: ${response.status} ${response.statusText} for ${endpoint}`);
  }

  // The backend wraps responses in a { data: ... } object per api.go writeData()
  const result = await response.json();
  return result.data as T;
}
