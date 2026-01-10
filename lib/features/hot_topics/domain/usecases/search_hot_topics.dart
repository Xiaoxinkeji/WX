import '../entities/hot_topic.dart';
import '../entities/topic_source.dart';
import '../repositories/hot_topics_repository.dart';

class SearchHotTopicsUseCase {
  const SearchHotTopicsUseCase(this._repository);

  final HotTopicsRepository _repository;

  Future<List<HotTopic>> call(
    String query, {
    TopicSource? source,
    bool forceRefresh = false,
  }) {
    return _repository.searchHotTopics(
      query,
      source: source,
      forceRefresh: forceRefresh,
    );
  }
}
