import '../entities/hot_topic.dart';
import '../entities/topic_source.dart';

abstract class HotTopicsRepository {
  /// Get hot topics.
  ///
  /// If [source] is null, fetches/returns a merged list from all sources.
  Future<List<HotTopic>> getHotTopics({
    TopicSource? source,
    bool forceRefresh = false,
  });

  /// Force refresh hot topics (bypassing cache).
  Future<List<HotTopic>> refreshHotTopics({TopicSource? source});

  /// Search hot topics.
  ///
  /// Search semantics are platform-specific and best-effort.
  Future<List<HotTopic>> searchHotTopics(
    String query, {
    TopicSource? source,
    bool forceRefresh = false,
  });
}
