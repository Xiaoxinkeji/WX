import '../../../domain/entities/topic_source.dart';
import '../../models/hot_topic_model.dart';
import 'ttl_cache.dart';

class HotTopicsCache {
  HotTopicsCache({
    required Duration ttl,
    DateTime Function()? now,
  })  : _hotListCache = TtlCache<String, List<HotTopicModel>>(ttl: ttl, now: now),
        _searchCache = TtlCache<String, List<HotTopicModel>>(ttl: ttl, now: now);

  final TtlCache<String, List<HotTopicModel>> _hotListCache;
  final TtlCache<String, List<HotTopicModel>> _searchCache;

  List<HotTopicModel>? readHotTopics({TopicSource? source}) {
    return _hotListCache.read(_hotKey(source));
  }

  void writeHotTopics({TopicSource? source, required List<HotTopicModel> topics}) {
    _hotListCache.write(_hotKey(source), topics);
  }

  List<HotTopicModel>? readSearch({TopicSource? source, required String query}) {
    return _searchCache.read(_searchKey(source, query));
  }

  void writeSearch({
    TopicSource? source,
    required String query,
    required List<HotTopicModel> topics,
  }) {
    _searchCache.write(_searchKey(source, query), topics);
  }

  void invalidateHotTopics({TopicSource? source}) {
    _hotListCache.invalidate(_hotKey(source));
  }

  void invalidateSearch({TopicSource? source, required String query}) {
    _searchCache.invalidate(_searchKey(source, query));
  }

  void clearAll() {
    _hotListCache.clear();
    _searchCache.clear();
  }

  String _hotKey(TopicSource? source) => 'hot:${source?.key ?? 'all'}';

  String _searchKey(TopicSource? source, String query) {
    final normalized = query.trim().toLowerCase();
    return 'search:${source?.key ?? 'all'}:$normalized';
  }
}
