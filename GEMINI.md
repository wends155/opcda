# 🚀 Project Workflow: opcda Auditor

## 📂 Documentation Ecosystem
*   **`GEMINI.md`**: **Operational Source of Truth**. Rules, workflows, and agent behaviors.
*   **`architecture.md`**: **Technical Source of Truth**. The immutable design spec.
*   **`context.md`**: **Contextual Source of Truth**. The project's memory bank.
*   **`task.md`**: **Execution Tracker**. Strictly for tracking approved implementation steps. **Mandatory:** The first item in any execution list must be `[ ] Create Git Checkpoint`.

> 🛑 **Restricted Access**: Only **High-Reasoning Models** (Gemini 3 Pro / Claude Opus) are authorized to edit `GEMINI.md`, `architecture.md`, and `context.md`.
> *   **The Builder** (Flash/Lower models) is **Read-Only** for these files and must strictly follow them.

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
    * **Compliance**: Must strictly implement plans AND adhere to `architecture.md`. If a plan contradicts the architecture, STOP and query the Architect.
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

## 📝 Documentation Standards
**Rule:** Code must be self-documenting and strictly commented.
1.  **Universal Coverage**: EVERY symbol (Exported AND Unexported) must have a `godoc` comment.
    *   **Public**: `func Connect(...)` -> `// Connect establishes...`
    *   **Private**: `type serverProvider interface` -> `// serverProvider defines the internal contract...`
2.  **Format**: Comments MUST start with the symbol name.
3.  **Context**: 
    *   Explain *what* it does.
    *   Explain *why* it exists (if non-obvious).
    *   Provide examples for complex logic.
4.  **Verification**: Auditors MUST use `mcp_godoc_get_doc` (not `grep`) to verify coverage and rendering.

## 🛡️ Data Safety Protocol
**Rule:** Prevention of accidental data loss is paramount.
1.  **Deletion Restriction**: NEVER programmatically delete source files (`.go`, `.md`, etc.) in the project root or source directories.
    *   **Allowed**: Deleting specific files in `./logs/`, `./temp/`, or the artifact directory.
    *   **Prohibited**: `rm *`, `rm ./*`, or broad wildcard deletions in the root.
2.  **Checkout vs Delete**: If a file needs to be reverted, use `git checkout <file>` instead of deleting it.
3.  **Artifact Isolation**: All temporary artifacts (plans, reports, test files) must be kept in the artifact directory or a dedicated `./temp` folder. Do not mix them with source code.
4.  **Mandatory Checkpoints**: Before applying any changes (Phase 2: Act), the Agent MUST create a git checkpoint using `scripts/gcom "wip: checkpoint before <task>"`. This ensures an easy revert path (`git reset --hard HEAD~1`) if destructive errors occur.
