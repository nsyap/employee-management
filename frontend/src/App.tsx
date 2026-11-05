import { useCallback, useEffect, useState } from "react";
import "./App.css";

type Employee = {
  id: number;
  name: string;
  email: string;
  department: string;
};

const API = "http://localhost:8080/api/employees";

function App() {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [department, setDepartment] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const loadEmployees = useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch(API);
      if (!response.ok) throw new Error("Could not load employees");

      const data: Employee[] = await response.json();
      setEmployees(data);
      setError("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Request failed");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadEmployees();
  }, [loadEmployees]);

  async function addEmployee(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");

    try {
      const response = await fetch(API, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, email, department }),
      });

      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error ?? "Could not add employee");
      }

      setName("");
      setEmail("");
      setDepartment("");
      await loadEmployees();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Request failed");
    } finally {
      setBusy(false);
    }
  }

  async function deleteEmployee(id: number) {
    if (!window.confirm("Delete this employee?")) return;

    setError("");
    try {
      const response = await fetch(`${API}/${id}`, {
        method: "DELETE",
      });

      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error ?? "Could not delete employee");
      }

      await loadEmployees();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Request failed");
    }
  }

  return (
    <main className="container">
      <header>
        <p className="eyebrow">HUMAN RESOURCES</p>
        <h1>Employee Management</h1>
        <p className="subtitle">
          Manage your employee directory in one place.
        </p>
      </header>

      <section className="panel">
        <h2>Add employee</h2>

        <form onSubmit={addEmployee}>
          <label>
            Full name
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Alice Tan"
              required
            />
          </label>

          <label>
            Email
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="alice@example.com"
              required
            />
          </label>

          <label>
            Department
            <input
              value={department}
              onChange={(e) => setDepartment(e.target.value)}
              placeholder="e.g. IT"
              required
            />
          </label>

          <button type="submit" disabled={busy}>
            {busy ? "Saving..." : "Add employee"}
          </button>
        </form>
      </section>

      <section className="panel">
        <div className="list-header">
          <h2>Employee directory</h2>
          <span className="count">{employees.length} employees</span>
        </div>

        {error && <p className="error" role="alert">{error}</p>}

        {loading ? (
          <p>Loading employees...</p>
        ) : employees.length === 0 ? (
          <p>No employees found. Add your first employee above.</p>
        ) : (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Email</th>
                  <th>Department</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                {employees.map((employee) => (
                  <tr key={employee.id}>
                    <td>{employee.name}</td>
                    <td>{employee.email}</td>
                    <td>
                      <span className="department">
                        {employee.department}
                      </span>
                    </td>
                    <td>
                      <button
                        className="delete"
                        onClick={() => void deleteEmployee(employee.id)}
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </main>
  );
}
export default App;
