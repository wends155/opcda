# 🗺️ Project Context: opcda

> **AI Instructions:** This file is the Source of Truth. Update this file during the **Phase 4: Summarize** stage of the TARS workflow.

---

## 🏗️ System Overview
* **Goal:** Standalone Go-based OPC DA client library (migrated to `wends155/opcda`).
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

## Verification

### 🛠️ Recent Changes (Last 3 Cycles)
1.  **2026-02-12 (Migration):** Successfully migrated library to `github.com/wends155/opcda`. Renamed module project-wide and updated all imports.
2.  **2026-02-12 (CI Infrastructure):** Mirrored simulation assets to simulation-assets and updated `test.yaml` for CI.
3.  **2026-02-12 (Documentation/COM):** Improved `com` package documentation to adhere to `go doc` standards. Added runnable examples in `example_test.go` and created a comprehensive source map in `com_source_map.md`.
4.  **2026-02-12 (Nil-Safety & Error Handling):** Implemented comprehensive nil-safety across the `opcda` package. Refactored `VARIANT.Value()` to return `(interface{}, error)` and replaced all internal `panic` calls with proper error propagation. Added a regression test suite in `opcda_error_test.go`.
5.  **2026-02-12 (Unsafe Audit):** Completed a comprehensive security audit of `unsafe` usage in the `com` package. Verified Vtble orders, struct alignments, and memory handling. Refactored syscall patterns to satisfy `go vet` where possible and ensure strict pointer safety. Audited remaining warnings as low-risk COM interop patterns.

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
