import '../../domain/entities/hot_topic.dart';
import '../../domain/entities/topic_source.dart';
import '../../domain/repositories/hot_topics_repository.dart';
import '../datasources/cache/hot_topics_cache.dart';
import '../datasources/hot_topic_source.dart';
import '../models/hot_topic_model.dart';

class HotTopicsRepositoryImpl implements HotTopicsRepository {
  HotTopicsRepositoryImpl({
    required List<HotTopicSource> sources,
    required HotTopicsCache cache,
  })  : _sources = List<HotTopicSource>.unmodifiable(sources),
        _cache = cache;

  final List<HotTopicSource> _sources;
  final HotTopicsCache _cache;

  @override
  Future<List<HotTopic>> getHotTopics({
    TopicSource? source,
    bool forceRefresh = false,
  }) async {
    if (!forceRefresh) {
      final cached = _cache.readHotTopics(source: source);
      if (cached != null) return cached;
    }

    if (source != null) {
      final adapter = _findSource(source);
      final topics = await adapter.fetchHotTopics();
      final sorted = _sortedForSingleSource(topics);
      _cache.writeHotTopics(source: source, topics: sorted);
      return sorted;
    }

    final attempts = await _fetchAllAttempts(
      (src) => src.fetchHotTopics(),
    );

    final merged = _mergeAndSortAcrossSources(
      attempts.map((a) => a.topics).toList(growable: false),
    );

    if (merged.isEmpty) {
      final firstError = attempts.cast<_FetchAttempt?>().firstWhere(
            (a) => a?.error != null,
            orElse: () => null,
          );
      if (firstError != null) {
        Error.throwWithStackTrace(firstError.error!, firstError.stackTrace!);
      }
    }

    _cache.writeHotTopics(source: null, topics: merged);
    return merged;
  }

  @override
  Future<List<HotTopic>> refreshHotTopics({TopicSource? source}) {
    _cache.invalidateHotTopics(source: source);
    return getHotTopics(source: source, forceRefresh: true);
  }

  @override
  Future<List<HotTopic>> searchHotTopics(
    String query, {
    TopicSource? source,
    bool forceRefresh = false,
  }) async {
    final normalized = query.trim();
    if (normalized.isEmpty) {
      return getHotTopics(source: source, forceRefresh: forceRefresh);
    }

    if (!forceRefresh) {
      final cached = _cache.readSearch(source: source, query: normalized);
      if (cached != null) return cached;
    }

    if (source != null) {
      final adapter = _findSource(source);
      final topics = await adapter.search(normalized);
      final sorted = _sortedForSingleSource(topics);
      _cache.writeSearch(source: source, query: normalized, topics: sorted);
      return sorted;
    }

    final attempts = await _fetchAllAttempts(
      (src) => src.search(normalized),
    );

    final merged = _mergeAndSortAcrossSources(
      attempts.map((a) => a.topics).toList(growable: false),
    );

    if (merged.isEmpty) {
      final firstError = attempts.cast<_FetchAttempt?>().firstWhere(
            (a) => a?.error != null,
            orElse: () => null,
          );
      if (firstError != null) {
        Error.throwWithStackTrace(firstError.error!, firstError.stackTrace!);
      }
    }

    _cache.writeSearch(source: null, query: normalized, topics: merged);
    return merged;
  }

  HotTopicSource _findSource(TopicSource source) {
    return _sources.firstWhere(
      (s) => s.source == source,
      orElse: () => throw StateError('No HotTopicSource registered for ${source.key}'),
    );
  }

  List<HotTopicModel> _sortedForSingleSource(List<HotTopicModel> topics) {
    final copy = List<HotTopicModel>.of(topics);
    copy.sort((a, b) => a.rank.compareTo(b.rank));
    return copy;
  }

  Future<List<_FetchAttempt>> _fetchAllAttempts(
    Future<List<HotTopicModel>> Function(HotTopicSource source) loader,
  ) async {
    final tasks = _sources.map((source) async {
      try {
        final topics = await loader(source);
        return _FetchAttempt(
          source: source,
          topics: topics,
          error: null,
          stackTrace: null,
        );
      } catch (e, st) {
        return _FetchAttempt(
          source: source,
          topics: const <HotTopicModel>[],
          error: e,
          stackTrace: st,
        );
      }
    });

    return Future.wait(tasks);
  }

  List<HotTopicModel> _mergeAndSortAcrossSources(List<List<HotTopicModel>> fetched) {
    final merged = <HotTopicModel>[];
    for (final list in fetched) {
      merged.addAll(list);
    }

    final sourceOrder = {
      for (var i = 0; i < _sources.length; i++) _sources[i].source: i,
    };

    merged.sort((a, b) {
      final ao = sourceOrder[a.source] ?? 999;
      final bo = sourceOrder[b.source] ?? 999;
      final bySource = ao.compareTo(bo);
      if (bySource != 0) return bySource;
      return a.rank.compareTo(b.rank);
    });

    return merged;
  }
}

class _FetchAttempt {
  const _FetchAttempt({
    required this.source,
    required this.topics,
    required this.error,
    required this.stackTrace,
  });

  final HotTopicSource source;
  final List<HotTopicModel> topics;
  final Object? error;
  final StackTrace? stackTrace;
}
