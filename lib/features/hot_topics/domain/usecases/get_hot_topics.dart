import '../entities/hot_topic.dart';
import '../entities/topic_source.dart';
import '../repositories/hot_topics_repository.dart';

class GetHotTopicsUseCase {
  const GetHotTopicsUseCase(this._repository);

  final HotTopicsRepository _repository;

  Future<List<HotTopic>> call({
    TopicSource? source,
    bool forceRefresh = false,
  }) {
    return _repository.getHotTopics(
      source: source,
      forceRefresh: forceRefresh,
    );
  }
}
