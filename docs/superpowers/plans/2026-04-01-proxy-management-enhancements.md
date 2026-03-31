# Proxy Management Enhancements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add selectable proxy page sizes, region-based filtering, and a safer batch-delete flow to the proxy management page without changing backend APIs or database schema.

**Architecture:** Keep the existing `api.listProxies()` one-shot fetch, and compute region options, filtered rows, page counts, and visible rows inside `frontend/src/pages/Proxies.tsx`. Reuse shared UI primitives already present in the repo, especially `Select` for compact dropdown controls and `useConfirmDialog` for destructive confirmation, while leaving `api.ts`, `admin/handler.go`, and database code unchanged.

**Tech Stack:** React 19, TypeScript, Vite, react-i18next, shared UI components in `frontend/src/components/ui`, existing admin API

---

## File Structure

- Modify: `frontend/src/pages/Proxies.tsx`
  - Add page-size and region filter state
  - Derive `regionOptions`, `filteredProxies`, `totalPages`, and `pagedProxies`
  - Improve batch-delete flow with shared confirmation dialog and in-flight state
  - Update toolbar, empty state, and pagination footer UI
- Modify: `frontend/src/locales/zh.json`
  - Add Chinese labels for region filter, page-size selector, filtered-empty state, and proxy batch-delete confirmation
- Modify: `frontend/src/locales/en.json`
  - Add English labels for the same new UI states

## Testing Strategy

This repo does not currently include a frontend unit-test runner in `frontend/package.json`, so this feature should **not** add Vitest/RTL scaffolding just to satisfy a localized UI enhancement. Verification for this plan is:

- `npm run typecheck` in `frontend/`
- `npm run build` in `frontend/`
- Manual regression on the proxy page:
  - page size changes
  - region filtering
  - batch delete confirmation/loading/selection cleanup
  - existing proxy test / toggle / single delete / pool switch still work

### Task 1: Add Proxy Filter And Page-Size State

**Files:**
- Modify: `F:/自动化/codexproxy/codex2api/frontend/src/pages/Proxies.tsx`
- Modify: `F:/自动化/codexproxy/codex2api/frontend/src/locales/zh.json`
- Modify: `F:/自动化/codexproxy/codex2api/frontend/src/locales/en.json`
- Verify: `F:/自动化/codexproxy/codex2api/frontend/package.json`

- [ ] **Step 1: Add locale entries for the new controls and empty states**

Insert new `proxies` keys in both locale files so the page can render labels without inline hardcoded text.

```json
{
  "regionFilter": "地区筛选",
  "allRegions": "全部地区",
  "untestedRegion": "未测试",
  "pageSize": "每页显示",
  "filteredCount": "筛选后 {{count}} 个代理",
  "noFilteredProxies": "当前筛选条件下没有代理",
  "noFilteredProxiesDesc": "试试切换地区筛选，或先测试代理以获取地区信息。"
}
```

```json
{
  "regionFilter": "Region",
  "allRegions": "All regions",
  "untestedRegion": "Untested",
  "pageSize": "Per page",
  "filteredCount": "{{count}} proxies after filtering",
  "noFilteredProxies": "No proxies match the current filter",
  "noFilteredProxiesDesc": "Try another region filter, or test proxies first to populate location data."
}
```

- [ ] **Step 2: Add page-size and region filter state in `Proxies.tsx`**

Replace the fixed page-size constant with stateful control values and add sentinel constants for filter options.

```tsx
const PAGE_SIZE_OPTIONS = [10, 50, 100, 200] as const
const REGION_ALL = '__all__'
const REGION_UNTESTED = '__untested__'

const [page, setPage] = useState(1)
const [pageSize, setPageSize] = useState<number>(10)
const [regionFilter, setRegionFilter] = useState(REGION_ALL)
```

- [ ] **Step 3: Derive region options, filtered rows, and paged rows from the fetched proxy list**

Compute filter options and the visible dataset in render scope, then keep page resets deterministic when controls change.

