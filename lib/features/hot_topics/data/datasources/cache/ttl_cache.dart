class CacheEntry<V> {
  CacheEntry({
    required this.value,
    required this.storedAt,
    required this.ttl,
  });

  final V value;
  final DateTime storedAt;
  final Duration ttl;

  bool isExpired(DateTime now) => now.isAfter(storedAt.add(ttl));
}

class TtlCache<K, V> {
  TtlCache({
    required Duration ttl,
    DateTime Function()? now,
  })  : _ttl = ttl,
        _now = now ?? DateTime.now;

  final Map<K, CacheEntry<V>> _store = {};
  final Duration _ttl;
  final DateTime Function() _now;

  V? read(K key) {
    final entry = _store[key];
    if (entry == null) return null;

    final now = _now();
    if (entry.isExpired(now)) {
      _store.remove(key);
      return null;
    }

    return entry.value;
  }

  void write(K key, V value) {
    _store[key] = CacheEntry(value: value, storedAt: _now(), ttl: _ttl);
  }

  void invalidate(K key) => _store.remove(key);

  void clear() => _store.clear();

  void clearExpired() {
    final now = _now();
    final expiredKeys = <K>[];
    _store.forEach((key, entry) {
      if (entry.isExpired(now)) expiredKeys.add(key);
    });
    for (final key in expiredKeys) {
      _store.remove(key);
    }
  }

  int get size => _store.length;
}
