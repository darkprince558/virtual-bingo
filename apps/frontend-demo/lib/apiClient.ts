const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

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

  const response = await fetch(`${API_BASE_URL}${endpoint}`, config);

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData?.error?.message || `API Error: ${response.status} ${response.statusText}`);
  }

  // The backend wraps responses in a { data: ... } object per api.go writeData()
  const result = await response.json();
  return result.data as T;
}
