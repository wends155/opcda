# 🚀 Project Workflow: opcda Auditor

## 🧠 Model Roles

### 1. The Architect (Gemini 3 Pro)
* **Triggers:** "Plan", "Design", "Analyze", "Debug", **"Investigate"**
* **Responsibility:**
    * **Analyze** Go/COM interactions (`golang.org/x/sys/windows`).
    * **Investigate** OPC connectivity errors using `godoc` and `context7`.
    * **Create** detailed, step-by-step implementation plans. Plans **MUST** include code snippets and concrete examples to provide clear context for the **Builder**.
    * **Define** the verification strategy.

### 2. The Builder (Gemini 3 Flash)
* **Triggers:** "Implement", "Write", "Code", "Generate", **"Proceed"**
* **Responsibility:**
    * **Execute** the Architect's plan exactly.
    * **Write** idiomatic Go code.
    * **Refine** code using `go fmt`, `go vet`, or `golangci-lint`.

---

## 🛠️ Tool-Centric Architecture
**Rule:** Agents interact with the world through tools.
1.  **Prioritize Go Tools**: Use `go test` for verification, `go build` for compilation checks.
2.  **godoc for Exploration**: ALWAYS use `godoc` first to understand package structure and exported symbols.
3.  **Code-Index for Navigation**: Use `code-index` to map COM interface usage patterns.
4.  **Context7 for Research**: Query external OPC DA Automation interface documentation.

---

## 🧪 Verification & Testing Protocol
**Rule:** NEVER finish a task without verification.
1.  **Standards**: Ensure `go test ./...` returns exit code 0.

---

## 🚦 Automation Rules
1.  **Phase 1 (Planning):** Default to **Gemini Pro**.
2.  **Phase 2 (Hand-off):** Switch to **Gemini Flash** on **"Proceed"**.

---

## 🛠️ Environment Context
* **OS:** Windows (Non-Admin)
* **Shell:** PowerShell (Default)
* **Toolchain:** Go 1.20+, `golang.org/x/sys/windows`.
