export type Skill = { id: number; code: string; name: string };
export type Profile = {
  id: number;
  firstName: string;
  lastName: string;
  avatarUrl: string;
  mobilePhone: string;
  contactEmail: string;
  address: string;
  skills: Skill[];
};
export type ChangeLog = {
  id: number;
  changeType: string;
  fieldName: string;
  oldValue: string;
  newValue: string;
  createdAt: string;
};
export type User = { id: number; email: string; role: string };
export type Employee = { id: number; email: string; role: string; isActive: boolean; profile: Profile | null };
export type CreateEmployeePayload = {
  email: string;
  password: string;
  role: string;
  firstName: string;
  lastName: string;
  mobilePhone: string;
  contactEmail: string;
  address: string;
  skillIds: number[];
};

const API_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";
const API_VERSION = "/api/v1";

type ApiSuccess<T> = { success: true; data: T };
type ApiFailure = { success: false; error: { code: string; message: string }; requestId: string };

export class ApiError extends Error {
  requestId: string;
  code: string;

  constructor(message: string, code: string, requestId: string) {
    super(message);
    this.code = code;
    this.requestId = requestId;
  }
}

async function request<T>(path: string, token?: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (!(init.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  const res = await fetch(`${API_URL}${path}`, { ...init, headers });
  const payload = (await res.json()) as ApiSuccess<T> | ApiFailure;
  if (!res.ok || !payload.success) {
    const failed = payload as ApiFailure;
    throw new ApiError(failed.error?.message ?? "Request failed", failed.error?.code ?? "API_ERROR", failed.requestId ?? "");
  }
  return payload.data;
}

export const api = {
  login: (email: string, password: string) =>
    request<{ token: string; expiresAt: string; user: User }>(`${API_VERSION}/auth/login`, undefined, {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  me: (token: string) => request<Profile>(`${API_VERSION}/profile`, token),
  skills: () => request<Skill[]>(`${API_VERSION}/master-data/skills`),
  updateContact: (token: string, data: Pick<Profile, "mobilePhone" | "contactEmail" | "address">) =>
    request<Profile>(`${API_VERSION}/profile/contact`, token, { method: "PATCH", body: JSON.stringify(data) }),
  updateSkills: (token: string, skillIds: number[]) =>
    request<Profile>(`${API_VERSION}/profile/skills`, token, { method: "PATCH", body: JSON.stringify({ skillIds }) }),
  uploadAvatar: (token: string, file: File) => {
    const form = new FormData();
    form.append("avatar", file);
    return request<Profile>(`${API_VERSION}/profile/avatar`, token, { method: "POST", body: form });
  },
  history: (token: string) => request<ChangeLog[]>(`${API_VERSION}/profile/change-logs?days=7`, token),
  resetDemo: (token: string) => request<Profile>(`${API_VERSION}/demo/profile/reset`, token, { method: "POST" }),
  employees: (token: string) => request<Employee[]>(`${API_VERSION}/admin/employees`, token),
  createEmployee: (token: string, data: CreateEmployeePayload) =>
    request<Employee>(`${API_VERSION}/admin/employees`, token, { method: "POST", body: JSON.stringify(data) }),
  setEmployeeActive: (token: string, id: number, isActive: boolean) =>
    request<Employee>(`${API_VERSION}/admin/employees/${id}/status`, token, { method: "PATCH", body: JSON.stringify({ isActive }) }),
  assetUrl: (path: string) => (path ? `${API_URL}${path}` : ""),
};