```tsx
const regionOptions = [
  { value: REGION_ALL, label: t('proxies.allRegions') },
  { value: REGION_UNTESTED, label: t('proxies.untestedRegion') },
  ...Array.from(new Set(
    proxies
      .map((proxy) => proxy.test_location.trim())
      .filter(Boolean)
  ))
    .sort((a, b) => a.localeCompare(b))
    .map((location) => ({ value: location, label: location })),
]

const filteredProxies = proxies.filter((proxy) => {
  if (regionFilter === REGION_ALL) return true
  if (regionFilter === REGION_UNTESTED) return !proxy.test_location.trim()
  return proxy.test_location === regionFilter
})

const totalPages = Math.max(1, Math.ceil(filteredProxies.length / pageSize))
const pagedProxies = filteredProxies.slice((page - 1) * pageSize, page * pageSize)

useEffect(() => {
  setPage(1)
}, [pageSize, regionFilter])

useEffect(() => {
  if (page > totalPages) setPage(totalPages)
}, [page, totalPages])
```

- [ ] **Step 4: Add the compact filter controls inside the proxy table card**

Reuse the existing shared `Select` component instead of introducing a new dropdown implementation.

```tsx
<div className="flex items-center justify-between gap-3 px-4 py-3 border-b border-border flex-wrap">
  <div className="flex items-center gap-3 flex-wrap">
    <div className="min-w-[180px]">
      <div className="mb-1 text-[11px] font-semibold uppercase tracking-[0.08em] text-muted-foreground">
        {t('proxies.regionFilter')}
      </div>
      <Select
        compact
        value={regionFilter}
        onValueChange={setRegionFilter}
        options={regionOptions}
      />
    </div>
    <div className="min-w-[120px]">
      <div className="mb-1 text-[11px] font-semibold uppercase tracking-[0.08em] text-muted-foreground">
        {t('proxies.pageSize')}
      </div>
      <Select
        compact
        value={String(pageSize)}
        onValueChange={(value) => setPageSize(Number(value))}
        options={PAGE_SIZE_OPTIONS.map((size) => ({ value: String(size), label: String(size) }))}
      />
    </div>
  </div>
  <div className="text-xs text-muted-foreground">
    {t('proxies.filteredCount', { count: filteredProxies.length })}
  </div>
</div>
```

- [ ] **Step 5: Make the empty state respect active filters**

When the full list is non-empty but the filtered list is empty, show a filter-specific empty state instead of pretending there are no proxies at all.

```tsx
) : filteredProxies.length === 0 ? (
  <div className="text-center py-16 text-muted-foreground">
    <Globe className="size-12 mx-auto mb-3 opacity-30" />
    <p className="text-sm font-medium">{t('proxies.noFilteredProxies')}</p>
    <p className="text-xs mt-1">{t('proxies.noFilteredProxiesDesc')}</p>
  </div>
) : (
```

- [ ] **Step 6: Run static verification after the filter/page-size changes**

Run:

```powershell
npm run typecheck
```

Expected: TypeScript completes without errors from the new state, `Select` props, or translation keys.

### Task 2: Harden Batch Delete Flow And Selection Cleanup

**Files:**
- Modify: `F:/自动化/codexproxy/codex2api/frontend/src/pages/Proxies.tsx`
- Modify: `F:/自动化/codexproxy/codex2api/frontend/src/locales/zh.json`
- Modify: `F:/自动化/codexproxy/codex2api/frontend/src/locales/en.json`

- [ ] **Step 1: Add locale entries for proxy batch-delete confirmation and loading state**

Add dedicated proxy strings instead of reusing account-delete copy.

```json
{
  "batchDeleteTitle": "批量删除代理",
  "batchDeleteDesc": "{{count}} 个选中的代理将从代理池中移除。请确认当前选择无误。",
  "batchDeleteConfirm": "确认删除",
  "deletingSelected": "删除中 ({{count}})"
}
```

```json
{
  "batchDeleteTitle": "Delete Selected Proxies",
  "batchDeleteDesc": "{{count}} selected proxies will be removed from the proxy pool. Please confirm the selection is correct.",
  "batchDeleteConfirm": "Confirm Delete",
  "deletingSelected": "Deleting ({{count}})"
}
```

- [ ] **Step 2: Reuse the shared confirm dialog hook in `Proxies.tsx`**

Import and initialize the existing hook instead of using `window.confirm`.

```tsx
import { useConfirmDialog } from '../hooks/useConfirmDialog'

const { confirm, confirmDialog } = useConfirmDialog()
const [batchDeleting, setBatchDeleting] = useState(false)
```

- [ ] **Step 3: Replace the current batch-delete handler with confirmed, in-flight-safe logic**

