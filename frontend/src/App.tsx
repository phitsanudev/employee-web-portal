import { FormEvent, useEffect, useMemo, useState } from "react";
import {
  Camera,
  Eye,
  EyeOff,
  History,
  LayoutDashboard,
  LogOut,
  Mail,
  MapPin,
  Phone,
  RotateCcw,
  Save,
  ShieldCheck,
  Sparkles,
  UserRound,
  Users,
} from "lucide-react";
import { ApiError, ChangeLog, Employee, Profile, Skill, User, api } from "./api/client";

type Toast = { type: "success" | "error"; message: string; requestId?: string };
type View = "profile" | "history" | "employees";

const DEMO_EMAIL = "demo@employee.dev";
const DEMO_PASSWORD = "password123";

function getStoredUser() {
  const raw = localStorage.getItem("employee_portal_user");
  if (!raw) return null;

  try {
    return JSON.parse(raw) as User;
  } catch {
    localStorage.removeItem("employee_portal_user");
    return null;
  }
}

export function App() {
  const [token, setToken] = useState(() => localStorage.getItem("employee_portal_token") ?? "");
  const [user, setUser] = useState<User | null>(() => getStoredUser());
  const [profile, setProfile] = useState<Profile | null>(null);
  const [skills, setSkills] = useState<Skill[]>([]);
  const [history, setHistory] = useState<ChangeLog[]>([]);
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [toast, setToast] = useState<Toast | null>(null);
  const [loading, setLoading] = useState(false);
  const [activeView, setActiveView] = useState<View>("profile");
  const [employeesResetKey, setEmployeesResetKey] = useState(0);

  useEffect(() => {
    api.skills().then(setSkills).catch(showError);
  }, []);

  useEffect(() => {
    if (token) {
      refresh(token);
    }
  }, [token, user?.role]);

  async function refresh(activeToken = token) {
    setLoading(true);
    try {
      const [nextProfile, nextHistory] = await Promise.all([api.me(activeToken), api.history(activeToken)]);
      setProfile(nextProfile);
      setHistory(nextHistory);
      if (user?.role === "admin") {
        setEmployees(await api.employees(activeToken));
      }
    } catch (err) {
      if (err instanceof ApiError && err.code === "UNAUTHORIZED") {
        logout();
        setToast({ type: "error", message: "Session expired. Please sign in again.", requestId: err.requestId });
        return;
      }
      showError(err);
    } finally {
      setLoading(false);
    }
  }

  function showError(err: unknown) {
    if (err instanceof ApiError) {
      setToast({ type: "error", message: err.message, requestId: err.requestId });
      return;
    }
    setToast({ type: "error", message: "Unexpected error" });
  }

  function handleLogin(nextToken: string, nextUser: User) {
    localStorage.setItem("employee_portal_token", nextToken);
    localStorage.setItem("employee_portal_user", JSON.stringify(nextUser));
    setToken(nextToken);
    setUser(nextUser);
    setToast({ type: "success", message: "Welcome back" });
  }

  function logout() {
    localStorage.removeItem("employee_portal_token");
    localStorage.removeItem("employee_portal_user");
    setToken("");
    setUser(null);
    setProfile(null);
    setHistory([]);
    setEmployees([]);
  }

  async function resetDemo() {
    setLoading(true);
    try {
      const nextProfile = await api.resetDemo(token);
      setProfile(nextProfile);
      await refresh();
      setToast({ type: "success", message: "Demo data reset" });
    } catch (err) {
      showError(err);
    } finally {
      setLoading(false);
    }
  }

  if (!token) {
    return <LoginPage toast={toast} onDismissToast={() => setToast(null)} onLogin={handleLogin} onError={showError} />;
  }

  return (
    <main className="dashboard-layout">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-icon">EP</div>
          <div>
            <strong>EmpPortal</strong>
            <span>Employee Portal</span>
          </div>
        </div>
        <nav className="sidebar-nav">
          <p>Main</p>
          <button className={activeView === "profile" ? "active" : ""} onClick={() => setActiveView("profile")}>
            <LayoutDashboard size={18} /> Profile
          </button>
          <button className={activeView === "history" ? "active" : ""} onClick={() => setActiveView("history")}>
            <History size={18} /> Change History
          </button>
          {user?.role === "admin" && (
            <button
              className={activeView === "employees" ? "active" : ""}
              onClick={() => {
                setActiveView("employees");
                setEmployeesResetKey((value) => value + 1);
              }}
            >
              <Users size={18} /> Employees
            </button>
          )}
          <p>Account</p>
          <button onClick={logout}>
            <LogOut size={18} /> Log out
          </button>
        </nav>
      </aside>

      <section className="main-area">
        <header className="app-topbar">
          <div />
          <div className="topbar-actions">
            <span className="role-badge">{user?.role === "admin" ? "admin" : "employee"}</span>
          </div>
        </header>

        <section className="content-area">
          {toast && <ToastMessage toast={toast} onDismiss={() => setToast(null)} />}

          {activeView === "profile" && (
            <>
              <PageHeading title="My Profile" subtitle="View and update your employee information" />
              <div className="profile-layout">
                <ProfileCard
                  profile={profile}
                  loading={loading}
                  token={token}
                  onReset={resetDemo}
                  onUploaded={(next) => {
                    setProfile(next);
                    refresh();
                    setToast({ type: "success", message: "Avatar updated" });
                  }}
                  onError={showError}
                />
                <EditPanel
                  token={token}
                  profile={profile}
                  skills={skills}
                  onUpdated={(next) => {
                    setProfile(next);
                    refresh();
                    setToast({ type: "success", message: "Profile updated" });
                  }}
                  onError={showError}
                />
              </div>
            </>
          )}

          {activeView === "history" && (
            <>
              <PageHeading title="Change History" subtitle="Profile updates within the last 7 days" />
              <HistoryPanel items={history} />
            </>
          )}

          {activeView === "employees" && user?.role === "admin" && (
            <>
              <PageHeading title="Employees" subtitle="Create and manage employee accounts" />
              <AdminPanel
                token={token}
                skills={skills}
                employees={employees}
                resetKey={employeesResetKey}
                onChanged={async () => {
                  setEmployees(await api.employees(token));
                  setToast({ type: "success", message: "Employee data updated" });
                }}
                onError={showError}
              />
            </>
          )}
        </section>

        <footer>Copyright © 2026 Employee Portal. Built with Go Gin and React.</footer>
      </section>
      {loading && <div className="loading-bar">Syncing profile data...</div>}
    </main>
  );
}

