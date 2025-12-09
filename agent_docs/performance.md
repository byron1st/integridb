# Performance Expectations

| Operation | MVP | Phase 2 (with snapshots) |
|-----------|-----|--------------------------|
| Insert | ~1ms | ~1-2ms |
| Update | ~2ms | ~2-3ms |
| Query (by ID) | ~1ms | ~1ms |
| Full Replay (1000 events) | ~100ms | ~10ms |
| Verify Single Event | O(n) ~100ms | O(log n) ~1ms |
| Verify Full Chain | ~50ms | ~50ms |

# Storage Estimates

| Events | Database Size | With Snapshots + Compaction |
|--------|--------------|----------------------------|
| 10,000 | ~10 MB | ~3 MB |
| 100,000 | ~100 MB | ~20 MB |
| 1,000,000 | ~1 GB | ~150 MB |
