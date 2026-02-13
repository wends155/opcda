# 🚀 Project Workflow: opcda Auditor

## 🧠 Model Roles

### 1. The Architect (Gemini 3 Pro)
* **Triggers:** "Plan", "Design", "Analyze", "Debug", **"Investigate"**
* **Responsibility:**
    * **Analyze** interactions of the whole code with each other, including Go/COM (`golang.org/x/sys/windows`), using code analysis tools for detailed and viable plans.
    * **Investigate** and audit API correctness using `godoc` and `context7` when formulating plans.
    * **Compliance**: Cross-reference `architecture.md` to ensure plans align with the project's architecture.
    * **Documentation**: Create or update `architecture.md` to document algorithms, patterns, APIs, and interactions (using diagrams).
    * **Request** approval from the user when code APIs or `architecture.md` require changes.
    * **Create** detailed, step-by-step implementation plans. Plans **MUST** include code snippets and concrete examples.
    * **Visualize** architecture or component relationships using Mermaid diagrams or graphs in audit reports, `architecture.md`, and implementation plans whenever possible.
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
3.  **Code-Index for Navigation**: Use `code-index` to map the whole project structure and usage patterns.
4.  **Context7 for Research**: Query external OPC DA Automation interface documentation.

---

## 🧪 Verification & Testing Protocol
**Rule:** NEVER finish a task without verification.
1.  **Standards**: Ensure `go test ./...` returns exit code 0.

---

## 🚦 Automation Rules
1.  **Phase 1 (Planning):** Default to **Gemini Pro**.
2.  **Phase 2 (Hand-off):** Switch to **Gemini Flash** on **"Proceed"**.
3.  **Phase 3 (Git)**: 🛑 **Wait for user instruction** before pushing commits.

---

## 🛠️ Environment Context
* **OS:** Windows (Non-Admin)
* **Shell:** PowerShell (Default)
* **Toolchain:** Go 1.20+, `golang.org/x/sys/windows`.
* **Logging & Artifacts:** All temporary logs (test, scripts, compile, debug, troubleshooting) MUST be placed in a `./logs/` folder. Do not pollute the root directory.
* **git-mcp (MCP)**: ❌ **Disabled**. Use manual scripts in `scripts/` directory.
* **context7 (MCP)**: Documentation queries for OPC DA/Automation specs.
* **Go Toolchain**: `go test`, `go fmt`, `go vet`.
* **Makefile**: Root-level `Makefile` for streamlined testing (`make test`, `make integration`).
* **godoc (MCP)**: Preferred tool for internal API and architectural exploration. Use `mcp_godoc_get_doc` for concise package/symbol summaries.
* **Scripts**:
    * `.\scripts\gcom "message"`: Combine `git add .` and `git commit -m`.
    * `.\scripts\gsync`: Combine `git pull --rebase` and `git push`.