function ToastMessage({ toast, onDismiss }: { toast: Toast; onDismiss: () => void }) {
  return (
    <div className={`toast ${toast.type}`}>
      <span>{toast.message}</span>
      {toast.requestId && <small>Request ID: {toast.requestId}</small>}
      <button onClick={onDismiss}>Dismiss</button>
    </div>
  );
}

function PageHeading({ title, subtitle }: { title: string; subtitle: string }) {
  return (
    <div className="page-heading">
      <div>
        <h1>{title}</h1>
        <p>{subtitle}</p>
      </div>
    </div>
  );
}

function LoginPage({
  toast,
  onDismissToast,
  onLogin,
  onError,
}: {
  toast: Toast | null;
  onDismissToast: () => void;
  onLogin: (token: string, user: User) => void;
  onError: (err: unknown) => void;
}) {
  const [email, setEmail] = useState(DEMO_EMAIL);
  const [password, setPassword] = useState(DEMO_PASSWORD);
  const [showPassword, setShowPassword] = useState(false);
  const [emailError, setEmailError] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit(event: FormEvent) {
    event.preventDefault();
    const normalizedEmail = email.trim();
    if (!isValidEmail(normalizedEmail)) {
      setEmailError("Please enter a valid email address.");
      return;
    }
    setEmailError("");
    setLoading(true);
    try {
      const result = await api.login(normalizedEmail, password);
      onLogin(result.token, result.user);
    } catch (err) {
      onError(err);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="login-screen">
      <section className="login-panel">
        <div className="brand-mark">
          <ShieldCheck size={32} />
        </div>
        <p className="eyebrow">Senior Go Assignment</p>
        <h1>Employee Web Portal</h1>
        {toast && <ToastMessage toast={toast} onDismiss={onDismissToast} />}
        <form onSubmit={submit}>
          <label>
            Email
            <input
              value={email}
              aria-invalid={Boolean(emailError)}
              onChange={(e) => {
                setEmail(e.target.value);
                if (emailError) setEmailError("");
              }}
            />
            {emailError && <span className="field-error">{emailError}</span>}
          </label>
          <label>
            Password
            <span className="password-field">
              <input type={showPassword ? "text" : "password"} value={password} onChange={(e) => setPassword(e.target.value)} />
              <button type="button" onClick={() => setShowPassword((value) => !value)} title={showPassword ? "Hide password" : "Show password"}>
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </span>
          </label>
          <button className="primary" disabled={loading}>
            {loading ? "Signing in..." : "Sign in"}
          </button>
        </form>
      </section>
    </main>
  );
}

function ProfileCard({
  profile,
  loading,
  token,
  onReset,
  onUploaded,
  onError,
}: {
  profile: Profile | null;
  loading: boolean;
  token: string;
  onReset: () => void;
  onUploaded: (profile: Profile) => void;
  onError: (err: unknown) => void;
}) {
  const initials = profile ? `${profile.firstName[0]}${profile.lastName[0]}` : "";

  async function uploadAvatar(file?: File) {
    if (!file) return;
    try {
      onUploaded(await api.uploadAvatar(token, file));
    } catch (err) {
      onError(err);
    }
  }

  return (
    <article className="content-card profile-card">
      <div className="avatar-wrap">
        <div className="avatar">
          {profile?.avatarUrl ? <img src={api.assetUrl(profile.avatarUrl)} alt="Profile" /> : <span>{initials || <UserRound />}</span>}
        </div>
        <label className="avatar-upload" title="Upload avatar">
          <Camera size={18} />
          <input type="file" accept="image/png,image/jpeg,image/webp" onChange={(e) => uploadAvatar(e.target.files?.[0])} />
        </label>
      </div>
      <p className="eyebrow">Read-only identity</p>
      <h2>{profile ? `${profile.firstName} ${profile.lastName}` : loading ? "Loading..." : "Employee"}</h2>
      <div className="contact-list">
        <span><Phone size={16} /> {profile?.mobilePhone || "-"}</span>
        <span><Mail size={16} /> {profile?.contactEmail || "-"}</span>
        <span><MapPin size={16} /> {profile?.address || "-"}</span>
      </div>
      <div className="skill-row">
        {profile?.skills.map((skill) => <span key={skill.id}>{skill.name}</span>)}
      </div>
      <button className="secondary reset-button" onClick={onReset}>
        <RotateCcw size={16} /> Reset demo data
      </button>
    </article>
  );
}

function EditPanel({
  token,
  profile,
  skills,
  onUpdated,
  onError,
}: {
  token: string;
  profile: Profile | null;
  skills: Skill[];
  onUpdated: (profile: Profile) => void;
  onError: (err: unknown) => void;
}) {
  const [saving, setSaving] = useState(false);
  const [mobilePhone, setMobilePhone] = useState("");
  const [contactEmail, setContactEmail] = useState("");
  const [address, setAddress] = useState("");
  const [selected, setSelected] = useState<number[]>([]);

  useEffect(() => {
    if (!profile) return;
    setMobilePhone(profile.mobilePhone);
    setContactEmail(profile.contactEmail);
    setAddress(profile.address);
    setSelected(profile.skills.map((skill) => skill.id));
  }, [profile]);

  const selectedSet = useMemo(() => new Set(selected), [selected]);

  async function saveContact(event: FormEvent) {
    event.preventDefault();
    setSaving(true);
    try {
      onUpdated(await api.updateContact(token, { mobilePhone, contactEmail, address }));
    } catch (err) {
      onError(err);
    } finally {
      setSaving(false);
    }
  }

  async function saveSkills() {
    setSaving(true);
    try {
      onUpdated(await api.updateSkills(token, selected));
    } catch (err) {
      onError(err);
    } finally {
      setSaving(false);
    }
  }

  return (
    <article className="content-card edit-panel">
      <div className="panel-heading">
        <div>
          <p className="eyebrow">Editable profile</p>
          <h2>Contact and skills</h2>
        </div>
      </div>
      <form className="form-grid" onSubmit={saveContact}>
        <label>
          First name
          <input value={profile?.firstName ?? ""} readOnly />
        </label>
        <label>
          Last name
          <input value={profile?.lastName ?? ""} readOnly />
        </label>
        <label>
          Mobile phone
          <input value={mobilePhone} onChange={(e) => setMobilePhone(e.target.value)} />
        </label>
        <label>
          Contact email
          <input value={contactEmail} onChange={(e) => setContactEmail(e.target.value)} />
        </label>
        <label className="full">
          Address
          <textarea value={address} onChange={(e) => setAddress(e.target.value)} />
        </label>
        <button className="primary" disabled={saving}>
          <Save size={18} /> Save contact
        </button>
      </form>
      <div className="skills-editor">
        {skills.map((skill) => (
          <button
            type="button"
            key={skill.id}
            className={selectedSet.has(skill.id) ? "chip selected" : "chip"}
            onClick={() => setSelected((items) => (items.includes(skill.id) ? items.filter((id) => id !== skill.id) : [...items, skill.id]))}
          >
            <Sparkles size={14} /> {skill.name}
          </button>
        ))}
        <button className="secondary" onClick={saveSkills} disabled={saving}>
          Save skills
        </button>
      </div>
    </article>
  );
}

function AdminPanel({
  token,
  skills,
  employees,
  resetKey,
  onChanged,
  onError,
}: {
  token: string;
  skills: Skill[];
  employees: Employee[];
  resetKey: number;
  onChanged: () => void;
  onError: (err: unknown) => void;
}) {
  const [saving, setSaving] = useState(false);
  const emptyForm = {
    email: "",
    password: "",
    role: "employee",
    firstName: "",
    lastName: "",
    mobilePhone: "",
    contactEmail: "",
    address: "",
    skillIds: [] as number[],
  };
  const [form, setForm] = useState(emptyForm);

  useEffect(() => {
    setForm(emptyForm);
  }, [resetKey]);

  async function create(event: FormEvent) {
    event.preventDefault();
    setSaving(true);
    try {
      await api.createEmployee(token, form);
      setForm(emptyForm);
      onChanged();
    } catch (err) {
      onError(err);
    } finally {
      setSaving(false);
    }
  }

  async function toggle(employee: Employee) {
    setSaving(true);
    try {
      await api.setEmployeeActive(token, employee.id, !employee.isActive);
      onChanged();
    } catch (err) {
      onError(err);
    } finally {
      setSaving(false);
    }
  }

  return (
    <article className="content-card admin-panel">
      <div className="panel-heading">
        <div>
          <p className="eyebrow">Admin</p>
          <h2>Add Employee</h2>
        </div>
        <span className="count-pill">{employees.length} users</span>
      </div>
      <form className="form-grid" onSubmit={create}>
        <label>
          Login email
          <input value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value, contactEmail: e.target.value })} />
        </label>
        <label>
          Password
          <input value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} />
        </label>
        <label>
          First name
          <input value={form.firstName} onChange={(e) => setForm({ ...form, firstName: e.target.value })} />
        </label>
        <label>
          Last name
          <input value={form.lastName} onChange={(e) => setForm({ ...form, lastName: e.target.value })} />
        </label>
        <label>
          Mobile phone
          <input value={form.mobilePhone} onChange={(e) => setForm({ ...form, mobilePhone: e.target.value })} />
        </label>
        <label>
          Role
          <select value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value })}>
            <option value="employee">employee</option>
            <option value="admin">admin</option>
          </select>
        </label>
        <label className="full">
          Address
          <textarea value={form.address} onChange={(e) => setForm({ ...form, address: e.target.value })} />
        </label>
        <div className="full skills-editor">
          {skills.map((skill) => (
            <button
              type="button"
              key={skill.id}
              className={form.skillIds.includes(skill.id) ? "chip selected" : "chip"}
              onClick={() =>
                setForm((current) => ({
                  ...current,
                  skillIds: current.skillIds.includes(skill.id)
                    ? current.skillIds.filter((id) => id !== skill.id)
                    : [...current.skillIds, skill.id],
                }))
              }
            >
              <Sparkles size={14} /> {skill.name}
            </button>
          ))}
        </div>
        <button className="primary" disabled={saving}>
          <UserRound size={18} /> Create employee
        </button>
      </form>

      <div className="employee-table">
        <div className="employee-table-head">
          <span>Name</span>
          <span>Email</span>
          <span>Role</span>
          <span>Status</span>
          <span>Action</span>
        </div>
        {employees.map((employee) => (
          <div className="employee-row" key={employee.id}>
            <strong>{employee.profile ? `${employee.profile.firstName} ${employee.profile.lastName}` : employee.email}</strong>
            <span>{employee.email}</span>
            <span>{employee.role}</span>
            <span className={employee.isActive ? "status-label active" : "status-label"}>{employee.isActive ? "Active" : "Inactive"}</span>
            <button className={employee.isActive ? "status active" : "status"} onClick={() => toggle(employee)} disabled={saving}>
              {employee.isActive ? "Deactivate" : "Activate"}
            </button>
          </div>
        ))}
      </div>
    </article>
  );
}