Keep cross-page selection semantics, but block duplicate submits and retain selection on failure.

```tsx
const handleBatchDelete = async () => {
  if (selected.size === 0 || batchDeleting) return

  const confirmed = await confirm({
    title: t('proxies.batchDeleteTitle'),
    description: t('proxies.batchDeleteDesc', { count: selected.size }),
    tone: 'destructive',
    confirmVariant: 'destructive',
    confirmText: t('proxies.batchDeleteConfirm'),
  })
  if (!confirmed) return

  setBatchDeleting(true)
  try {
    await api.batchDeleteProxies([...selected])
    setSelected(new Set())
    await reload()
  } catch {
    // keep selection so the user can retry
  } finally {
    setBatchDeleting(false)
  }
}
```

- [ ] **Step 4: Prune stale selections whenever the proxy list changes**

This keeps the selection model stable after reloads or deletions without breaking cross-page selection.

```tsx
useEffect(() => {
  const validIds = new Set(proxies.map((proxy) => proxy.id))
  setSelected((prev) => {
    const next = new Set([...prev].filter((id) => validIds.has(id)))
    return next.size === prev.size ? prev : next
  })
}, [proxies])
```

- [ ] **Step 5: Update the batch-delete button and page markup to reflect the improved flow**

Disable the destructive action while a request is in-flight and mount the shared dialog node once near the page root.

```tsx
{selected.size > 0 && (
  <button
    onClick={handleBatchDelete}
    disabled={batchDeleting}
    className="flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-semibold bg-destructive/10 border border-destructive/30 text-destructive hover:bg-destructive/20 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
  >
    {batchDeleting ? <Loader2 className="size-4 animate-spin" /> : <Trash2 className="size-4" />}
    {batchDeleting
      ? t('proxies.deletingSelected', { count: selected.size })
      : t('proxies.deleteSelected', { count: selected.size })}
  </button>
)}

{confirmDialog}
```

- [ ] **Step 6: Run build verification after the destructive-flow changes**

Run:

```powershell
npm run build
```

Expected: Vite completes a production build successfully with the updated proxy page and translation files.

### Task 3: Manual Regression And Closeout

**Files:**
- Verify: `F:/自动化/codexproxy/codex2api/frontend/src/pages/Proxies.tsx`
- Verify: `F:/自动化/codexproxy/codex2api/frontend/src/locales/zh.json`
- Verify: `F:/自动化/codexproxy/codex2api/frontend/src/locales/en.json`

- [ ] **Step 1: Start the frontend locally against the existing backend**

Run:

```powershell
npm run dev
```

Expected: Vite dev server starts and the admin UI remains reachable under the existing proxy configuration.

- [ ] **Step 2: Verify the new page-size control**

Check manually:

```text
1. Open /admin/proxies
2. Switch per-page between 10, 50, 100, 200
3. Confirm the visible row count changes and page resets to page 1
4. Confirm pagination totals follow the filtered dataset, not the raw unfiltered list
```

- [ ] **Step 3: Verify the region filter**

Check manually:

```text
1. Choose a concrete tested region from the dropdown
2. Confirm only proxies with matching test_location remain visible
3. Choose "Untested" and confirm only rows with empty test_location remain
4. Return to "All regions" and confirm the full list returns
```

- [ ] **Step 4: Verify batch delete and selection behavior**

Check manually:

```text
1. Select proxies across more than one page
2. Click the batch-delete button and confirm the destructive dialog appears
3. Cancel once to confirm nothing changes
4. Confirm delete and verify the button enters a disabled loading state
5. After reload, confirm selected IDs are cleared and page number remains valid
```

- [ ] **Step 5: Run existing proxy-page regressions**

Check manually:

```text
1. Toggle proxy enabled/disabled
2. Run single-proxy test
3. Run test-all
4. Delete one proxy with the single delete action
5. Toggle proxy pool on/off
```

- [ ] **Step 6: Commit**

Run:

```powershell
git add frontend/src/pages/Proxies.tsx frontend/src/locales/zh.json frontend/src/locales/en.json docs/superpowers/specs/2026-04-01-proxy-management-design.md docs/superpowers/plans/2026-04-01-proxy-management-enhancements.md
git commit -m "feat(proxy): 增强代理管理交互"
```

Expected: A single feature commit containing the proxy page enhancements and their design/plan documents.
