# 🗺️ Project Context: opcda

> **AI Instructions:** This file is the Source of Truth. Update this file during the **Phase 4: Summarize** stage of the TARS workflow.

---

## 🏗️ System Overview
* **Goal:** Standalone Go-based OPC DA client library (migrated to `wends155/opcda`).
* **Core Stack:** Go 1.20+, `golang.org/x/sys/windows` (COM/OLE).
* **Architecture Pattern:** Go wrapper around OPC DA Automation interfaces, providing an idiomatic Go API for industrial data access. See [architecture.md](file:///c:/Users/WSALIGAN/code/opcda/architecture.md) for deep technical details.

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

### 🛠️ Recent Changes (Last 3 Cycles)
1.  **2026-02-12 (Migration):** Successfully migrated library to `github.com/wends155/opcda`. Renamed module project-wide and updated all imports.
2.  **2026-02-12 (CI Infrastructure):** Mirrored simulation assets to [simulation-assets](https://github.com/wends155/opcda/releases/tag/simulation-assets) and updated `test.yaml` for independent CI.
3.  **2026-02-12 (Platform Constraints):** Added explicit `//go:build windows` constraints across all core packages.

### 🧩 Active Components & APIs
* `opcda`: Core Go package.
    * `OPCServer`: Main struct for connecting to OPC servers.
* `com`: Low-level COM wrapper package.

---

## 📜 Decision Log (The "Why")
*Records why specific paths were taken to prevent circular reasoning in future "Think" phases.*

* **Transition to Go:** The project is a Go implementation of the OPC DA client, diverging from the legacy Python-based OpenOPC model while maintaining similar high-level concepts.

---

## 🚧 Technical Debt & Pending Logic
* **Next Steps:** Ensure comprehensive test coverage for COM interface wrappers in the `com` package.

---

## 🧪 Tooling & Scripts
* **code-index (MCP)**: In-memory symbol indexer.
* **git-mcp (MCP)**: Automated version control.
* **context7 (MCP)**: Documentation queries for OPC DA/Automation specs.
* **Go Toolchain**: `go test`, `go fmt`, `go vet`.
* **godoc (MCP)**: Preferred tool for internal API and architectural exploration. Use `mcp_godoc_get_doc` for concise package/symbol summaries.