function HistoryPanel({ items }: { items: ChangeLog[] }) {
  const [filter, setFilter] = useState("all");
  const [search, setSearch] = useState("");
  const filtered = (filter === "all" ? items : items.filter((item) => item.changeType === filter)).filter((item) => {
    const term = search.trim().toLowerCase();
    if (!term) return true;
    return [item.changeType, item.fieldName, item.oldValue, item.newValue].some((value) => value.toLowerCase().includes(term));
  });
  return (
    <article className="content-card history-panel">
      <div className="history-toolbar">
        <div>
          <div className="table-title">
            <span className="table-icon"><History size={18} /></span>
            <h2>Change History Table</h2>
          </div>
          <p>Searchable table for profile update logs.</p>
        </div>
        <input className="history-search" placeholder="Search logs" value={search} onChange={(event) => setSearch(event.target.value)} />
      </div>
      <div className="segmented">
        {["all", "contact", "skills", "avatar"].map((item) => (
          <button key={item} className={filter === item ? "active" : ""} onClick={() => setFilter(item)}>
            {item}
          </button>
        ))}
      </div>
      {filtered.length === 0 ? (
        <p className="empty-state">No profile changes yet.</p>
      ) : (
        <div className="history-table-wrap">
          <table className="history-table">
            <thead>
              <tr>
                <th>Type</th>
                <th>Field</th>
                <th>Before</th>
                <th>After</th>
                <th>Date</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((item) => (
                <tr key={item.id}>
                  <td className="type-cell"><span className={`type-badge ${item.changeType}`}>{item.changeType}</span></td>
                  <td><strong>{formatField(item.fieldName)}</strong></td>
                  <td className="table-value">{item.oldValue || "-"}</td>
                  <td className="table-value">{item.newValue || "-"}</td>
                  <td>{new Date(item.createdAt).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </article>
  );
}

function formatField(value: string) {
  return value.replace(/_/g, " ");
}

function isValidEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(value);
}
