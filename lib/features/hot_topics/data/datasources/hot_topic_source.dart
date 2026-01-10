import '../../domain/entities/topic_source.dart';
import '../models/hot_topic_model.dart';

/// Data source adapter for a single hot-topic platform (e.g. Weibo/Zhihu).
abstract class HotTopicSource {
  TopicSource get source;

  Future<List<HotTopicModel>> fetchHotTopics();

  /// Best-effort search.
  ///
  /// Implementations may choose to:
  /// - call an official/unofficial search endpoint
  /// - or fallback to filtering the current hot list
  Future<List<HotTopicModel>> search(String query);
}
