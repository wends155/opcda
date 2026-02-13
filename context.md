# 🗺️ Project Context: opcda

> **AI Instructions:** This file is the Source of Truth. Update this file during the **Phase 4: Summarize** stage of the TARS workflow.
> **Update Strategy**:
> *   **Current State**: Summarize *what* was built (Features, APIs).
> *   **Decision Log**: Record *why* choices were made (Constraints, Alternatives).
> *   **Lessons Learned**: Record *how* to avoid past mistakes (Tooling quirks, Process improvements).
> *   **Goal**: Enable "Recursive Summarization" to keep the context window lean without losing wisdom.

---

## 🏗️ System Overview
* **Goal:** Standalone Native Go-based OPC DA client library. This is the overarching objective of the project.
* **Core Stack:** Go 1.20+, `golang.org/x/sys/windows` (COM/OLE).
* **Architecture Pattern:** Go wrapper around OPC DA Automation interfaces, providing an idiomatic Go API for industrial data access. See [architecture.md](file:///c:/Users/WSALIGAN/code/opcda/architecture.md) for deep technical details and [com_source_map.md](file:///c:/Users/WSALIGAN/code/opcda/com/com_source_map.md) for an overview of the low-level communication package.

---

## 💻 Environment & Constraints
* **Host OS:** Windows (Non-Admin)
* **Shell Environment:** PowerShell / BusyBox (via Scoop)
* **Toolchain:** Go, `golang.org/x/sys/windows`.
* **Strict Rules:**
    1. No `sudo`/Admin commands.
    2. Scripts must be portable where possible (using standard Go).
    3. **Shell Syntax**: Use `;` for command chaining in PowerShell; use `&&` ONLY in BusyBox/sh scripts.

---

## 📍 Current State (Recursive Summary)
*This section is updated by the Architect after every successful implementation.*

### com/com.go & com/const.go

- **Change**: Documented `COAUTHIDENTITY` struct and added authentication identity constants (`SEC_WINNT_AUTH_IDENTITY_ANSI`, `SEC_WINNT_AUTH_IDENTITY_UNICODE`).
- **Impact**: Provides clearer guidance for developers implementing custom authentication for COM objects.
- **Reference**: [PR #XX] Documenting COAUTHIDENTITY

### Global Dependency Injection & Mocking (2026-02-12)

- **Change**: Refactored core components (`OPCServer`, `OPCGroup`, `OPCItems`, `OPCItem`, `OPCBrowser`) to use internal provider interfaces (`serverProvider`, `groupProvider`, `itemMgtProvider`). 
- **Impact**: Decoupled high-level Go logic from physical COM implementations. Enabled pure unit testing with mocks in `*_test.go`.
- **Test Isolation**: Implemented `//go:build integration` tags to separate server-dependent tests from unit tests. 
- **Automation**: Introduced a `Makefile` with `test` and `integration` targets for standardized test execution.
- **Registry Fix**: Resolved "Access is denied" when connecting to local servers by using local `registry.CLASSES_ROOT` instead of `OpenRemoteKey("localhost", ...)`.
- **Benefit**: CI/CD can now verify core logic without requiring initialized Windows COM or specialized OPC simulators.

### OPCBrowser Refactor & Diagramming Rule (2026-02-13)

- **Change**: Refactored `opcbrowser.go` to use `browserProvider` pattern and `comBrowserProvider` struct, standardizing the `Release()` signature.
- **Impact**: Aligned `opcbrowser` with the project's dependency injection pattern, improving mockability and interface uniformity.
- **Rules**: Updated `GEMINI.md` to require architecture diagrams in audit/implementation reports.
- **Rules**: Updated `GEMINI.md` to require architecture diagrams in audit/implementation reports.
- **Verification**: `go test` passes for unit tests.

### Log Directory Standardization (2026-02-13)

- **Change**: Enforced all temporary logs to be written to `./logs/`.
- **Tooling**: Updated `Makefile` to create `./logs/` and redirect test output. Added `logs/` to `.gitignore`.
- **Purpose**: Clean up workspace and centralize artifact management.

### Documentation Standards (2026-02-13)

- **Change**: Updated `GEMINI.md` to mandate `godoc` comments for **ALL** symbols (exported and unexported).
- **Rule**: Comments must start with the symbol name and explain *what* and *why*.
- **Verification**: Enforced usage of `mcp_godoc_get_doc` for verifying documentation coverage.
- **Rationale**: Ensures code is self-documenting and accessible via standard Go tools and AI agents.

### Git Tooling Migration (2026-02-13)

- **Change**: Replaced `git-mcp-server` with manual scripts (`scripts/gcom`, `scripts/gsync`) due to path traversal issues on Windows.
- **Impact**: Improved reliability of version control operations.
- **Workflow**: Documented new script usage in `GEMINI.md`.

### Documentation Sync (2026-02-13)

- **Impact**: Documentation now accurately reflects the decoupled architecture and dependency injection used for testing.

### Documentation Audit & Grammar Enforcement (2026-02-13)

- **Change**: Audited `opcserver.go` and `GEMINI.md` for documentation compliance.
- **Standards**: Updated `GEMINI.md` to enforce strict punctuation (full sentences) and specific phrasing ("SymbolName verb..." pattern).
- **Fixes**: Added missing comments to `OPCServer`, `ServerInfo`, and internal helper functions in `opcserver.go`.
- **Verification**: Validated zero missing comments using `grep` (PowerShell `Select-String`) as a fallback for failing `godoc` environment.

### Lessons Learned (2026-02-13)

- **Git Workflow**: `git-mcp-server` is unreliable on Windows/Non-Admin due to path traversal issues. Use custom scripts (`scripts/gcom`, `scripts/gsync`) for consistent, error-free version control.
- **Automation Safety**: Always wait for user instruction before pushing commits (Phase 3). This avoids premature synchronization of unverified changes.
- **Log Management**: Centralize all temporary outputs in `./logs/`. Use `make clean` to maintain a tidy workspace.
- **Documentation**: A comprehensive API Reference in `architecture.md` is critical for developer onboarding and AI context limits.
- **Artifacts**: Maintain `task.md`, `implementation_plan.md`, and `walkthrough.md` as living documents to track progress and decisions.
- **Data Safety (2026-02-13)**: Never use `rm` on the project root to clean up artifacts. Always target specific files or directories (e.g., `rm ./logs/*.log`). If unsure, do not delete.
- **Process (2026-02-13)**: Mandated git checkpoints before every execution phase to mitigate risk of accidental deletion or regression.
- **Godoc Usage (2026-02-13)**: Confirmed `mcp_godoc_get_doc` fails with "module verification" errors when using absolute paths for local packages. **Action**: Always use `path="."` combined with `working_dir="<project_root>"` for local queries.
- **Process Breach (2026-02-13)**: Skipping the "Think" phase leads to unapproved changes and context drift. **Action**: Implemented "Planning Gate" protocol in `GEMINI.md` to lock the Agent into Planning Mode until user approval.
- **File Editing (2026-02-13)**: `replace_file_content` failed repeatedly on `GEMINI.md` due to subtle spacing/newline mismatches. **Action**: When diffs fail twice, switch to `write_to_file` for a complete overwrite to ensure data integrity.

### Thread Safety implementation (2026-02-13)

- **Change**: Implemented partial thread-safety for `OPCItem`, `OPCGroup`, and `OPCGroups`.
- **Details**:
    - **OPCItem**: Added `sync.RWMutex` to protect internal state (`value`, `quality`, `timestamp`, etc.).
    - **OPCGroup**: Used `callbackLock` to protect event listener slices during registration and firing (using copy-on-write pattern for firing).
    - **OPCGroups**: Added `Lock()` to `Release` method to prevent race conditions during iteration.
- **Verification**: Validated with `race_test.go` (removed after verification) covering concurrent item access and group event registration.
- **Documentation**: Updated `architecture.md` with Concurrency Model details.

### 🧩 Active Components & APIs
* `opcda`: Core Go package.
    * `OPCServer`: Main struct for connecting to OPC servers.
* `com`: Low-level COM wrapper package. Detailed in [com_source_map.md](file:///c:/Users/WSALIGAN/code/opcda/com/com_source_map.md). Updated with standardized doc comments and pointer safety patterns.

---

## 📜 Decision Log (The "Why")
*Records why specific paths were taken to prevent circular reasoning in future "Think" phases.*

* **Transition to Go:** The project is a Go implementation of the OPC DA client, diverging from the legacy Python-based OpenOPC model.
* **Documentation Standard:** Adopted a consistent doc comment pattern for COM interfaces to improve readability and internal API discoverability via `go doc`.
* **Pointer Safety:** Standardized `unsafe.Pointer` conversions around syscalls to use direct `syscall.Syscall` and immediate casting, ensuring compatibility with Go's static analysis tools and preventing pointer tracking failures.
* **Error Handling Strategy:** Moved from `panic`-driven error handling to explicit `error` returns. `VARIANT.Value()` signature was updated to `(interface{}, error)` to allow graceful handling of date and array conversion failures, preventing runtime crashes in production environments.
* **Defensive API (Nil-Safety):** Implemented defensive nil-receiver checks across all public `opcda` methods to ensure that calls on zero-initialized or failed connection objects return a structured error instead of a segmentation fault.
* **Structural DI via Implicit Interfaces**: Successfully refactored the entire library to support DI without modifying the `com` package. This was achieved by defining internal interfaces that wrap COM methods, allowing for physical COM interaction in production and mocked behaviors in tests.
* **Universal Test Isolation**: Established a clear separation between Unit tests (mocked, fast, always passing) and Integration tests (requires real OPC server). Integration tests are now in separate files and can be targeted specifically, while `go test ./...` remains a reliable indicator of code health.
* **Defensive Registry Access**: Implemented logic to detect "localhost" in node names and use the local registry API directly, bypassing permissions issues inherent in remote registry calls on local systems.

---

## 🚧 Technical Debt & Pending Logic
* **Next Steps:** Investigate performance benchmarks to verify the impact of the high-level `opcda` logic on COM stability under heavy load.

---

## 🧪 Tooling & Scripts
* **code-index (MCP)**: ✅ **Verified**. Configured for `c:\Users\WSALIGAN\code\opcda`. Advanced search and deep indexing functional.
* **go-analyzer (MCP)**: ✅ **Verified**. Successfully detecting workspace via `go.work`. Diagnostics and symbol search functional.
* **git-mcp (MCP)**: Automated version control.
* **context7 (MCP)**: Documentation queries for OPC DA/Automation specs.
* **Go Toolchain**: `go test`, `go fmt`, `go vet`.
* **Makefile**: Root-level `Makefile` for streamlined testing (`make test`, `make integration`).
* **godoc (MCP)**: Preferred tool for internal API and architectural exploration. Use `mcp_godoc_get_doc` for concise package/symbol summaries.

---

## 📊 MCP Tool Usability Assessment (2026-02-12)

| Tool Category | Usability Findings |
| :--- | :--- |
| **Architecture Exploration** | `go-analyzer` and `mcp_godoc_get_doc` are most effective for mapping COM Vtble layouts and resolving cross-package dependencies. |
| **Semantic Search** | MCP tools provide a critical "semantic bridge" for finding unexported symbols or complex dependency chains that standard `grep` misses. |
| **Search Reliability** | Standard `grep`/`ls` remain faster for high-frequency "sanity checks" due to zero latency. |
| **Platform Constraints** | `code-index` may encounter friction on Windows (BusyBox) if internal search flags (e.g., `--exclude-dir`) are unsupported by the host grep. |
| **Overall Verdict** | MCP tools are "surgical" and best used for deep architectural work, while terminal tools are "broad" and best for quick navigation. |
